package dto

import "github.com/lvfeng-z/library-squirrel-sdk/gen"

// SettingChangedRequest 设置变更通知请求（别名 gen.SettingChangedRequest，proto 单源）。
// 宿主在插件设置落库（保存/重置）成功后异步推送：Source 为变更来源（save=保存更新 /
// reset=重置回默认，操作语义由来源决定），Keys 为本次变更的设置键集合（单次保存/重置
// 为一个键，留批量余量）。
type SettingChangedRequest = gen.SettingChangedRequest

// 设置变更来源（SettingChangedRequest.Source 取值，封闭词汇）
const (
	// SettingChangeSourceSave 保存更新（宿主设置服务保存落库）
	SettingChangeSourceSave = "save"
	// SettingChangeSourceReset 重置回默认（宿主设置服务重置落库）
	SettingChangeSourceReset = "reset"
)

// SettingChangeHandler 设置变更处置函数：宿主推送设置变更通知时分派到此，插件自行
// 反应（如重载配置热生效——激活时快照的设置改后不重启进程即生效）。ctx 为 Activate
// 时构造的插件上下文，req 携带来源与键；无返回值——错误由插件自行记日志。处置函数
// 在 RPC goroutine 执行，须快速返回且并发安全自理（宿主侧通知带超时预算，慢处置会被
// RPC 取消）。
type SettingChangeHandler func(ctx PluginContext, req *SettingChangedRequest)
