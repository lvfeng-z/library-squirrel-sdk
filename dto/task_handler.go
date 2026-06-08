package dto

import "io"

// TaskHandler 任务处理器接口
// 插件实现此接口来处理任务
type TaskHandler interface {
	// Create 创建任务
	// url: 需解析的url
	// 返回 TaskCreateResult（批量或流式）或错误
	Create(url string) (*TaskCreateResult, error)
	// CreateWorkInfo 生成作品信息
	// task: 需处理的任务
	// 返回作品信息或错误
	CreateWorkInfo(task *TaskDTO) (*WorkResponse, error)
	// Start 开始任务
	// task: 需开始的任务
	// 返回资源读取器（io.ReadCloser）、WorkResponse 或错误
	// 调用方负责关闭返回的 ReadCloser
	Start(task *TaskDTO) (io.ReadCloser, *WorkResponse, error)
	// Retry 重试任务
	// task: 需要重试的任务
	// 返回作品信息或错误
	Retry(task *TaskDTO) (*WorkResponse, error)
	// Pause 暂停任务
	// param: 暂停任务所需的参数
	Pause(param *TaskResParam) error
	// Stop 停止任务
	// param: 停止任务所需的参数
	Stop(param *TaskResParam) error
	// Resume 恢复任务
	// param: 恢复任务所需的参数
	// 返回作品信息或错误
	Resume(param *TaskResParam) (io.ReadCloser, *WorkResponse, error)
	// GetThumbnail 获取缩略图
	// taskData: 插件在 Create 阶段存储的任务数据（JSON）
	// 返回缩略图数据或 nil（插件决定不提供缩略图时返回 nil）
	GetThumbnail(taskData string) (*ThumbnailResponse, error)
}
