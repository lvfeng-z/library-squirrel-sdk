package dto

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

// LocalTagWithBaseTagDTO 本地标签及其基础标签数据传输对象
type LocalTagWithBaseTagDTO struct {
	LocalTag *LocalTagDTO `json:"localTag,omitempty"`
	BaseTag  *LocalTagDTO `json:"baseTag,omitempty"`
}
