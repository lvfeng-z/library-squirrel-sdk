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
