package dto

// TaskDTO 任务数据传输对象
type TaskDTO struct {
	ID                int64    `json:"id"`
	HasChild          *bool    `json:"hasChild"`
	Pid               *int64   `json:"pid"`
	TaskName          *string  `json:"taskName"`
	SiteID            *int64   `json:"siteId"`
	SiteWorkID        *string  `json:"siteWorkId"`
	URL               *string  `json:"url"`
	Status            int      `json:"status"`
	PendingResourceID *int64   `json:"pendingResourceId"`
	Continuable       *bool    `json:"continuable"`
	PluginPublicID    *string  `json:"pluginPublicId"`
	PluginExtensionID *string  `json:"pluginExtensionId"`
	PluginData        *string  `json:"pluginData"`
	ErrorMessage      *string  `json:"errorMessage"`
	InvolvedRoles     []string `json:"involvedRoles"` // 任务涉及的 store_type 集合(创建期声明,universe);nil=未确定;用于前端按任务自选展示
	ResourceType      string   `json:"resourceType"`  // 任务产生的 resource 的资源类型(预定义值);空=未声明
	CreateTime        int64    `json:"createTime"`
	UpdateTime        int64    `json:"updateTime"`
}
