package dto

import "github.com/lvfeng-z/library-squirrel-sdk/gen"

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

// TaskChildrenBatch 子行批次块（别名 gen.TaskChildrenBatch，proto 单源；定义随流产物
// 条目类型落位本文件）。契约 21：大批次创建时父块 children 置空、子行拆多块发射，
// 四条不变量见 proto 消息注释（仅流式 mode 合法、按 plugin_task_id 指称聚合、
// 内联继续有效且混用都计入、未收批统一回滚——流异常中断/取消提前收流/error
// 收流一律由宿主删父任务与已落库子行）。
type TaskChildrenBatch = gen.TaskChildrenBatch

// CreateStreamItem Create 流式发射条目（契约 21）：task 块（TaskCreateResponse）或
// 子行批次块（TaskChildrenBatch）两种形态，恰居其一——由构造器保证，消费侧
//（transport.sendCreateResult）按形态发对应 CreateChunk。大批次分帧的发射载体：
// 父块经 TaskItem 发出（children 置空），子行按批经 ChildrenBatchItem 发出、
// 末批 last=true 必为该父批次序列的最后一条。
type CreateStreamItem struct {
	task          *TaskCreateResponse
	childrenBatch *TaskChildrenBatch
}

// TaskItem 构造 task 块形态的发射条目。
func TaskItem(task *TaskCreateResponse) *CreateStreamItem {
	return &CreateStreamItem{task: task}
}

// ChildrenBatchItem 构造子行批次块形态的发射条目：pluginTaskId 指称归属父任务
//（=TaskCreateResponse.PluginTaskId，不依赖块位置），last=true 声明该父的末批。
func ChildrenBatchItem(pluginTaskId string, items []*TaskCreateChildResponse, last bool) *CreateStreamItem {
	return &CreateStreamItem{childrenBatch: &TaskChildrenBatch{
		PluginTaskId: pluginTaskId,
		Items:        items,
		Last:         last,
	}}
}

// Task task 块形态载荷；非该形态返回 nil。
func (i *CreateStreamItem) Task() *TaskCreateResponse {
	return i.task
}

// ChildrenBatch 子行批次块形态载荷；非该形态返回 nil。
func (i *CreateStreamItem) ChildrenBatch() *TaskChildrenBatch {
	return i.childrenBatch
}

// TaskCreateResult 插件 Create 方法的返回结果，支持批量或流式。
// reason 由插件经 SetReason 声明、经 Create 流末尾的 error 块透传给主程序，
// intent 由插件经 SetIntent 声明、经 Create 流末尾的 intent 块透传给主程序，
// 与 array/stream 一样是同一结构体在写入方（插件/代理）与读取方（服务端/主程序）间复用的字段。
// 流式双形态互斥：条目形态（StreamItemsResult 构造，可携 task 块/子行批次块）与
// 旧应答形态（StreamResult 构造，逐条发 task 块）——前者 Items() 非 nil、Stream() 为 nil，
// 后者反之。
type TaskCreateResult struct {
	array         []*TaskCreateResponse
	stream        <-chan *TaskCreateResponse
	itemStream    <-chan *CreateStreamItem
	isStream      bool
	reason        string
	intentKind    CreateIntentKind
	intentMessage string
}

// BatchResult 创建批量模式的结果
func BatchResult(responses []*TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{array: responses}
}

// StreamResult 创建流式模式的结果（旧应答形态）：channel 每条应答发射为 task 块，
// 行为与历史版本一致——既有调用方（pixiv/bilibili 等）零改动。
func StreamResult(ch <-chan *TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{stream: ch, isStream: true}
}

// StreamItemsResult 创建流式模式的结果（条目形态，契约 21）：channel 条目可携
// task 块/子行批次块两种形态，供大批次创建分帧发射；子行批次块仅流式 mode 合法，
// 本构造恒为流式。
func StreamItemsResult(ch <-chan *CreateStreamItem) *TaskCreateResult {
	return &TaskCreateResult{itemStream: ch, isStream: true}
}

// IsStream 是否为流式模式
func (r *TaskCreateResult) IsStream() bool {
	return r.isStream
}

// Array 获取批量结果（仅批量模式有效）
func (r *TaskCreateResult) Array() []*TaskCreateResponse {
	return r.array
}

// Stream 获取流式 channel（仅旧应答形态有效；条目形态返回 nil，经 Items 取）
func (r *TaskCreateResult) Stream() <-chan *TaskCreateResponse {
	return r.stream
}

// Items 获取流式条目 channel（仅条目形态有效；旧应答形态返回 nil，经 Stream 取）
func (r *TaskCreateResult) Items() <-chan *CreateStreamItem {
	return r.itemStream
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
