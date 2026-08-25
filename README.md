# library-squirrel-sdk

[LibrarySquirrel](https://gitee.com/lv__feng/library-squirrel) 的插件开发 SDK：定义插件与宿主之间的 gRPC 协议（proto 契约与生成代码）、`PluginContext` 宿主能力访问绑定，以及插件渲染契约类型（`dto/render`）。

插件以独立进程运行，经本 SDK 与宿主通信——主程序与插件之间是进程边界，互不传染对方协议义务。

## 许可协议

本项目采用 [MIT](LICENSE) 许可协议。

Copyright (c) 2026 lvfeng
