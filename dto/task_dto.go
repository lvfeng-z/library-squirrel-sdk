package dto

// TaskDTO 任务数据传输对象
type TaskDTO struct {
	ID                   int64   `json:"id"`
	HasChild             *bool   `json:"hasChild"`
	Pid                  *int64  `json:"pid"`
	TaskName             *string `json:"taskName"`
	SiteID               *int64  `json:"siteId"`
	SiteWorkID           *string `json:"siteWorkId"`
	URL                  *string `json:"url"`
	Status               int     `json:"status"`
	PendingResourceID    *int64  `json:"pendingResourceId"`
	Continuable          *bool   `json:"continuable"`
	PluginPublicID       *string `json:"pluginPublicId"`
	PluginExtensionID *string `json:"pluginExtensionId"`
	PluginData           *string `json:"pluginData"`
	ErrorMessage         *string `json:"errorMessage"`
	CreateTime           int64   `json:"createTime"`
	UpdateTime           int64   `json:"updateTime"`
}

// TaskProgressDTO 任务进度DTO
type TaskProgressDTO struct {
	Task     *TaskDTO `json:"task,omitempty"`
	Total    *int64   `json:"total,omitempty"`
	Finished *int64   `json:"finished,omitempty"`
	SiteName *string  `json:"siteName,omitempty"`
	Schedule *int     `json:"schedule,omitempty"`
}

// TaskProgressTreeDTO 任务进度树DTO
type TaskProgressTreeDTO struct {
	TaskProgress *TaskProgressDTO       `json:"taskProgress,omitempty"`
	Children     []*TaskProgressTreeDTO `json:"children,omitempty"`
	HasChildren  *bool                  `json:"hasChildren,omitempty"`
	IsLeaf       *bool                  `json:"isLeaf,omitempty"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Pid                  int64  `json:"pid"`
	TaskName             string `json:"taskName"`
	SiteID               int    `json:"siteId"`
	SiteWorkID           string `json:"siteWorkId"`
	URL                  string `json:"url"`
	HasChild             bool   `json:"hasChild"`
	PluginPublicID       string `json:"pluginPublicId"`
	PluginExtensionID string `json:"pluginExtensionId"`
	PluginData           string `json:"pluginData"`
}

// TreeDataPageDTO 任务树数据分页DTO
type TreeDataPageDTO struct {
	TreeID   int64                  `json:"treeId"`
	TreeName string                 `json:"treeName"`
	Total    int64                  `json:"total"`
	Tasks    []*TaskProgressTreeDTO `json:"tasks"`
}
