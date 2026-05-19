package pluginsdk

// TaskResParam 任务和资源参数
type TaskResParam struct {
	Task         *Task   `json:"task"`
	ResourceID   int64   `json:"resourceId"`
	ResourcePath string  `json:"resourcePath"`
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
	Work         *Work                `json:"work"`
	Site         *SiteDTO             `json:"site"`
	LocalAuthors []*LocalAuthorDTO    `json:"localAuthors"`
	LocalTags    []*LocalTagDTO       `json:"localTags"`
	SiteAuthors  []*TaskSiteAuthorDTO `json:"siteAuthors"`
	SiteTags     []*TaskSiteTagDTO    `json:"siteTags"`
	WorkSets     []*TaskWorkSetDTO    `json:"workSets"`
	Resource     *TaskResourceDTO     `json:"resource"`
}

// SiteDTO 站点信息
type SiteDTO struct {
	ID              int64   `json:"id"`
	SiteName        *string `json:"siteName"`
	SiteDescription *string `json:"siteDescription"`
	Homepage        *string `json:"homepage"`
	CreateTime      int64   `json:"createTime"`
	UpdateTime      int64   `json:"updateTime"`
}

// LocalAuthorDTO 本地作者
type LocalAuthorDTO struct {
	ID         int64   `json:"id"`
	AuthorName *string `json:"authorName"`
	Introduce  *string `json:"introduce"`
	LastUse    *int64  `json:"lastUse"`
	CreateTime int64   `json:"createTime"`
	UpdateTime int64   `json:"updateTime"`
}

// LocalTagDTO 本地标签
type LocalTagDTO struct {
	ID             int64   `json:"id"`
	LocalTagName   *string `json:"localTagName"`
	BaseLocalTagID *int64  `json:"baseLocalTagId"`
	Description    *string `json:"description"`
	LastUse        *int64  `json:"lastUse"`
	CreateTime     int64   `json:"createTime"`
	UpdateTime     int64   `json:"updateTime"`
}

// TaskSiteAuthorDTO 任务处理器站点作者DTO
type TaskSiteAuthorDTO struct {
	SiteAuthorID string `json:"siteAuthorId"`
	AuthorName   string `json:"authorName"`
	URL          string `json:"url"`
}

// TaskSiteTagDTO 任务处理器站点标签DTO
type TaskSiteTagDTO struct {
	SiteTagID   string `json:"siteTagId"`
	TagName     string `json:"tagName"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// TaskWorkSetDTO 任务处理器作品集DTO
type TaskWorkSetDTO struct {
	SiteWorkSetID string `json:"siteWorkSetId"`
	WorkSetName string `json:"workSetName"`
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
}
