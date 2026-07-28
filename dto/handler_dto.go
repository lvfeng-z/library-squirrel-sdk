package dto

import (
	"fmt"
	"io"
)

// StoreSpec 的 role(generation)契约常量,供插件与主程序共用,避免魔法字符串漂移。
// 与主程序 entity 包的 StoreType*/Generation* 字面量一致。
const (
	StoreRoleImage      = "image"      // 图片(image 资源主体;article 内嵌图多例)
	StoreRoleDocument   = "document"   // 文档文件(article 正文 .md;document 原文件 .pdf/.docx)
	StoreRoleThumbnail  = "thumbnail"  // 缩略图/封面
	StoreRoleVideoTrack = "videoTrack" // 分离流视频原料(无音频)
	StoreRoleAudioTrack = "audioTrack" // 分离流音频原料
	StoreRoleVideoMain  = "videoMain"  // 视频可播放主体(封装原文件 downloaded 或合并产物 derived)

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

// TaskResumeParam 续传参数(每条 downloaded store 独立偏移,按 role+store_seq 身份化)
type TaskResumeParam struct {
	Task          *TaskDTO             `json:"task"`
	StreamOffsets []*StoreResumeOffset `json:"streamOffsets"` // 未完成 downloaded store 的续传偏移;同 role 多 store 各自独立(N-同 role 多 store 支持)
}

// StoreResumeOffset 单条 downloaded store 的续传偏移(身份化:role+store_seq 唯一定位一个 store)
type StoreResumeOffset struct {
	Role     string `json:"role"`     // store_type(image/document/videoTrack/audioTrack/...)
	StoreSeq int32  `json:"storeSeq"` // 同 role 内的 store 序号(单例为 0;N-同 role 多 store 各自 0..N-1)
	Offset   int64  `json:"offset"`   // 该 store 已写入字节数(磁盘 stat 为权威)
}

// OffsetForRole 按 role 查询续传偏移(单例场景便捷方法,取首个命中)。
// N-同 role 多 store 须插件自行遍历 StreamOffsets 按 store_seq 精确匹配。
func (p *TaskResumeParam) OffsetForRole(role string) (offset int64, found bool) {
	for _, o := range p.StreamOffsets {
		if o != nil && o.Role == role {
			return o.Offset, true
		}
	}
	return 0, false
}

// String 让 []*StoreResumeOffset 的 %v 日志可读(否则打印指针地址)
func (o *StoreResumeOffset) String() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{%s/%d:%d}", o.Role, o.StoreSeq, o.Offset)
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
	Role              string        `json:"role"`                        // store_type: image | document | thumbnail | videoTrack | audioTrack | videoMain
	Generation        string        `json:"generation"`                  // downloaded(流式可续传) | derived(一次性派生)
	ReadCloser        io.ReadCloser `json:"-"`                           // 资源数据流(downloaded=流式 reader;derived=一次性 payload 包装的 reader),调用方负责 Close
	Format            string        `json:"format"`                      // 文件扩展名
	Size              int64         `json:"size"`                        // 完整资源大小(非 Range 续传剩余字节,206 需据 Content-Range 还原);-1 未知
	SuggestName       string        `json:"suggestName,omitempty"`       // 插件建议文件名
	Description       string        `json:"description,omitempty"`       // 资源描述(多 store 命名可选拼段;空则省略描述段)
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
