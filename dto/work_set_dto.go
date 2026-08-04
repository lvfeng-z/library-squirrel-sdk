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
	Description            *string `json:"description"`
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
	WorkSet       *WorkSetDTO       `json:"workSet"`
	CoverWork     *WorkDTO          `json:"coverWork,omitempty"`
	CoverResource *ResourceFullDTO  `json:"coverResource,omitempty"`
}

// WorkOrderEntry 作品在作品集内的原站排序条目（插件返回集内全序，主程序据此写 site_sort_order）
type WorkOrderEntry struct {
	SiteWorkID string `json:"siteWorkId"`
	SortOrder  int64  `json:"sortOrder"`
}

// WorkOrderQuerier 可选能力接口：插件实现此接口以提供作品集内作品的原站顺序
// 未实现此接口的插件，主程序查询原站序时得到空响应（site_sort_order 保持空，仅本地序生效）
type WorkOrderQuerier interface {
	// QueryWorkSetOrder 返回作品集内作品的原站全序；空切片=插件不掌握
	QueryWorkSetOrder(siteId int64, siteWorkSetId string) ([]*WorkOrderEntry, error)
}
