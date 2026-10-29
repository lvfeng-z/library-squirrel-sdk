package dto

import "testing"

// TestTaskCreateResultReasonReadWrite reason 基础读写：
// 未声明返回空串，SetReason 声明后 Reason 读回同值（批量与流式两种构造同形）
func TestTaskCreateResultReasonReadWrite(t *testing.T) {
	batch := BatchResult(nil)
	if got := batch.Reason(); got != "" {
		t.Fatalf("批量结果未声明 reason 应为空串, 实得 %q", got)
	}
	batch.SetReason("未发现可导入的文件")
	if got := batch.Reason(); got != "未发现可导入的文件" {
		t.Fatalf("批量结果 reason = %q, 期望 %q", got, "未发现可导入的文件")
	}

	stream := StreamResult(nil)
	if got := stream.Reason(); got != "" {
		t.Fatalf("流式结果未声明 reason 应为空串, 实得 %q", got)
	}
	stream.SetReason("部分任务创建失败")
	if got := stream.Reason(); got != "部分任务创建失败" {
		t.Fatalf("流式结果 reason = %q, 期望 %q", got, "部分任务创建失败")
	}
}

// TestCreateStreamItemForms 发射条目两形态恰居其一；流式双形态互斥——
// 条目形态 Items() 携 channel 而 Stream() 为 nil，旧应答形态反之（零改动兼容）
func TestCreateStreamItemForms(t *testing.T) {
	task := TaskItem(&TaskCreateResponse{PluginTaskId: "p"})
	if task.Task() == nil || task.Task().PluginTaskId != "p" || task.ChildrenBatch() != nil {
		t.Fatalf("TaskItem 应仅携 task 形态, 实得 task=%v children=%v", task.Task(), task.ChildrenBatch())
	}
	batch := ChildrenBatchItem("p", []*TaskCreateChildResponse{{TaskName: "c"}}, true)
	cb := batch.ChildrenBatch()
	if batch.Task() != nil || cb == nil || cb.PluginTaskId != "p" || len(cb.Items) != 1 || !cb.Last {
		t.Fatalf("ChildrenBatchItem 应仅携子行批次形态(1 行, last=true), 实得 %+v", cb)
	}

	itemCh := make(chan *CreateStreamItem)
	items := StreamItemsResult(itemCh)
	if !items.IsStream() || items.Items() == nil || items.Stream() != nil {
		t.Fatal("条目形态结果应为流式且 Items() 非 nil、Stream() 为 nil")
	}
	legacy := StreamResult(make(chan *TaskCreateResponse))
	if !legacy.IsStream() || legacy.Stream() == nil || legacy.Items() != nil {
		t.Fatal("旧应答形态结果应为流式且 Stream() 非 nil、Items() 为 nil")
	}
}

// TestTaskCreateResultIntentReadWrite intent 基础读写：
// 未声明返回 (CreateIntentNone, "")，SetIntent 声明后 Intent 读回两态与文案
//（批量与流式两种构造同形；两态各验一次）
func TestTaskCreateResultIntentReadWrite(t *testing.T) {
	batch := BatchResult(nil)
	if kind, msg := batch.Intent(); kind != CreateIntentNone || msg != "" {
		t.Fatalf("批量结果未声明 intent 应为 (None, \"\"), 实得 (%d, %q)", kind, msg)
	}
	batch.SetIntent(CreateIntentToParse, "已转入解析工作台")
	if kind, msg := batch.Intent(); kind != CreateIntentToParse || msg != "已转入解析工作台" {
		t.Fatalf("批量结果 intent = (%d, %q), 期望 (%d, %q)", kind, msg, CreateIntentToParse, "已转入解析工作台")
	}

	stream := StreamResult(nil)
	stream.SetIntent(CreateIntentCancelled, "已取消本次导入")
	if kind, msg := stream.Intent(); kind != CreateIntentCancelled || msg != "已取消本次导入" {
		t.Fatalf("流式结果 intent = (%d, %q), 期望 (%d, %q)", kind, msg, CreateIntentCancelled, "已取消本次导入")
	}
}
