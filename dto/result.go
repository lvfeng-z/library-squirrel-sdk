package dto

// CreateTaskResult 主程序 CreateTask 的返回结果
type CreateTaskResult struct {
	Succeed       bool
	AddedQuantity int
	Msg           string
}

// TaskCreateResult 插件 Create 方法的返回结果，支持批量或流式
type TaskCreateResult struct {
	array    []*TaskCreateResponse
	stream   <-chan *TaskCreateResponse
	isStream bool
}

// BatchResult 创建批量模式的结果
func BatchResult(responses []*TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{array: responses}
}

// StreamResult 创建流式模式的结果
func StreamResult(ch <-chan *TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{stream: ch, isStream: true}
}

// IsStream 是否为流式模式
func (r *TaskCreateResult) IsStream() bool {
	return r.isStream
}

// Array 获取批量结果（仅批量模式有效）
func (r *TaskCreateResult) Array() []*TaskCreateResponse {
	return r.array
}

// Stream 获取流式 channel（仅流式模式有效）
func (r *TaskCreateResult) Stream() <-chan *TaskCreateResponse {
	return r.stream
}
