package identity

import (
	"errors"
	"testing"
)

// TestCheckSiteOwnershipMatch 键相符返回 nil（归属命中，插件继续执行拉取）
func TestCheckSiteOwnershipMatch(t *testing.T) {
	if err := CheckSiteOwnership(Bilibili.Key, Bilibili.Key); err != nil {
		t.Fatalf("键相符应返回 nil: %v", err)
	}
}

// TestCheckSiteOwnershipMismatch 键不符返回可被 errors.Is 识别的 ErrSiteNotOwned
// （插件实现经本辅助返回未归属错误，SDK 服务端适配层按哨兵转译跨进程状态码）
func TestCheckSiteOwnershipMismatch(t *testing.T) {
	err := CheckSiteOwnership(Bilibili.Key, Pixiv.Key)
	if err == nil {
		t.Fatal("键不符应返回错误")
	}
	if !errors.Is(err, ErrSiteNotOwned) {
		t.Fatalf("错误应可被 errors.Is 识别为 ErrSiteNotOwned: %v", err)
	}
	if errors.Is(CheckSiteOwnership(Bilibili.Key, Bilibili.Key), ErrSiteNotOwned) {
		t.Fatal("键相符的错误判定不应误报 ErrSiteNotOwned")
	}
}
