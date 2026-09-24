# library-squirrel-sdk

[LibrarySquirrel](https://gitee.com/lv__feng/library-squirrel) 的插件开发 SDK：定义插件与宿主之间的 gRPC 协议（proto 契约与生成代码）、`PluginContext` 宿主能力访问绑定，以及插件渲染契约类型（`dto/render`）。

插件以独立进程运行，经本 SDK 与宿主通信——主程序与插件之间是进程边界，互不传染对方协议义务。

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
