package dto

// SiteTagDTO 站点标签数据传输对象
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

// SiteTagFullDTO 站点标签完整DTO
type SiteTagFullDTO struct {
	SiteTag  *SiteTagDTO  `json:"siteTag,omitempty"`
	LocalTag *LocalTagDTO `json:"localTag,omitempty"`
	Site     *SiteDTO     `json:"site,omitempty"`
}

// SiteTagLocalRelateDTO 站点标签与本地标签关联DTO
type SiteTagLocalRelateDTO struct {
	SiteTag             *SiteTagDTO  `json:"siteTag,omitempty"`
	LocalTag            *LocalTagDTO `json:"localTag,omitempty"`
	Site                *SiteDTO     `json:"site,omitempty"`
	HasSameNameLocalTag bool         `json:"hasSameNameLocalTag"`
}
