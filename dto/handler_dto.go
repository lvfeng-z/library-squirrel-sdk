package dto

// TaskResParam 任务和资源参数
type TaskResParam struct {
	Task            *TaskDTO `json:"task"`
	ResourceID      int64    `json:"resourceId"`
	ResourcePath    string   `json:"resourcePath"`
	DownloadedBytes int64    `json:"downloadedBytes"`
}

// TaskCreateResponse 任务创建响应
type TaskCreateResponse struct {
	PluginTaskID string                     `json:"pluginTaskId"`
	TaskName     string                     `json:"taskName"`
	SiteWorkID   string                     `json:"siteWorkId"`
	URL          string                     `json:"url"`
	PluginData   string                     `json:"pluginData"`
	SiteName     string                     `json:"siteName"`
	Children     []*TaskCreateChildResponse `json:"children"`
}

// TaskCreateChildResponse 子任务创建响应
type TaskCreateChildResponse struct {
	TaskName   string `json:"taskName"`
	SiteWorkID string `json:"siteWorkId"`
	URL        string `json:"url"`
	PluginData string `json:"pluginData"`
	SiteName   string `json:"siteName"`
}

// WorkResponse 作品响应
type WorkResponse struct {
	Work         *WorkDTO             `json:"work"`
	Site         *SiteDTO             `json:"site"`
	LocalAuthors []*LocalAuthorDTO    `json:"localAuthors"`
	LocalTags    []*LocalTagDTO       `json:"localTags"`
	SiteAuthors  []*TaskSiteAuthorDTO `json:"siteAuthors"`
	SiteTags     []*TaskSiteTagDTO    `json:"siteTags"`
	WorkSets     []*TaskWorkSetDTO    `json:"workSets"`
	Resource     *TaskResourceDTO     `json:"resource"`
}

// TaskSiteAuthorDTO 任务处理器站点作者DTO
type TaskSiteAuthorDTO struct {
	SiteAuthorID    string `json:"siteAuthorId"`
	AuthorName      string `json:"authorName"`
	FixedAuthorName string `json:"fixedAuthorName"`
	Introduce       string `json:"introduce"`
	Homepage        string `json:"homepage"`
}

// TaskSiteTagDTO 任务处理器站点标签DTO
type TaskSiteTagDTO struct {
	SiteTagID   string `json:"siteTagId"`
	TagName     string `json:"tagName"`
	Description string `json:"description"`
}

// TaskWorkSetDTO 任务处理器作品集DTO
type TaskWorkSetDTO struct {
	SiteWorkSetID string `json:"siteWorkSetId"`
	WorkSetName   string `json:"workSetName"`
}

// TaskResourceDTO 任务处理器资源DTO
type TaskResourceDTO struct {
	ResourceID   int64  `json:"resourceId"`
	URL          string `json:"url"`
	Type         string `json:"type"`
	Format       string `json:"format"`
	LocalPath    string `json:"localPath"`
	RemotePath   string `json:"remotePath"`
	Size         int64  `json:"size"`
	Completeness int    `json:"completeness"`
	SuggestName  string `json:"suggestName"`
	Continuable  *bool  `json:"continuable"` // 插件声明当前资源是否支持在已有文件上续传
}
