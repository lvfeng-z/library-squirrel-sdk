package transport

// ContractVersion 当前插件契约版本（业务契约，非 go-plugin 传输协议版本）。
//
// 主程序与插件之间的 DTO/RPC/能力契约的代际编号。仅破坏性变更才 +1：
// 删/改 proto 字段类型、改 DTO 结构、改 RPC 签名、改前端 props 契约。
// proto 加字段（向前兼容，旧插件忽略新字段）不 bump。
//
// 与 Handshake.ProtocolVersion 分工：
//   - ProtocolVersion 是 gRPC 传输握手版本（hashicorp/go-plugin，主程序与插件
//     引用同一 Handshake 常量，编译期锁定，无需运行时校验）。
//   - ContractVersion 是业务契约版本（插件 manifest 声明编译时锁定的契约版本，
//     主程序加载时与 currentContractVersion / minSupportedContractVersion 比对，
//     过新/过旧均拒绝加载）。
const ContractVersion = 2

// 版本历史：
//   1 — 初始契约：A 类 proto 单源、能力声明化、render.Context 断链契约（C 节点）
//   2 — GetValue/GetAllValues 返回带 schemaVersion：配置 schema 版本感知（E 节点）
