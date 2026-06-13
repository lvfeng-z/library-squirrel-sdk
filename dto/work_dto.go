package dto

// WorkDTO 作品数据传输对象
type WorkDTO struct {
	ID                  int64   `json:"id"`
	SiteID              *int64  `json:"siteId"`
	SiteWorkID          *string `json:"siteWorkId"`
	SiteWorkName        *string `json:"siteWorkName"`
	SiteAuthorID        *string `json:"siteAuthorId"`
	SiteWorkDescription *string `json:"siteWorkDescription"`
	SiteUploadTime      *int64  `json:"siteUploadTime"`
	SiteUpdateTime      *int64  `json:"siteUpdateTime"`
	NickName            *string `json:"nickName"`
	LocalAuthorID       *int64  `json:"localAuthorId"`
	LastView            *int64  `json:"lastView"`
	CreateTime          int64   `json:"createTime"`
	UpdateTime          int64   `json:"updateTime"`
}

// WorkFullDTO 作品完整信息DTO
type WorkFullDTO struct {
	Work         *WorkDTO             `json:"work,omitempty"`
	LocalAuthors []*RankedLocalAuthor `json:"localAuthors,omitempty"`
	SiteAuthors  []*RankedSiteAuthor  `json:"siteAuthors,omitempty"`
	Site         *SiteDTO             `json:"site,omitempty"`
	LocalTags    []*LocalTagDTO       `json:"localTags,omitempty"`
	SiteTags     []*SiteTagFullDTO    `json:"siteTags,omitempty"`
	Resource     *ResourceFullDTO     `json:"resource,omitempty"` // 单个活跃资源（含 PersistentStore 信息）
}
