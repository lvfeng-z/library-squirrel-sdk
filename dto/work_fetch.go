package dto

import "context"

// WorkFetcher 作品拉取接口
// 插件实现此接口处理任务。Start/Resume 返回 StoreSpec 流集合(含下载型 downloaded 与派生型 derived)。
type WorkFetcher interface {
	// Create 创建任务
	Create(url string) (*TaskCreateResult, error)
	// CreateWorkInfo 生成作品信息
	CreateWorkInfo(task *TaskDTO) (*WorkResponse, error)
	// Start 开始任务,返回所选 storeRoles 的资源产出声明集合与作品信息
	// storeRoles 为本次执行所选 store_type 子集(空=全量),插件据此选择性产出,避免生成被丢弃的 store
	// ctx 为 gRPC stream context:主程序取消任务时 ctx 经 stream 传播到插件,SDK 的 serveSpecsPull
	// 据此 Close reader 中断在途读取;插件 reader 实现需保证 Close 可中断阻塞中的 Read
	Start(ctx context.Context, task *TaskDTO, storeRoles []string) ([]*StoreSpec, *WorkResponse, error)
	// Retry 重试任务
	Retry(task *TaskDTO) (*WorkResponse, error)
	// Pause 暂停任务(任务级,广播到全部 stream)
	Pause(param *TaskResParam) error
	// Stop 停止任务(任务级)
	Stop(param *TaskResParam) error
	// Resume 恢复任务:按 StreamOffsets 续传未完成 downloaded 轨、整轨重产 derived 轨
	// ctx 语义同 Start
	Resume(ctx context.Context, param *TaskResumeParam) ([]*StoreSpec, *WorkResponse, error)
}

// HeartbeatCreateFetcher Create 流等待期心跳的可选接入接口：在 WorkFetcher 之上
// 提供 CreateWithHeartbeat——SDK 服务端调用它时构造心跳上报器传入，插件在
// Create 的长等待点（等用户输入等）周期调用 reporter.Heartbeat() 维持宿主空闲
// 检测窗（越过 Create 首块空闲超时墙）。未实现本接口的插件照旧走 Create，
// 行为零变化。reporter 由 SDK 管理生命周期（handler 返回后关闭，结果块发送
// 独占流），实现方只调用不关闭
type HeartbeatCreateFetcher interface {
	WorkFetcher
	// CreateWithHeartbeat 带心跳上报器的 Create：主体逻辑同 Create，
	// 长等待点经 reporter 上报心跳
	CreateWithHeartbeat(url string, reporter HeartbeatReporter) (*TaskCreateResult, error)
}
