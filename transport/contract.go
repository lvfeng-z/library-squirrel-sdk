package transport

// ContractVersion 当前插件契约版本（业务契约，非 go-plugin 传输协议版本）。
//
// 主程序与插件之间的 DTO/RPC/能力契约的代际编号。破坏性变更必 +1：
// 删/改 proto 字段类型、改 DTO 结构、改 RPC 签名、改前端 props 契约。
// proto 加字段本身向前兼容（旧插件忽略新字段）不 bump；但当加字段开启一族
// 新能力（宿主/插件按契约版本识别该能力是否可用）时，作为能力标识 +1。
//
// 与 Handshake.ProtocolVersion 分工：
//   - ProtocolVersion 是 gRPC 传输握手版本（hashicorp/go-plugin，主程序与插件
//     引用同一 Handshake 常量，编译期锁定，无需运行时校验）。
//   - ContractVersion 是业务契约版本（插件 manifest 声明编译时锁定的契约版本，
//     主程序加载时与 currentContractVersion / minSupportedContractVersion 比对，
//     过新/过旧均拒绝加载）。
const ContractVersion = 7

// 版本历史：
//   1 — 初始契约：A 类 proto 单源、能力声明化、render.Context 断链契约（C 节点）
//   2 — GetValue/GetAllValues 返回带 schemaVersion：配置 schema 版本感知（E 节点）
//   3 — 资源类型扩展：插件可经 manifest resourceTypes 段声明自定义 ResourceType；audio 作为内置资源类型
//   4 — StoreSpecMeta 加 expectedSha256（来源侧声明期望哈希，下载完整性校验）；Task 删 pendingResourceId
//       （主程序任务执行面暂存模式改造，资源定位改由主程序按 task_id 直查，插件零消费）。
//       实体读写 RPC（谱系 S 节点）未并入，落地时再 bump
//   5 — GetStoreRelPath RPC 整链退役（落盘身份键命名改造：最终路径由 SDK storepath 派生函数本地推导，
//       插件不再向主程序查询落盘路径；主程序 minSupportedContractVersion 同步升 5，未声明版本插件拒载）
//   6 — 新增 LibraryQuery 服务（Tier 1 库查询契约：作品/资源与 store/作者/标签/作品集/站点/工作目录
//       只读查询，身份键复合寻址、无界集合强制分页、默认只返回活数据）；删除 HostService 死声明
//       GetWorkSetBySiteWorkSetId（无桥接无调用的废弃 RPC，查询能力吸收为 LibraryQuery.GetWorkSetBySiteKey，
//       按 (site_key, site_work_set_id) 复合键寻址——删 RPC 属破坏性变更故升版）
//   7 — 周边数据写面契约（作品及周边数据统一写面首期）：任务声明期周边三 DTO
//       （TaskSiteAuthorDTO/TaskSiteTagDTO/TaskWorkSetDTO）加可选 siteKey 字段——周边数据跨站寻址：
//       声明站点≠作品站点时 find-only 引用既有行（不存在报错），缺省空=作品所属站点（本站 upsert 行为
//       不变）；加字段向前兼容，作为周边写面新能力标识升版（主程序 minSupportedContractVersion 维持 5，
//       v5/v6 插件混装载不受影响）
