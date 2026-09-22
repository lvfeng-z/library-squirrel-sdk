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
const ContractVersion = 11

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
//   8 — 关联级维度体系（tag namespace 与 author role 同构为作品-实体关联行上的开放维度）：
//       SiteTagInfo 删 namespace 字段（ns 由作品-标签关联行承载，标签实体不承载；字段号 6 reserved 不复用）；
//       TaskSiteTagDTO.namespace 语义定为关联级（本作品上该标签的 ns）；TaskSiteAuthorDTO 加 roleName
//       （声明面：本作品上该作者的 role）；ListAuthorsByWorkId 返回面由实体级 DTO 整体改为关联条目
//       （WorkLocalAuthorEntry/WorkSiteAuthorEntry，与标签侧 ListTagsByWorkId 返回形态对称，携带关联级
//       role_name）。返回消息字段类型更换属线级破坏（旧编译插件按旧消息类型解析会错读），SiteTagInfo
//       删字段属源级破坏——主程序 minSupportedContractVersion 须同步升至 8，min<8 即宣称兼容自己已破坏
//       的线协议
//   9 — 插件声明面重构（能力包模型）：声明面移入 extensions 段条目级——顶层 capabilities 由
//       extensions 各条目的包内声明取代（siteAuthorFetch 携 sites 作用域、workOrderQuery 与
//       workSetRelationQuery 下沉 taskHandlers[].options、resourceTypeProvider 取消），门控粒度
//       相应改为（插件, 扩展点）条目级；站点归属自判机制退役——插件侧的身份键比对辅助、跨进程
//       未归属错误信号及其 gRPC 状态码转译与宿主侧判定一并删除，作者拉取候选改由宿主按插件已
//       声明的站点范围收窄，插件不再自判归属。声明面结构更换与导出符号删除属源级破坏——
//       主程序 minSupportedContractVersion 须同步升至 9
//  10 — 插件清单结构变更：用户设置项声明（settings 段）住清单根级（extensions 段只承载
//       能力包声明，不承载 settings 子段）。段位置变更属宿主读清单的源级破坏——主程序
//       minSupportedContractVersion 须同步升至 10，低于 10 的清单在加载期拒收
//  11 — siteAuthorFetch 数组化 + 注册面声明化：拉取请求 FetchSiteAuthorInfoRequest 加
//       extensionId（候选粒度从插件升为（插件, 条目）全键，插件侧服务端按条目 id 分派、
//       未命中报 InvalidArgument，WithSiteAuthorFetcher 改为按条目可多次注册）；HostService
//       五个运行时注册/监听 RPC 退役（RegisterTaskHandler/RegisterSiteBrowser/
//       UnregisterSiteBrowser/RegisterUrlListener/UnregisterUrlListener）——taskHandlers/
//       siteBrowsers 由宿主激活期按清单条目派生注册，URL 监听迁入清单 urlPatterns 字段。
//       删 RPC 属线级破坏——主程序 minSupportedContractVersion 须同步升至 11
