package dto

// CreateTaskResult 主程序 CreateTask 的返回结果
type CreateTaskResult struct {
	Succeed       bool
	AddedQuantity int
	Msg           string
}

// CreateIntentKind Create 流意图信号的两态：插件经 Create 流末尾的意图块声明
// 「本流不产出任务」的成功收尾形态。零值 None 兼作「未声明意图」的判定哨兵
//（消费侧以 kind != CreateIntentNone 判定意图在场）。
type CreateIntentKind int32

const (
	// CreateIntentNone 未声明意图（零值）
	CreateIntentNone CreateIntentKind = 0
	// CreateIntentToParse 转入解析：本流零任务，控制权移交插件的解析交互面
	CreateIntentToParse CreateIntentKind = 1
	// CreateIntentCancelled 已取消：用户取消本次创建，本流零任务
	CreateIntentCancelled CreateIntentKind = 2
)

// TaskCreateResult 插件 Create 方法的返回结果，支持批量或流式。
// reason 由插件经 SetReason 声明、经 Create 流末尾的 error 块透传给主程序，
// intent 由插件经 SetIntent 声明、经 Create 流末尾的 intent 块透传给主程序，
// 与 array/stream 一样是同一结构体在写入方（插件/代理）与读取方（服务端/主程序）间复用的字段。
type TaskCreateResult struct {
	array         []*TaskCreateResponse
	stream        <-chan *TaskCreateResponse
	isStream      bool
	reason        string
	intentKind    CreateIntentKind
	intentMessage string
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

// SetReason 声明本次创建的业务原因（用户可读文本，如「未发现可导入的文件」），
// 用于零任务或部分成功后向用户解释。流式模式必须在关闭 stream channel 之前调用：
// 消费侧对 reason 的读取依赖 channel close 建立的 happens-before 顺序。
func (r *TaskCreateResult) SetReason(reason string) {
	r.reason = reason
}

// Reason 读取 SetReason 声明的原因，未声明确返回空串。
// 流式模式必须在流消费完毕（channel close、range 循环退出）之后调用，
// 此前读取与生产侧 SetReason 之间不存在 happens-before 保证。
func (r *TaskCreateResult) Reason() string {
	return r.reason
}

// SetIntent 声明本次创建以意图信号收尾：零任务的成功收尾（转入解析/已取消 + 用户可读文案，
// 如「已转入解析工作台」「已取消本次导入」），主程序消费为成功零任务 + Msg=文案。
// 与 SetReason 互斥——reason 是失败收尾的解释、intent 是成功收尾的意图，同设时意图生效
//（发送侧只发意图块）；意图收尾本流不应携带任务声明，携带时主程序按任务数收口、意图被忽略。
// 流式模式必须在关闭 stream channel 之前调用：消费侧对意图的读取依赖
// channel close 建立的 happens-before 顺序。
func (r *TaskCreateResult) SetIntent(kind CreateIntentKind, message string) {
	r.intentKind = kind
	r.intentMessage = message
}

// Intent 读取 SetIntent 声明的意图信号，未声明确返回 (CreateIntentNone, "")。
// 流式模式必须在流消费完毕（channel close、range 循环退出）之后调用，
// 此前读取与生产侧 SetIntent 之间不存在 happens-before 保证。
func (r *TaskCreateResult) Intent() (CreateIntentKind, string) {
	return r.intentKind, r.intentMessage
}
