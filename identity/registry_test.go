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
