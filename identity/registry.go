// Package identity 站点跨库标识注册表——site_key 的唯一权威分配来源。
//
// 键值为站点品牌 slug（小写字母/数字/连字符，^[a-z][a-z0-9-]{1,30}$，如 "pixiv"），
// 由注册表中心分配：新站点经 PR 向本包贡献常量，维护者查重（键值与既有条目、
// 与站点官方品牌名一致性）后合并，随 SDK 发布生效。键一经发布不可变，永不跟随
// 站点展示名变化；键与名职责分离——键是跨库身份（比较键），名是展示（可改可重名）。
// 注册表只增不改（已收录条目的键永不重排、永不复用）。
//
// 治理分层：站点键分配是编译期内置的强中心化保证（未注册键被主程序拒绝）；
// 站点侧实体 ID（siteWorkId/siteTagId/siteAuthorId/siteWorkSetId）的忠实性
// 运行时不可校验，靠各常量附带的「站点级 ID 约定注记」+ 插件准入审查收敛。
package identity

// Site 注册表条目：站点的跨库身份键与权威展示信息。
type Site struct {
	// Key 站点唯一身份键（品牌 slug：^[a-z][a-z0-9-]{1,30}$，如 "pixiv"）。
	// 跨库站点匹配、查重、关联一律以此键为准；键一经发布不可变，
	// 站点名仅作展示，不参与身份判定
	Key string
	// Name 权威站点名（纯展示）。站点行创建时取此值，插件自报站点名被忽略。
	Name string
	// Homepage 站点主页（纯展示）。本地虚拟站点无主页，为空串。
	Homepage string
}

// 注册表。常量即引用（编译期约束）、godoc 即规范；条目只增不改。
var (
	// Pixiv 插画社区站点。
	//
	// 站点级 ID 约定注记（同站点多插件产出收敛到同一条 norm）：
	//   siteWorkId    = illust id（多页作品的子任务为页级 id，取自原图 URL）
	//   siteTagId     = tag 词（tag 名即 id）
	//   siteAuthorId  = 作者 user id
	//   siteWorkSetId = illust id（多页作品的作品集即该作品自身）
	Pixiv = Site{Key: "pixiv", Name: "pixiv", Homepage: "https://www.pixiv.net"}

	// Bilibili 视频社区站点。
	//
	// 站点级 ID 约定注记（同站点多插件产出收敛到同一条 norm）：
	//   siteWorkId    = 视频：父任务 bvid、分 P 子任务 {bvid}_{cid}；
	//                   图文动态：父任务动态 id、子任务 {动态 id}_{序号}；
	//                   专栏：专栏 id 原值
	//   siteTagId     = 标签名（tag 名即 id）
	//   siteAuthorId  = UP 主 mid（十进制字符串）
	//   siteWorkSetId = 多 P 视频的 bvid（一个多 P 视频即一个作品集）
	Bilibili = Site{Key: "bilibili", Name: "bilibili", Homepage: "https://www.bilibili.com"}

	// Local 本地导入虚拟站点——官方捆绑插件 localImport 的落库站点，无主页。
	// 与外部站点同构入表，键分配无差别。
	//
	// 站点级 ID 约定注记：
	//   siteWorkId    = 文件内容哈希（单文件导入）；local-dir-{目录相对路径}（目录导入）
	//   siteTagId     = siteTag:{标签名}
	//   siteAuthorId  = siteAuthor:{作者名}
	//   siteWorkSetId = workSet:{作品集名}
	Local = Site{Key: "local", Name: "local"}
)

// registry 注册表条目的有序承载，条目顺序即注册序；全量枚举（All）的稳定输出序
// 以此切片承载（map 无序，不能作序源）。新条目只追加到尾部。
var registry = []Site{
	Pixiv,
	Bilibili,
	Local,
}

// registryIndex 键到条目的索引，供 Lookup 检索；由 registry 派生，非独立数据源。
var registryIndex = func() map[string]Site {
	index := make(map[string]Site, len(registry))
	for _, site := range registry {
		index[site.Key] = site
	}
	return index
}()

// Lookup 按键查注册表条目；主程序的键校验与站点行创建共用此入口。
// 未注册键（含空键）返回 false，拒绝策略由调用方决定。
func Lookup(key string) (Site, bool) {
	s, ok := registryIndex[key]
	return s, ok
}

// All 返回注册表全量条目，按注册序稳定输出；供宿主将注册表投影为主库站点行
// 等需要遍历全部已注册站点的场景。返回切片为注册表副本，调用方修改不影响注册表。
func All() []Site {
	all := make([]Site, len(registry))
	copy(all, registry)
	return all
}
