package dto

// SiteAuthorDTO 站点作者数据传输对象
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

// SiteAuthorFullDTO 站点作者完整DTO
type SiteAuthorFullDTO struct {
	SiteAuthor  *SiteAuthorDTO  `json:"siteAuthor,omitempty"`
	LocalAuthor *LocalAuthorDTO `json:"localAuthor,omitempty"`
	Site        *SiteDTO        `json:"site,omitempty"`
}

// SiteAuthorLocalRelateDTO 站点作者与本地作者关联DTO
type SiteAuthorLocalRelateDTO struct {
	SiteAuthor             *SiteAuthorDTO  `json:"siteAuthor,omitempty"`
	LocalAuthor            *LocalAuthorDTO `json:"localAuthor,omitempty"`
	Site                   *SiteDTO        `json:"site,omitempty"`
	HasSameNameLocalAuthor bool            `json:"hasSameNameLocalAuthor"`
}

// RankedSiteAuthor 带排序的站点作者
type RankedSiteAuthor struct {
	Author    SiteAuthorDTO `json:"author"`
	RoleName  string        `json:"roleName"`
	SortOrder int           `json:"sortOrder"`
}

// RankedSiteAuthorWithWorkId 带作品ID的站点作者
type RankedSiteAuthorWithWorkId struct {
	Author    SiteAuthorDTO `json:"author"`
	RoleName  string        `json:"roleName"`
	SortOrder int           `json:"sortOrder"`
	WorkId    int64         `json:"workId"`
}
