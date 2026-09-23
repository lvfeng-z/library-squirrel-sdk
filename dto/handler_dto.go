package dto

import (
	"fmt"
	"io"

	"github.com/lvfeng-z/library-squirrel-sdk/contract"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// StoreSpec 的 role(generation)与 ResourceType 契约常量，别名 SDK contract 包（单一真相源）。
// StoreRole* 保留旧名兼容现有插件引用（值 = contract.StoreType*）。
const (
	StoreRoleImage      = contract.StoreTypeImage
	StoreRoleDocument   = contract.StoreTypeDocument
	StoreRoleThumbnail  = contract.StoreTypeThumbnail
	StoreRoleVideoTrack = contract.StoreTypeVideoTrack
	StoreRoleAudioTrack = contract.StoreTypeAudioTrack
	StoreRoleVideoMain  = contract.StoreTypeVideoMain
	StoreRoleAudioMain  = contract.StoreTypeAudioMain

	GenerationDownloaded = contract.GenerationDownloaded
	GenerationDerived    = contract.GenerationDerived

	ResourceTypeImage    = contract.ResourceTypeImage
	ResourceTypeVideo    = contract.ResourceTypeVideo
	ResourceTypeArticle  = contract.ResourceTypeArticle
	ResourceTypeDocument = contract.ResourceTypeDocument
	ResourceTypeAudio    = contract.ResourceTypeAudio
	ResourceTypeUnknown  = contract.ResourceTypeUnknown
)

// TaskResParam 任务和资源参数(Pause/Stop 共用)（别名 gen.TaskResParam，proto 单源）
type TaskResParam = gen.TaskResParam

// TaskResumeParam 续传参数(每条 downloaded store 独立偏移,按 role+store_seq 身份化)
// 保留为 struct：含 OffsetForRole 便捷方法,别名后无法在 dto 包为 gen 类型定义方法。
type TaskResumeParam struct {
	Task *TaskDTO `json:"task"`
	// 主程序磁盘暂存中全部未提交 downloaded 轨的已落盘偏移(可能已达该轨完整大小——多轨并发
	// 下小轨先写满、等待其余轨道一起提交是常态)。插件是站点与主程序间的兼容层,单轨完成状态
	// 的确认归插件职责:对每条偏移与来源侧比对——已写满(来源现值==偏移)→返回 Size=偏移、立即
	// EOF 流的 spec 或空过(主程序经 Start 重产);来源现值与偏移不符(内容变更)→返回以来源现值
	// 为 Size 的完整流,主程序按声明大小截断整轨重下。详见 proto/plugin.proto 同名消息注释
	StreamOffsets []*StoreResumeOffset `json:"streamOffsets"`
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

// StoreResumeOffset 单条 downloaded store 的续传偏移(身份化:role+store_seq 唯一定位一个 store)
// 保留为 struct：含 String 方法,且 gen.StoreResumeOffset 已有同名 proto 方法(冲突,无法别名)。
type StoreResumeOffset struct {
	Role     string `json:"role"`     // store_type(image/document/videoTrack/audioTrack/...)
	StoreSeq int32  `json:"storeSeq"` // 同 role 内的 store 序号(单例为 0;N-同 role 多 store 各自 0..N-1)
	Offset   int64  `json:"offset"`   // 该 store 已写入字节数(磁盘 stat 为权威)
}

// TaskCreateResponse 任务创建响应（别名 gen.TaskCreateResponse，proto 单源）
type TaskCreateResponse = gen.TaskCreateResponse

// TaskCreateChildResponse 子任务创建响应（别名 gen.TaskCreateChildResponse，proto 单源）
type TaskCreateChildResponse = gen.TaskCreateChildResponse

// WorkResponse 作品响应(仅承载作品级信息;资源细节由 StoreSpec 承载)（别名 gen.WorkResponse，proto 单源）
type WorkResponse = gen.WorkResponse

// StoreSpec 单条资源产出声明(对应一个 store)
// 手写保留：含 io.ReadCloser(非可序列化),非 proto 类型;元数据经 StoreSpecMeta 跨 gRPC、reader 数据走 stream data 块。
type StoreSpec struct {
	Role              string        `json:"role"`                        // store_type: image | document | thumbnail | videoTrack | audioTrack | videoMain | audioMain
	Generation        string        `json:"generation"`                  // downloaded(流式可续传) | derived(一次性派生)
	ReadCloser        io.ReadCloser `json:"-"`                           // 资源数据流(downloaded=流式 reader;derived=一次性 payload 包装的 reader),调用方负责 Close
	Format            string        `json:"format"`                      // 文件扩展名
	Size              int64         `json:"size"`                        // 完整资源大小(非 Range 续传剩余字节,206 需据 Content-Range 还原);-1 未知
	SuggestName       string        `json:"suggestName,omitempty"`       // 插件建议文件名
	Description       string        `json:"description,omitempty"`       // 资源描述(多 store 命名可选拼段;空则省略描述段)
	Continuable       *bool         `json:"continuable,omitempty"`       // 是否支持续传(derived 恒为 false)
	ResumeWriteOffset *int64        `json:"resumeWriteOffset,omitempty"` // 续传写入偏移(仅 Resume 返回的 spec);nil=信任主程序 stat 的 offset,非 nil=插件指定确切位置
	ExpectedSha256    *string       `json:"expectedSha256,omitempty"`    // 来源侧声明的期望 SHA256(下载内容完整性校验);nil=来源未声明,主程序跳过校验
}

// TaskSiteAuthorDTO 作品拉取站点作者DTO（别名 gen.TaskSiteAuthorDTO，proto 单源）
type TaskSiteAuthorDTO = gen.TaskSiteAuthorDTO

// TaskSiteTagDTO 作品拉取站点标签DTO（别名 gen.TaskSiteTagDTO，proto 单源）
type TaskSiteTagDTO = gen.TaskSiteTagDTO

// TaskWorkSetDTO 作品拉取作品集DTO（别名 gen.TaskWorkSetDTO，proto 单源）
type TaskWorkSetDTO = gen.TaskWorkSetDTO
