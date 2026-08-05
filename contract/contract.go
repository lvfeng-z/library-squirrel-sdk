// Package contract 定义主程序与插件共享的契约枚举常量（store_type / resource_type / generation）。
//
// 本包是这些字符串枚举值的单一真相源：主程序 entity 与 SDK dto 经 const 别名 re-export，
// 改一处即处处同步（编译期绑定）。封闭枚举——新增值须同步 entity 的 ResourceTypeRegistry/
// validStoreTypes 与 dto 的消费方。
package contract

// StoreType 标识 resource_store 的业务角色（封闭枚举，6 值）；资源类型规约体系据此组合表达结构。
const (
	StoreTypeImage      = "image"      // 图片(image 资源主体;article 内嵌图多例)
	StoreTypeDocument   = "document"   // 文档文件(article 正文 .md;document 原文件 .pdf/.docx)
	StoreTypeThumbnail  = "thumbnail"  // 缩略图/封面
	StoreTypeVideoTrack = "videoTrack" // 视频轨(分离流视频原料)
	StoreTypeAudioTrack = "audioTrack" // 音频轨(分离流音频原料)
	StoreTypeVideoMain  = "videoMain"  // 视频可播放主体(封装原文件 downloaded 或合并产物 derived)
)

// Generation 标识 store 实例的生成方式（数据获取时序，决定执行与续传语义）。
// store 实例属性（每行 resource_store 一个值），由产出方决定，不从 store_type 推断——
// 同一 store_type 可跨 generation（如 videoMain:封装原文件 downloaded、合并产物 derived）。
const (
	GenerationDownloaded = "downloaded" // 流式下载获得,支持断点续传
	GenerationDerived    = "derived"    // 一次性派生产物,不可续传
)

// ResourceType 预定义资源类型（封闭枚举，5 值）。
// 插件创建资源时必须声明其一；主程序不推断、不兜底，空/未知值在写入路径抛错。
const (
	ResourceTypeImage    = "image"    // 图片资源(单图/多图每子资源)
	ResourceTypeVideo    = "video"    // 视频资源(可播放主体 videoMain 必含;分离流另含 videoTrack+audioTrack)
	ResourceTypeArticle  = "article"  // 图文紧密结合文档(正文 markdown + 内嵌图)
	ResourceTypeDocument = "document" // 现成文档原文件(pdf/docx/...)
	ResourceTypeUnknown  = "unknown"  // 合法显式值:插件确实无法分类时声明
)
