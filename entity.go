package pluginsdk

// Task 任务（等价类型，主程序负责转换）
type Task struct {
	ID                    int64   `json:"id"`
	CreateTime            int64   `json:"createTime"`
	UpdateTime            int64   `json:"updateTime"`
	HasChild              *bool   `json:"hasChild"`
	Pid                   *int64  `json:"pid"`
	TaskName              *string `json:"taskName"`
	SiteID                *int64  `json:"siteId"`
	SiteWorkID            *string `json:"siteWorkId"`
	URL                   *string `json:"url"`
	Status                int     `json:"status"`
	PendingResourceID     *int64  `json:"pendingResourceId"`
	Continuable           *bool   `json:"continuable"`
	PluginPublicID        *string `json:"pluginPublicId"`
	PluginContributionID  *string `json:"pluginContributionId"`
	PluginData            *string `json:"pluginData"`
	ErrorMessage          *string `json:"errorMessage"`
}

// Work 作品（等价类型，主程序负责转换）
type Work struct {
	ID                   int64   `json:"id"`
	CreateTime           int64   `json:"createTime"`
	UpdateTime           int64   `json:"updateTime"`
	SiteID               *int64  `json:"siteId"`
	SiteWorkID           *string `json:"siteWorkId"`
	SiteWorkName         *string `json:"siteWorkName"`
	SiteAuthorID         *string `json:"siteAuthorId"`
	SiteWorkDescription  *string `json:"siteWorkDescription"`
	SiteUploadTime       *int64  `json:"siteUploadTime"`
	SiteUpdateTime       *int64  `json:"siteUpdateTime"`
	NickName             *string `json:"nickName"`
	LocalAuthorID        *int64  `json:"localAuthorId"`
	LastView             *int64  `json:"lastView"`
}

// WorkSet 作品集（等价类型，主程序负责转换）
type WorkSet struct {
	ID                     int64   `json:"id"`
	CreateTime             int64   `json:"createTime"`
	UpdateTime             int64   `json:"updateTime"`
	SiteID                 *int64  `json:"siteId"`
	SiteWorkSetID          *string `json:"siteWorkSetId"`
	SiteWorkSetName        *string `json:"siteWorkSetName"`
	SiteAuthorID           *string `json:"siteAuthorId"`
	SiteWorkSetDescription *string `json:"siteWorkSetDescription"`
	SiteUploadTime         *int64  `json:"siteUploadTime"`
	SiteUpdateTime         *int64  `json:"siteUpdateTime"`
	NickName               *string `json:"nickName"`
	LastView               *int64  `json:"lastView"`
}

// Site 站点（等价类型，主程序负责转换）
type Site struct {
	ID              int64   `json:"id"`
	CreateTime      int64   `json:"createTime"`
	UpdateTime      int64   `json:"updateTime"`
	SiteName        *string `json:"siteName"`
	SiteDescription *string `json:"siteDescription"`
	Homepage        *string `json:"homepage"`
}
