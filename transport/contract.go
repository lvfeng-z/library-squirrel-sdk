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
const ContractVersion = 16

// VersionEntry 契约版本史单条目。
type VersionEntry struct {
	// Number 契约代际号（表内自 1 连续编号，与位置一致）
	Number int
	// Summary 变更全量叙述：改了什么、破坏性质（源级/线级/非破坏）、
	// minSupportedContractVersion 联动与新旧版本组合的兼容行为——
	// bump 时必须把这几项写全，它是文档侧版本史的渲染源
	Summary string
}

// ContractHistory 契约版本史（单一源，条目自 1 连续编号至 ContractVersion）。
//
// 本表是各代契约变更叙述的唯一权威源：主仓 doc/plugin-dev-guide.md
// 「契约版本协商」节的版本史区块由本表逐条渲染生成（"- N — Summary"），
// 同步由主仓 backend/plugin/extension/contract_version_test.go 的
// TestDevGuideContractHistorySync 校验。修改流程：先改本表
// （bump ContractVersion 常量 + 追加条目），再运行该测试，按失败信息中的
// 期望文本整段替换文档区块——禁止在文档侧手改条目内容。
var ContractHistory = []VersionEntry{
	{1, "初始契约：A 类 proto 单源、能力声明化、render.Context 断链契约。"},
	{2, "GetValue/GetAllValues 返回带 schemaVersion（配置 schema 版本感知）。"},
	{3, "资源类型扩展：插件可经 manifest `resourceTypes` 段声明自定义 ResourceType，audio 作为内置资源类型。"},
	{4, "StoreSpec 加 `expectedSha256`（来源侧声明期望哈希，下载完整性校验）；Task 删 `pendingResourceId`（主程序任务执行面暂存模式改造，资源定位改由主程序按 task_id 直查，插件零消费）。"},
	{5, "落盘路径查询 RPC（GetStoreRelPath）整链退役：最终落盘路径改由 SDK `storepath` 派生函数本地推导（插件据任务身份 + specs 顺序自算，见 6.1 StoreSpec 顺序确定性条款），插件不再向主程序查询；主程序 `minSupportedContractVersion` 同步升 5，未声明版本的插件拒载。"},
	{6, "新增 `LibraryQuery` 库查询服务（Tier 1 只读：作品/资源与 store/作者/标签/作品集/站点/工作目录查询，身份键复合寻址、无界集合强制分页、默认只返回活数据，21 个端点见 5.1）；删除 HostService 死声明 `GetWorkSetBySiteWorkSetId`（无桥接无调用的废弃 RPC，查询能力吸收为 `LibraryQuery.GetWorkSetBySiteKey`，按 (site_key, site_work_set_id) 复合键寻址——删 RPC 属破坏性变更故升版）。主程序 `minSupportedContractVersion` 保持 5（v5 既有捆绑包仍可加载；升 6 属发布时重建捆绑包的动作）。"},
	{7, "周边数据写面契约（作品及周边数据统一写面首期）：任务声明期周边三 DTO（`TaskSiteAuthorDTO`/`TaskSiteTagDTO`/`TaskWorkSetDTO`）加可选 `siteKey` 字段——周边数据跨站寻址：声明站点≠作品站点时 find-only 引用既有行（宿主按站点表键解析，未注册键/缺行报错，不建行不改行），缺省空=作品所属站点（本站 upsert 行为不变，见 6.1「作品及周边数据写面契约」）。加字段向前兼容，作为周边写面新能力标识升版——主程序 `minSupportedContractVersion` 保持 5（v5/v6 插件混装载不受影响）。"},
	{8, "关联级维度体系（tag namespace 与 author role 同构为作品-实体关联行上的开放维度）：ns 从 site_tag 实体行收回关联级——`SiteTagInfo` 删 `Namespace` 字段（字段号 6 reserved 不复用，删字段属源级破坏），`TaskSiteTagDTO.Namespace` 保留、语义=本作品上该标签的关联级 ns；role 同构补齐——`TaskSiteAuthorDTO` 加 `RoleName` 声明面，`ListAuthorsByWorkId` 返回面由实体级 DTO 整体改为关联条目（`WorkLocalAuthorEntry`/`WorkSiteAuthorEntry`，与标签侧 ListTagsByWorkId 返回形态对称，携带关联级 role_name）——返回消息字段类型更换属线级破坏（旧编译插件按旧消息类型解析会错读）。主程序 `minSupportedContractVersion` 同步升 8（v8 以下插件拒载，捆绑包随之重建）。"},
	{9, "插件声明面重构（能力包模型）：顶层 `capabilities` 段取消，其声明移入 `extensions` 段——`siteAuthorFetch` 携 `sites` 作用域、`workOrderQuery` 与 `workSetRelationQuery` 下沉 `workFetch[].options`、`resourceTypeProvider` 取消（`resourceTypes` 迁入 `extensions` 段后「段存在即启用」），门控粒度相应改为（插件, 扩展点）条目级；站点归属自判机制退役——插件侧身份键比对辅助、跨进程未归属错误信号及其 gRPC 状态码转译与宿主侧判定一并删除，作者拉取候选改由宿主按插件已声明的站点范围收窄，插件不再自判归属。声明面结构更换与导出符号删除属源级破坏——主程序 `minSupportedContractVersion` 同步升 9（v9 以下插件拒载，捆绑包随之重建）。"},
	{10, "插件清单结构变更：用户设置项声明（settings 段）住清单根级（`extensions` 段只承载能力包声明，不承载 settings 子段，其子对象内该键在场即判不合格）。段位置变更属宿主读清单的源级破坏——主程序 `minSupportedContractVersion` 同步升 10（低于 10 的清单在加载期拒收，捆绑包随之重建）。"},
	{11, "注册面声明化 + siteAuthorFetch 实例化：`extensions.siteAuthorFetch` 由单对象 `{sites}` 改数组 `[{id,name,sites}]`（条目 id 插件内唯一、name 必填）；拉取请求 `FetchSiteAuthorInfoRequest` 加 `extensionId`（候选粒度从插件升为（插件, 条目）全键，插件侧服务端按条目 id 分派、未命中报 InvalidArgument，`WithSiteAuthorFetcher` 改为按条目可多次注册）；`workFetch`/`siteBrowsers` 的 `name` 变必填并由宿主激活期按清单条目派生注册（元数据 name/description 取清单）；`workFetch[]` 加可选 `urlPatterns`（URL 监听迁清单，宿主激活期建派生索引）；HostService 五个运行时注册/监听 RPC（RegisterTaskHandler/RegisterSiteBrowser/UnregisterSiteBrowser/RegisterUrlListener/UnregisterUrlListener）与 `PluginContext` 对应方法整体退役。删 RPC 属线级破坏——主程序 `minSupportedContractVersion` 同步升 11（v10 及以下插件包拒载并提示升级，捆绑包随之重建）。"},
	{12, "扩展点正名：作品拉取扩展点（workFetch / WorkFetcher / WorkFetchService）由上一代旧名（taskHandlers / TaskHandler / TaskHandlerService）正名而来——清单段名、SDK 接口与选项、gRPC 服务名、宿主侧类型与状态字段一并更换。清单段名更换属宿主读清单的源级破坏、gRPC 服务名更换属线级破坏——主程序 `minSupportedContractVersion` 同步升 12（v11 及以下插件包拒载并提示升级，捆绑包随之重建）。"},
	{13, "用户决策偏好域（插件偏好记忆）：HostService 线级新增 `GetPreference`/`SetPreference`/`ListMyPreferences` 三 RPC（插件经用户问答沉淀的决策记忆，与 `plugin_storage` 配置面正交，见 8.5；无删除 RPC——删除仅经宿主记忆管理面），SDK `PluginContext` 配套三方法。线级新增、非破坏——旧插件不调新 RPC，契约 12 插件在新宿主照常运行，故 `minSupportedContractVersion` 维持 12；新插件对旧宿主调用偏好面得 gRPC `Unimplemented`（降级纪律见 5.1「Unimplemented 降级」），既有面不受影响。"},
	{14, "流等待期心跳（插件在流式 RPC 的 handler 执行期周期向该流发送的保活块，语义=「本等待点有界且仍在推进」，Create/Start/Resume 三流一次定形）：CreateChunk/StreamChunk 双流 oneof 各增 Heartbeat 态（共享空消息 Heartbeat，留字段扩展位），`ActivateRequest` 增 `host_contract_version` 协商字段（宿主随 Activate 下发自身契约版本，插件侧运行时门控心跳发送——旧宿主上心跳 no-op，线形态与升级前一致）；SDK 新增通用心跳上报器（并发安全+限频+handler 返回后关闭）与两条接入通道（Create 可选接口 HeartbeatCreateFetcher、Start/Resume ctx 注入 HeartbeatFromContext），维持宿主的流空闲检测窗（越过 Create 首块 / Start·Resume 首响应的 60 秒空闲超时，见 6.1「流等待期心跳」）。线级新增非破坏（不接入心跳的插件含全部旧插件线形态零变化）——`minSupportedContractVersion` 维持 12；作为能力标识升版的依据=proto 新增共享消息与两流新 payload 态，宿主据此认定对心跳块的容忍度；契约 14 的插件包对旧宿主仍在加载期被拒。"},
	{15, "宿主代理解析服务：HostService 线级新增 `ResolveProxy` RPC（插件出网代理三级检测收归宿主：显式 > Windows 系统代理〔注册表〕> 环境变量，按请求目标地址匹配 scheme/NO_PROXY，逐请求现查无缓存；SDK `proxy` 包以默认 Transport 消费，见 7.1），SDK `PluginContext` 配套 `ResolveProxy` 方法 + `proxy` 包默认 Transport 构造器（RPC 决议每请求求值、每次调用 1 秒超时预算，失败/超时/Unimplemented 降级为「显式 > 环境变量」，请求不因解析失败或超时而失败）。线级新增、非破坏——旧插件不调新 RPC，契约 12 插件在新宿主照常运行，故 `minSupportedContractVersion` 维持 12；新插件对旧宿主调用解析面得 gRPC `Unimplemented`（SDK `proxy` 包走降级链，降级纪律见 5.1「Unimplemented 降级」），既有面不受影响。"},
	{16, "插件设置变更通知：PluginLifecycle 线级新增 `SettingChanged` 一元 RPC——宿主在 SaveSetting/ResetSetting 落库成功后向已激活插件异步推送 {source, keys} 纯通知（source=save/reset + 变更键集），响应 Empty 无回执面（处置语义由插件处置函数自理，SDK 经 `WithSettingChangeHandler` 注册处置函数，见四节「插件入口」）。线级新增、非破坏、纯通知无回执——未注册处置函数时 SDK 空操作+debug 日志（旧插件内嵌 Unimplemented 兜底，宿主静默降级；新插件对旧宿主只是收不到通知），宿主侧通知尽力而为（插件未激活跳过、发送失败/超时降级为 debug 日志），不构成保存流程的失败面——`minSupportedContractVersion` 维持 12，既有面不受影响。"},
}
