package identity

import "testing"

// TestIdentityRegistryUniqueKeys 注册表内各站点键两两互异——
// 同键条目会令 map 构建时静默合并，站点身份产生歧义
func TestIdentityRegistryUniqueKeys(t *testing.T) {
	all := []Site{Pixiv, Bilibili, Local}
	seen := make(map[string]string, len(all)) // 键 → 首个占用该键的站点名
	for _, s := range all {
		if s.Key == "" {
			t.Errorf("站点 %s 的键为空", s.Name)
			continue
		}
		if other, dup := seen[s.Key]; dup {
			t.Errorf("站点键重复：%s 同时用于 %s 与 %s", s.Key, other, s.Name)
			continue
		}
		seen[s.Key] = s.Name
	}
	if len(seen) != len(registry) {
		t.Errorf("注册表条目数(%d)与常量数(%d)不一致——存在同键条目被 map 合并", len(registry), len(all))
	}
}

// TestIdentityLookup 已注册键的 Lookup 结果与常量一致；未注册键与空键不命中
func TestIdentityLookup(t *testing.T) {
	for _, s := range []Site{Pixiv, Bilibili, Local} {
		got, ok := Lookup(s.Key)
		if !ok {
			t.Fatalf("已注册键 %s 查询未命中", s.Key)
		}
		if got != s {
			t.Errorf("Lookup(%q) 返回条目与常量不一致：got %+v, want %+v", s.Key, got, s)
		}
	}
	if _, ok := Lookup("s0a0b0c0d0e0f"); ok {
		t.Error("未注册键不应命中")
	}
	if _, ok := Lookup(""); ok {
		t.Error("空键不应命中")
	}
}

// TestIdentityAll 全量枚举与 Lookup 逐键一致且互相覆盖、多次调用序稳定、
// 返回切片为副本（调用方修改不泄漏进注册表）
func TestIdentityAll(t *testing.T) {
	first := All()
	if len(first) != len(registry) {
		t.Fatalf("全量枚举条目数(%d)与注册表(%d)不一致", len(first), len(registry))
	}
	seen := make(map[string]Site, len(first))
	for _, s := range first {
		if _, dup := seen[s.Key]; dup {
			t.Errorf("全量枚举出现重复键：%s", s.Key)
			continue
		}
		seen[s.Key] = s
		got, ok := Lookup(s.Key)
		if !ok {
			t.Errorf("All 中的键 %s 经 Lookup 未命中", s.Key)
			continue
		}
		if got != s {
			t.Errorf("键 %s 的 All 条目与 Lookup 不一致：All %+v, Lookup %+v", s.Key, s, got)
		}
	}
	// 覆盖性：注册表索引中的每个键都出现在 All 结果中
	for key := range registryIndex {
		if _, ok := seen[key]; !ok {
			t.Errorf("注册表键 %s 未出现在 All 结果中", key)
		}
	}
	// 序稳定：多次调用输出逐位一致
	second := All()
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("两次 All 第 %d 位不一致：%+v vs %+v", i, first[i], second[i])
		}
	}
	// 副本语义：修改返回切片不影响注册表
	first[0].Name = "已篡改"
	if All()[0].Name == "已篡改" {
		t.Error("All 返回切片未做副本——调用方修改泄漏进注册表")
	}
}
