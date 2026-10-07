# library-squirrel-sdk

[LibrarySquirrel](https://gitee.com/lv__feng/library-squirrel) 的插件开发 SDK：定义插件与宿主之间的 gRPC 协议（proto 契约与生成代码）、`PluginContext` 宿主能力访问绑定，以及插件渲染契约类型（`dto/render`）。

插件以独立进程运行，经本 SDK 与宿主通信——主程序与插件之间是进程边界，互不传染对方协议义务。

## 用户决策偏好域（PluginContext 三方法，契约 v13）

插件经用户问答沉淀的**决策记忆**存储：插件问（自建 dialog + 前端通信）→ 用户答 → 落宿主偏好域 → 用户可在主程序记忆管理页查看与删除，删除后插件下次**重新问**。三方法挂在 `PluginContext`（`dto.PluginContext`）：

```go
// 读：无记录返回 (nil, false, nil) 不报错——据此回落重新发起问答
v, ok, err := ctx.GetPreference("illust.form")

// 写：整值覆写（同键已存在则整体重写），值恒为完整信封
err = ctx.SetPreference("illust.form", &sdkdto.PreferenceValue{
    SchemaVersion: 1,             // 值结构版本，初值 1，规则类值演进时自管升版
    Title:    "图文形态",          // 写入时随值携带，供宿主记忆管理页展示
    Data:     `{"form":"multi"}`, // 插件自定义负载（JSON 文本），宿主只存不解释
})

// 列：本插件全部偏好键（跨插件键互不可见）
keys, err := ctx.ListMyPreferences()
```

要点：

- **与 GetValue/SetValue 的边界**（一句话判据）：删掉它之后用户会被重新问吗？会 → 偏好域；不会 → settings/plugin_storage。偏好域存「用户的决策」，plugin_storage 存「插件的配置」。
- **写入纪律**（约定不强制）：仅写经用户问答确认的决策——偏好用户可见（宿主记忆管理页），把插件内部状态当偏好写入会污染用户可见的记忆列表。
- **无删除方法**：「忘掉」是用户权利，删除仅经宿主管理面——插件只能覆写不能销毁记忆。
- **旧宿主降级**：契约 <13 的宿主未实现对应 RPC，调用得 gRPC `Unimplemented`，应优雅降级（回落为每次都问），不当数据错误处理。

## 流等待期心跳（契约 v14）

插件在流式 RPC（`Create`/`Start`/`Resume`）的 handler 执行期可周期上报**心跳块**保活当前流，越过宿主 60 秒空闲超时窗（`liveness.ReaderIdleTimeout`）——适用 handler 内的长等待（等用户输入、串行外部请求、重试退避循环）。语义＝「本等待点有界且仍在推进」：**每个上报心跳的等待点须保有自身超时**（弹窗 dismiss 上限、二维码有效期、HTTP 超时），在永不收口的等待里上报会让宿主失效 hang 检测。

- **接入**：`Create` 实现可选接口 `dto.HeartbeatCreateFetcher`（`CreateWithHeartbeat(url, reporter)`）；`Start`/`Resume` 经 `dto.HeartbeatFromContext(ctx)` 取用（未注入返回 `dto.NoopHeartbeat`，免判空）。上报器随每次 RPC 调用到达、按流程独立。
- **调用**：`reporter.Heartbeat()` 并发安全、限频 20 秒（`liveness.HeartbeatInterval`），可任意 tick 频率调用；生命周期由 SDK 管理（handler 返回即关闭），等待终结时须停止自挂的 ticker。旧宿主（契约 <14，含未协商）为 no-op，可按 `transport.HostContractVersion()` 自适应。
- **不支持**：`Start`/`Resume` 返回后的 pull 数据传输期无心跳——传输停滞超窗被杀是正确的 hang 检测。完整约定见主仓 `doc/plugin-dev-guide.md` 6.1「流等待期心跳」。

## 设置驱动的派生面热生效（resolver 契约）

插件可自带 resolver 脚本：宿主在激活完成后与每次设置落库后，以插件全量设置为输入求值它，得到各派生面（作品拉取/作者拉取/站点浏览器/资源类型/前端扩展）上已声明条目的参与度快照。`settingresolver/` 提供 TS 契约类型与本地 harness，**非破坏交付**——gRPC/proto 零改动、契约版本零 bump：

- **`settingresolver/contract.d.ts`** — resolver 契约类型（输入 `ResolverInput`、输出 `ResolverOutput`、条目 `ResolverEntry`、入口函数 `ResolverFn`），逐字镜像宿主 `backend/plugin/settingresolver/contract.go` 文档化的调用约定。宿主为契约权威源，本文件只提供类型面。纯 JS 脚本仓可经 JSDoc 引用：`/** @type {import('library-squirrel-sdk/settingresolver/contract').ResolverFn} */`（按本仓实际检出路径调整导入串）。
- **`settingresolver/harness.mjs`** — 本地 harness（零依赖，node 运行）：在镜像宿主求值环境的沙箱（全新运行时、零宿主绑定、`Date`/`Math.random` 已删除、500ms 超时、64KB 体积上限）中对样本设置跑插件自己的 resolver，报告通过/失败与输出 shape 校验结果；给 `--manifest plugin.json` 时以清单默认设置 dry-run 并核对声明集（镜像安装闸门）。

```bash
node settingresolver/harness.mjs resolver.js --manifest plugin.json \
  --settings '{"enableParticipation": "false"}'
```

脚本编写要点：顶层定义 `function resolve(input)`；输入 `input.settings` 的值以存储形态喂入（设置值统一以 string 存储，boolean 设置为 `"true"`/`"false"`，比较前自行归一）；输出 `{version: 1, entries: [...]}` 为快照语义（未列出条目 = 基线参与）；纯函数纪律（同一输入必产出同一输出）；语法面锚定 ES2015 常用语法（const/箭头/模板串/解构可用，async/await 与可选链等 ES2017+ 勿用——node harness 的语法面宽于宿主引擎，harness 通过不代表宿主语法面通过）。完整契约与失败语义见主仓 `doc/plugin-dev-guide.md` 的 resolver 章节。

## 许可协议

本项目采用 [MIT](LICENSE) 许可协议。

Copyright (c) 2026 lvfeng
