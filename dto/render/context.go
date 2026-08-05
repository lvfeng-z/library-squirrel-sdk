// Package render 定义插件资源渲染契约类型。
//
// 本包类型供插件 resourceViewer 扩展点（前端被动响应型渲染器）作为注入 props，
// 是插件与主程序之间的前端渲染契约：不跨 gRPC（区别于 SDK dto 包的 A 类 gRPC 契约），
// 受主程序 contractVersion 保护——破坏性变更（删/改字段）须提升契约版本。
//
// 字段集独立演进，与主程序 backend/base/model/dto 的展示 DTO 解耦：主程序展示 DTO
// 随前端展示需求变更不传导至本包，仅当「决定向插件暴露新的渲染信息」时才在此扩展
// （向前兼容地加字段，不提升契约版本）。A 类稳定子类型（WorkDTO/SiteDTO/LocalTagDTO/
// LocalAuthorDTO，别名自 proto 生成包）直接复用 sdkdto.*，不在此重定义。
package render

import sdkdto "github.com/lvfeng-z/library-squirrel-sdk/dto"

// Context 主程序渲染某 resourceType 资源、命中插件渲染器时注入的完整 props。
type Context struct {
	Work         *sdkdto.WorkDTO       `json:"work,omitempty"`
	LocalAuthors []*RankedLocalAuthor  `json:"localAuthors,omitempty"`
	SiteAuthors  []*RankedSiteAuthor   `json:"siteAuthors,omitempty"`
	Site         *sdkdto.SiteDTO       `json:"site,omitempty"`
	LocalTags    []*sdkdto.LocalTagDTO `json:"localTags,omitempty"`
	SiteTags     []*SiteTagFullDTO     `json:"siteTags,omitempty"`
	Resource     *ResourceFullDTO      `json:"resource,omitempty"`
}

// PersistentStoreDTO 文件持久存储信息。
type PersistentStoreDTO struct {
	ID                int64   `json:"id"`
	FilePath          *string `json:"filePath"`
	FileName          *string `json:"fileName"`
	FilenameExtension *string `json:"filenameExtension"`
	Status            int     `json:"status"`  // 0=未完成，1=完成
	Width             int     `json:"width"`   // 图片宽度（像素），非图片为 0
	Height            int     `json:"height"`  // 图片高度（像素），非图片为 0
	CreateTime        int64   `json:"createTime"`
	UpdateTime        int64   `json:"updateTime"`
}

// ResourceStoreDTO Resource 关联的单个 typed store。
type ResourceStoreDTO struct {
	StoreType  string              `json:"storeType"`  // image | document | thumbnail | videoTrack | audioTrack | videoMain
	Generation string              `json:"generation"` // downloaded | derived
	Store      *PersistentStoreDTO `json:"store,omitempty"`
}

// ResourceFullDTO 资源完整信息。Stores 为全量多轨 store（主数据源）；WorkStore 为展示主体
//（按资源类型 PrimaryRoles 优先级链派生）；ThumbnailStore 为缩略图便捷访问器。
type ResourceFullDTO struct {
	ID               int64               `json:"id"`
	WorkID           int64               `json:"workId"`
	TaskID           int64               `json:"taskId"`
	Enabled          bool                `json:"enabled"`
	SuggestName      *string             `json:"suggestName"`
	ResourceType     string              `json:"resourceType"`
	ResourceComplete int                 `json:"resourceComplete"`
	Stores           []ResourceStoreDTO  `json:"stores,omitempty"`
	WorkStore        *PersistentStoreDTO `json:"workStore,omitempty"`
	ThumbnailStore   *PersistentStoreDTO `json:"thumbnailStore,omitempty"`
	CreateTime       int64               `json:"createTime"`
	UpdateTime       int64               `json:"updateTime"`
}

// SiteAuthorDTO 站点作者。
type SiteAuthorDTO struct {
	ID                   int64   `json:"id"`
	SiteID               *int64  `json:"siteId"`
	SiteAuthorID         *string `json:"siteAuthorId"`
	AuthorName           *string `json:"authorName"`
	FixedAuthorName      *string `json:"fixedAuthorName"`
	SiteAuthorNameBefore *string `json:"siteAuthorNameBefore"`
	Introduce            *string `json:"introduce"`
	Homepage             *string `json:"homepage"`
	LocalAuthorID        *int64  `json:"localAuthorId"`
	LastUse              *int64  `json:"lastUse"`
	CreateTime           int64   `json:"createTime"`
	UpdateTime           int64   `json:"updateTime"`
}

// RankedLocalAuthor 带角色与排序的本地作者。
type RankedLocalAuthor struct {
	Author    sdkdto.LocalAuthorDTO `json:"author"`
	RoleName  string                `json:"roleName"`
	SortOrder int                   `json:"sortOrder"`
}

// RankedSiteAuthor 带角色与排序的站点作者。
type RankedSiteAuthor struct {
	Author    SiteAuthorDTO `json:"author"`
	RoleName  string        `json:"roleName"`
	SortOrder int           `json:"sortOrder"`
}

// SiteTagDTO 站点标签。
type SiteTagDTO struct {
	ID            int64   `json:"id"`
	SiteID        *int64  `json:"siteId"`
	SiteTagID     *string `json:"siteTagId"`
	SiteTagName   *string `json:"siteTagName"`
	BaseSiteTagID *string `json:"baseSiteTagId"`
	Description   *string `json:"description"`
	LocalTagID    *int64  `json:"localTagId"`
	LastUse       *int64  `json:"lastUse"`
	CreateTime    int64   `json:"createTime"`
	UpdateTime    int64   `json:"updateTime"`
}

// SiteTagFullDTO 站点标签完整信息（含绑定的本地标签与来源站点）。
type SiteTagFullDTO struct {
	SiteTag  *SiteTagDTO         `json:"siteTag,omitempty"`
	LocalTag *sdkdto.LocalTagDTO `json:"localTag,omitempty"`
	Site     *sdkdto.SiteDTO     `json:"site,omitempty"`
}
