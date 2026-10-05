package dto

import "github.com/lvfeng-z/library-squirrel-sdk/gen"

// PreferenceValue 偏好值信封（别名 gen.PreferenceValue，proto 单源）。
// SchemaVersion 标记值结构版本（初值 1，规则类值演进时插件自管升版——高于自身支持
// 版本的读取应 fail-fast，防静默数据损坏）；Title/Description 由插件写入时随值携带
//（决策标题在问答发生时才确定，供宿主记忆管理页展示）；Data 为插件自定义负载
//（JSON 文本，插件自行解释与结构化，宿主只存不解释）。
type PreferenceValue = gen.PreferenceValue
