package dto

import "context"

// TaskHandler 任务处理器接口
// 插件实现此接口处理任务。Start/Resume 返回 StoreSpec 流集合(含下载型 downloaded 与派生型 derived)。
type TaskHandler interface {
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
