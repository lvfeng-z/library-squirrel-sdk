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

// TaskResourceDTO 任务处理器资源 DTO（精简后）
type TaskResourceDTO struct {
	Size        int64  `json:"size"`        // 远程文件大小
	Type        string `json:"type"`        // 资源类型
	Format      string `json:"format"`      // 文件格式/扩展名（如 "jpg"、"mp4"）
	SuggestName string `json:"suggestName"` // 插件建议文件名
	Continuable *bool  `json:"continuable"` // 是否支持续传
}
