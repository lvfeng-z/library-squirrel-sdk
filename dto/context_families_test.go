package dto

import (
	"reflect"
	"testing"
)

// contextFamilies 能力族登记表：新增族（或既有族的增删）时同步维护此表，
// TestContextFamiliesPartition 据此锚定划分不变量
var contextFamilies = []reflect.Type{
	reflect.TypeOf((*KVStore)(nil)).Elem(),
	reflect.TypeOf((*Preference)(nil)).Elem(),
	reflect.TypeOf((*Proxy)(nil)).Elem(),
	reflect.TypeOf((*TaskTrigger)(nil)).Elem(),
	reflect.TypeOf((*FrontendEvents)(nil)).Elem(),
	reflect.TypeOf((*Environment)(nil)).Elem(),
	reflect.TypeOf((*LibraryQuery)(nil)).Elem(),
	reflect.TypeOf((*LogSink)(nil)).Elem(),
}

// TestContextFamiliesPartition 族划分结构不变量：①每族非空；②族间两两不相交
// （一个方法只归一族）；③复合接口方法集恰为全族并集——新方法必须「建族或
// 入族」，不得向 PluginContext 直接添加散方法，也不得让族游离于复合之外
func TestContextFamiliesPartition(t *testing.T) {
	composite := reflect.TypeOf((*PluginContext)(nil)).Elem()
	union := make(map[string]string) // 方法名 → 所属族名
	for _, f := range contextFamilies {
		if f.NumMethod() == 0 {
			t.Errorf("族 %s 为空（空族没有存在意义）", f.Name())
		}
		for i := 0; i < f.NumMethod(); i++ {
			name := f.Method(i).Name
			if prev, dup := union[name]; dup {
				t.Errorf("方法 %s 同时归属族 %s 与 %s（划分须两两不相交）", name, prev, f.Name())
			}
			union[name] = f.Name()
		}
	}
	for i := 0; i < composite.NumMethod(); i++ {
		name := composite.Method(i).Name
		if _, ok := union[name]; !ok {
			t.Errorf("PluginContext 方法 %s 不属于任何族（新方法必须归族，不得散挂复合接口）", name)
		}
	}
	if got, want := composite.NumMethod(), len(union); got != want {
		t.Errorf("PluginContext 方法数 %d ≠ 族并集方法数 %d（存在族未被复合接口嵌入，或登记表缺族）", got, want)
	}
}

// preferenceKeys 窄消费者示例：参数收窄为族类型而非 PluginContext——本函数在
// 编译期即无法触碰 KV/任务/库查询等其他能力族（能力面按族显形的消费侧收益）
func preferenceKeys(c Preference) ([]string, error) {
	return c.ListMyPreferences()
}

// TestContextNarrowConsumer 窄消费形态编译期成立（无行为断言，作族用法的活文档）
func TestContextNarrowConsumer(t *testing.T) {
	var _ func(Preference) ([]string, error) = preferenceKeys
}
