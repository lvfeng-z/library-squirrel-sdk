package transport

import (
	"context"
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// preferenceProviderStub 偏好域内存桩：Set 写入内存、Get 按 ok 语义读、List 列键，
// 用于验证 PluginContextClient → HostService → HostDeps.PreferenceProvider 全链
type preferenceProviderStub struct {
	store map[string]*dto.PreferenceValue
	keys  []string // ListMyPreferences 返回值（固定桩数据）
}

func newPreferenceProviderStub() *preferenceProviderStub {
	return &preferenceProviderStub{store: make(map[string]*dto.PreferenceValue)}
}

func (f *preferenceProviderStub) GetPreference(_ context.Context, key string) (*dto.PreferenceValue, bool, error) {
	v, ok := f.store[key]
	return v, ok, nil
}

func (f *preferenceProviderStub) SetPreference(_ context.Context, key string, value *dto.PreferenceValue) error {
	f.store[key] = value
	return nil
}

func (f *preferenceProviderStub) ListMyPreferences(_ context.Context) ([]string, error) {
	return f.keys, nil
}

// TestPluginContextPreferenceReachesProvider 偏好域三方法全链：PluginContextClient →
// HostService → HostDeps 注入的 PreferenceProvider。Set 写入的信封四字段线级往返不损，
// Get 无记录返回 (nil,false,nil) 不报错，List 键清单直通
func TestPluginContextPreferenceReachesProvider(t *testing.T) {
	provider := newPreferenceProviderStub()
	provider.keys = []string{"illust.form", "video.form"}
	conn := serveGRPC(t, func(s *grpc.Server) {
		RegisterHostService(s, HostDeps{PreferenceProvider: provider})
	})
	ctx := NewPluginContextClient(conn)

	// Set：完整值信封经线级到达 provider
	if err := ctx.SetPreference("illust.form", &dto.PreferenceValue{
		SchemaVersion: 1,
		Title:         "图文形态",
		Description:   "多图",
		Data:          `{"form":"multi"}`,
	}); err != nil {
		t.Fatalf("SetPreference 失败: %v", err)
	}
	got := provider.store["illust.form"]
	if got == nil {
		t.Fatal("SetPreference 未到达 provider")
	}
	if got.SchemaVersion != 1 || got.Title != "图文形态" || got.Description != "多图" || got.Data != `{"form":"multi"}` {
		t.Fatalf("信封线级往返失真: %+v", got)
	}

	// Get 有记录：值直通 + ok=true
	v, ok, err := ctx.GetPreference("illust.form")
	if err != nil || !ok {
		t.Fatalf("GetPreference 有记录: ok=%t err=%v", ok, err)
	}
	if v.Title != "图文形态" || v.SchemaVersion != 1 || v.Data != `{"form":"multi"}` {
		t.Fatalf("GetPreference 返回值不符: %+v", v)
	}

	// Get 无记录：(nil, false, nil)——无记录是合法状态而非错误，调用方据此重新发起问答
	v, ok, err = ctx.GetPreference("never.set")
	if err != nil {
		t.Fatalf("GetPreference 无记录不应报错: %v", err)
	}
	if ok || v != nil {
		t.Fatalf("GetPreference 无记录应返回 (nil,false,nil), 实得 (%v,%t,nil)", v, ok)
	}

	// List：本插件键清单直通
	keys, err := ctx.ListMyPreferences()
	if err != nil {
		t.Fatalf("ListMyPreferences 失败: %v", err)
	}
	if len(keys) != 2 || keys[0] != "illust.form" || keys[1] != "video.form" {
		t.Fatalf("键清单不符: %v", keys)
	}
}

// TestPluginContextPreferenceWithoutProviderUnimplemented 宿主未注入
// PreferenceProvider（宿主未装配偏好服务）：偏好调用得 codes.Unimplemented——
// 与库查询能力未配置的降级形态一致
func TestPluginContextPreferenceWithoutProviderUnimplemented(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		RegisterHostService(s, HostDeps{})
	})
	ctx := NewPluginContextClient(conn)
	if _, _, err := ctx.GetPreference("k"); status.Code(err) != codes.Unimplemented {
		t.Fatalf("GetPreference 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
	if err := ctx.SetPreference("k", &dto.PreferenceValue{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("SetPreference 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
	if _, err := ctx.ListMyPreferences(); status.Code(err) != codes.Unimplemented {
		t.Fatalf("ListMyPreferences 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
}

// legacyStorageStub 旧宿主既有面桩：仅实现 GetValue，验证偏好三方法缺席时既有面不受影响
type legacyStorageStub struct {
	dto.StorageProvider
}

func (f *legacyStorageStub) GetValue(_ context.Context, key string) (*dto.StorageValue, error) {
	return &dto.StorageValue{Value: "legacy:" + key, SchemaVersion: 12}, nil
}

// legacyHostServiceDesc 裁掉偏好三方法的服务描述：模拟契约 12 旧宿主进程内的
// HostService 注册形态（旧 gen 代码无此三方法，gRPC 对未注册方法回 Unimplemented）
func legacyHostServiceDesc() *grpc.ServiceDesc {
	desc := gen.HostService_ServiceDesc // 值拷贝，改切片头不触原变量
	methods := make([]grpc.MethodDesc, 0, len(desc.Methods))
	for _, m := range desc.Methods {
		switch m.MethodName {
		case "GetPreference", "SetPreference", "ListMyPreferences":
			continue
		}
		methods = append(methods, m)
	}
	desc.Methods = methods
	return &desc
}

// TestPreferenceRPCAgainstLegacyHost 旧契约兼容：新客户端（本 SDK）对契约 12 旧宿主
//（HostService 已注册但无偏好三方法）——偏好调用得 codes.Unimplemented，既有 KV 面
//（GetValue）照常工作，新增方法组不破坏旧宿主上的既有能力
func TestPreferenceRPCAgainstLegacyHost(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		s.RegisterService(legacyHostServiceDesc(), NewHostServiceServer(HostDeps{
			StorageProvider: &legacyStorageStub{},
		}))
	})
	ctx := NewPluginContextClient(conn)

	if _, _, err := ctx.GetPreference("k"); status.Code(err) != codes.Unimplemented {
		t.Fatalf("旧宿主 GetPreference 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
	if err := ctx.SetPreference("k", &dto.PreferenceValue{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("旧宿主 SetPreference 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
	if _, err := ctx.ListMyPreferences(); status.Code(err) != codes.Unimplemented {
		t.Fatalf("旧宿主 ListMyPreferences 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}

	// 既有面不受新增影响：同连接上 GetValue 照常到达宿主
	v, err := ctx.GetValue("token")
	if err != nil {
		t.Fatalf("旧宿主既有面 GetValue 失败: %v", err)
	}
	if v.Value != "legacy:token" {
		t.Fatalf("既有面返回值不符: %+v", v)
	}
}
