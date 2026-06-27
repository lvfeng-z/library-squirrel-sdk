package dto

import (
	"io"
)

// StoreSpec 的 role(generation)契约常量,供插件与主程序共用,避免魔法字符串漂移。
// 与主程序 entity 包的 StoreType*/Generation* 字面量一致。
const (
	StoreRoleMain       = "main"       // 主资源
	StoreRoleThumbnail  = "thumbnail"  // 缩略图/封面
	StoreRoleVideoTrack = "videoTrack" // 视频轨
	StoreRoleAudioTrack = "audioTrack" // 音频轨
	StoreRoleMerged     = "merged"     // 合并产物

	GenerationDownloaded = "downloaded" // 流式下载,支持断点续传
	GenerationDerived    = "derived"    // 一次性派生,不可续传
)

// TaskResParam 任务和资源参数(Pause/Stop 共用)
type TaskResParam struct {
	Task            *TaskDTO `json:"task"`
	ResourceID      int64    `json:"resourceId"`
	ResourcePath    string   `json:"resourcePath"`
	DownloadedBytes int64    `json:"downloadedBytes"`
}

// TaskResumeParam 续传参数(每条 downloaded 轨独立偏移)
type TaskResumeParam struct {
	Task          *TaskDTO         `json:"task"`
	StreamOffsets map[string]int64 `json:"streamOffsets"` // role → 该轨已写入字节数(仅未完成 downloaded 轨出现;derived 轨不出现,未完成即整轨重产)
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

// WorkResponse 作品响应(仅承载作品级信息;资源细节由 StoreSpec 承载)
type WorkResponse struct {
	Work         *WorkDTO             `json:"work"`
	Site         *SiteDTO             `json:"site"`
	LocalAuthors []*LocalAuthorDTO    `json:"localAuthors"`
	LocalTags    []*LocalTagDTO       `json:"localTags"`
	SiteAuthors  []*TaskSiteAuthorDTO `json:"siteAuthors"`
	SiteTags     []*TaskSiteTagDTO    `json:"siteTags"`
	WorkSets     []*TaskWorkSetDTO    `json:"workSets"`
}

// StoreSpec 单条资源产出声明(对应一个 store)
type StoreSpec struct {
	Role        string        `json:"role"`                                                                  // store_type: main | thumbnail | videoTrack | audioTrack | ...
	Generation  string        `json:"generation"`                                                            // downloaded(流式可续传) | derived(一次性派生)
	ReadCloser  io.ReadCloser `json:"-"`                                                                     // 资源数据流(downloaded=流式 reader;derived=一次性 payload 包装的 reader),调用方负责 Close
	Format      string        `json:"format"`                                                                // 文件扩展名
	Size        int64         `json:"size"`                                                                  // 远程大小;-1 未知
	SuggestName string        `json:"suggestName,omitempty"`                                                 // 插件建议文件名
	Continuable *bool         `json:"continuable,omitempty"`                                                 // 是否支持续传(derived 恒为 false)
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
