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
const ContractVersion = 5

// 版本历史：
//   1 — 初始契约：A 类 proto 单源、能力声明化、render.Context 断链契约（C 节点）
//   2 — GetValue/GetAllValues 返回带 schemaVersion：配置 schema 版本感知（E 节点）
//   3 — 资源类型扩展：插件可经 manifest resourceTypes 段声明自定义 ResourceType；audio 作为内置资源类型
//   4 — StoreSpecMeta 加 expectedSha256（来源侧声明期望哈希，下载完整性校验）；Task 删 pendingResourceId
//       （主程序任务执行面暂存模式改造，资源定位改由主程序按 task_id 直查，插件零消费）。
//       实体读写 RPC（谱系 S 节点）未并入，落地时再 bump
//   5 — GetStoreRelPath RPC 整链退役（落盘身份键命名改造：最终路径由 SDK storepath 派生函数本地推导，
//       插件不再向主程序查询落盘路径；主程序 minSupportedContractVersion 同步升 5，未声明版本插件拒载）
