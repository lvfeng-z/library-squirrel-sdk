// Package plugin 提供插件侧配置迁移助手等便捷能力。
package plugin

import "github.com/lvfeng-z/library-squirrel-sdk/dto"

// MigrateStep 单步配置迁移函数（把数据从 v-1 迁移到 v）。
// 内部用 ctx.GetValue 读旧值、ctx.SetValue/ctx.DeleteValue 写新值；
// 须幂等——读 key 的 SchemaVersion，已迁至目标则 no-op（下次 Activate 重跑安全）。
type MigrateStep func(ctx dto.PluginContext) error

// MigrateConfig 配置 schema 版本迁移助手，插件在 Activate 开头（注册扩展点、读配置之前）调用。
//
// 按 1..target 顺序执行各 step（缺失版本跳过，确保 v1→v2→v3 不跳序）。
// best-effort：某步报错则记日志、跳过、继续，不阻断激活；幂等：已迁 key 的 SchemaVersion 已=target，
// 下次 Activate 重跑时 step 自查跳过，自动断点续传。此机制与任务模块（taskManager）的失败重跑无关。
//
// 插件用法：
//
//	func Activate(ctx pluginsdk.PluginContext) {
//	    plugin.MigrateConfig(ctx, 3, map[int]plugin.MigrateStep{
//	        2: migrateV1to2, // dlQuality -> downloadQuality
//	        3: migrateV2to3, // quality 字符串 -> 枚举
//	    })
//	    // 注册扩展点、读配置 ...
//	}
func MigrateConfig(ctx dto.PluginContext, target int, steps map[int]MigrateStep) {
	for v := 1; v <= target; v++ {
		step, ok := steps[v]
		if !ok {
			continue
		}
		if err := step(ctx); err != nil {
			ctx.Warnf("config migrate step %d failed: %v (skipped, best-effort; will retry next activate)", v, err)
		}
	}
}
