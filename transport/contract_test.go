package transport

import "testing"

// TestContractHistoryStructure 版本史结构不变量：条目自 1 连续编号、末条编号等于
// ContractVersion 常量、无空叙述——bump 时漏追加条目或错编号线在此 fail-fast
// （文档侧同步由主仓 backend/plugin/extension 的 TestDevGuideContractHistorySync 锚定）
func TestContractHistoryStructure(t *testing.T) {
	if len(ContractHistory) == 0 {
		t.Fatal("ContractHistory 为空")
	}
	for i, e := range ContractHistory {
		if want := i + 1; e.Number != want {
			t.Errorf("ContractHistory[%d].Number = %d, 期望 %d（条目须自 1 连续编号）", i, e.Number, want)
		}
		if e.Summary == "" {
			t.Errorf("ContractHistory[%d].Summary 为空（bump 时必须同步追加版本史叙述）", i)
		}
	}
	if last := ContractHistory[len(ContractHistory)-1].Number; last != ContractVersion {
		t.Errorf("版本史末条编号 = %d, 与 ContractVersion = %d 不一致（bump 须同步追加版本史条目）", last, ContractVersion)
	}
}
