package dto

// WorkSetDTO 作品集数据传输对象
type WorkSetDTO struct {
	ID                     int64   `json:"id"`
	SiteID                 *int64  `json:"siteId"`
	SiteWorkSetID          *string `json:"siteWorkSetId"`
	SiteWorkSetName        *string `json:"siteWorkSetName"`
	SiteAuthorID           *string `json:"siteAuthorId"`
	SiteWorkSetDescription *string `json:"siteWorkSetDescription"`
	SiteUploadTime         *int64  `json:"siteUploadTime"`
	SiteUpdateTime         *int64  `json:"siteUpdateTime"`
	NickName               *string `json:"nickName"`
	LastView               *int64  `json:"lastView"`
	CreateTime             int64   `json:"createTime"`
	UpdateTime             int64   `json:"updateTime"`
}

// WorkSetWithWorksResultDTO 作品集及其作品信息
type WorkSetWithWorksResultDTO struct {
	WorkSet *WorkSetDTO    `json:"workSet"`
	Works   []*WorkFullDTO `json:"works,omitempty"`
}

// WorkSetWithCoverDTO 作品集及其封面作品信息
type WorkSetWithCoverDTO struct {
	WorkSet       *WorkSetDTO  `json:"workSet"`
	CoverWork     *WorkDTO     `json:"coverWork,omitempty"`
	CoverResource *ResourceDTO `json:"coverResource,omitempty"`
}
