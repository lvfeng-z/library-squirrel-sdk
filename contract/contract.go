// Package contract 定义主程序与插件共享的契约枚举常量（store_type / resource_type / generation）。
//
// 本包是这些字符串枚举值的单一真相源：主程序 entity 与 SDK dto 经 const 别名 re-export，
// 改一处即处处同步（编译期绑定）。此处为内置预定义值；ResourceType 除内置集外，插件可经
// 主程序 ResourceTypeRegistry 运行时注册自定义类型（不在此 const 集合，详见主程序 entity）。
package contract

// StoreType 标识 resource_store 的业务角色（内置 7 值，封闭枚举）。
// 资源类型规约体系据此组合表达结构；插件自定义 ResourceType 本次仅复用内置角色（自定义角色延后）。
const (
	StoreTypeImage      = "image"      // 图片(image 资源主体;article 内嵌图多例)
	StoreTypeDocument   = "document"   // 文档文件(article 正文 .md;document 原文件 .pdf/.docx)
	StoreTypeThumbnail  = "thumbnail"  // 缩略图/封面
	StoreTypeVideoTrack = "videoTrack" // 视频轨(分离流视频原料)
	StoreTypeAudioTrack = "audioTrack" // 音频轨(分离流音频原料)
	StoreTypeVideoMain  = "videoMain"  // 视频可播放主体(封装原文件 downloaded 或合并产物 derived)
	StoreTypeAudioMain  = "audioMain"  // 音频可播放主体(独立音频资源主体)
)

// Generation 标识 store 实例的生成方式（数据获取时序，决定执行与续传语义）。
// store 实例属性（每行 resource_store 一个值），由产出方决定，不从 store_type 推断——
// 同一 store_type 可跨 generation（如 videoMain:封装原文件 downloaded、合并产物 derived）。
const (
	GenerationDownloaded = "downloaded" // 流式下载获得,支持断点续传
	GenerationDerived    = "derived"    // 一次性派生产物,不可续传
)

// ResourceType 预定义资源类型（内置 6 值）。
// 插件创建资源时必须声明其一（或经主程序 ResourceTypeRegistry 注册的自定义类型）；
// 主程序不推断、不兜底，空/未知值在写入路径抛错。
const (
	ResourceTypeImage    = "image"    // 图片资源(单图/多图每子资源)
	ResourceTypeVideo    = "video"    // 视频资源(可播放主体 videoMain 必含;分离流另含 videoTrack+audioTrack)
	ResourceTypeArticle  = "article"  // 图文紧密结合文档(正文 markdown + 内嵌图)
	ResourceTypeDocument = "document" // 现成文档原文件(pdf/docx/...)
	ResourceTypeAudio    = "audio"    // 音频资源(可播放主体 audioMain 必含)
	ResourceTypeUnknown  = "unknown"  // 合法显式值:插件确实无法分类时声明
)
