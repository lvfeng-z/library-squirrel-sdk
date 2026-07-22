package dto

import (
	"io"
)

// StoreSpec 的 role(generation)契约常量,供插件与主程序共用,避免魔法字符串漂移。
// 与主程序 entity 包的 StoreType*/Generation* 字面量一致。
const (
	StoreRoleImage      = "image"      // 图片(image 资源主体;article 内嵌图多例)
	StoreRoleDocument   = "document"   // 文档文件(article 正文 .md;document 原文件 .pdf/.docx)
	StoreRoleThumbnail  = "thumbnail"  // 缩略图/封面
	StoreRoleVideoTrack = "videoTrack" // 视频轨
	StoreRoleAudioTrack = "audioTrack" // 音频轨
	StoreRoleMerged     = "merged"     // 合并产物

	GenerationDownloaded = "downloaded" // 流式下载,支持断点续传
	GenerationDerived    = "derived"    // 一次性派生,不可续传
)

// ResourceType 预定义资源类型常量(与主程序 entity 包的 ResourceType* 字面量一致)
const (
	ResourceTypeImage    = "image"    // 图片资源
	ResourceTypeVideo    = "video"    // 视频资源
	ResourceTypeArticle  = "article"  // 图文紧密结合文档(正文+内嵌图)
	ResourceTypeDocument = "document" // 现成文档原文件
	ResourceTypeUnknown  = "unknown"  // 插件确实无法分类时声明
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
	PluginTaskID  string                     `json:"pluginTaskId"`
	TaskName      string                     `json:"taskName"`
	SiteWorkID    string                     `json:"siteWorkId"`
	URL           string                     `json:"url"`
	PluginData    string                     `json:"pluginData"`
	SiteName      string                     `json:"siteName"`
	InvolvedRoles []string                   `json:"involvedRoles"` // 任务涉及的 store_type 集合(创建期声明,universe);空/nil=未确定,执行期插件下全量
	ResourceType  string                     `json:"resourceType"`  // 任务产生的 resource 的资源类型(预定义 image/video/article/document/unknown);空=未声明;有 children 时由各 child 声明
	Children      []*TaskCreateChildResponse `json:"children"`
}

// TaskCreateChildResponse 子任务创建响应
type TaskCreateChildResponse struct {
	TaskName      string   `json:"taskName"`
	SiteWorkID    string   `json:"siteWorkId"`
	URL           string   `json:"url"`
	PluginData    string   `json:"pluginData"`
	SiteName      string   `json:"siteName"`
	InvolvedRoles []string `json:"involvedRoles"` // 子任务涉及的 store_type 集合(创建期声明,universe);空/nil=未确定,执行期插件下全量
	ResourceType  string   `json:"resourceType"`  // 子任务产生的 resource 的资源类型(预定义值);空=未声明
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
	Role              string        `json:"role"`                        // store_type: image | document | thumbnail | videoTrack | audioTrack | merged
	Generation        string        `json:"generation"`                  // downloaded(流式可续传) | derived(一次性派生)
	ReadCloser        io.ReadCloser `json:"-"`                           // 资源数据流(downloaded=流式 reader;derived=一次性 payload 包装的 reader),调用方负责 Close
	Format            string        `json:"format"`                      // 文件扩展名
	Size              int64         `json:"size"`                        // 完整资源大小(非 Range 续传剩余字节,206 需据 Content-Range 还原);-1 未知
	SuggestName       string        `json:"suggestName,omitempty"`       // 插件建议文件名
	Continuable       *bool         `json:"continuable,omitempty"`       // 是否支持续传(derived 恒为 false)
	ResumeWriteOffset *int64        `json:"resumeWriteOffset,omitempty"` // 续传写入偏移(仅 Resume 返回的 spec);nil=信任主程序 stat 的 offset,非 nil=插件指定确切位置
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
