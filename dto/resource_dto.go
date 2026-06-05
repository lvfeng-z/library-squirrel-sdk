package dto

// ResourceDTO 资源数据传输对象
type ResourceDTO struct {
	ID               int64   `json:"id"`
	WorkID           int64   `json:"workId"`
	TaskID           int64   `json:"taskId"`
	Enabled          bool    `json:"enabled"`
	SuggestName      *string `json:"suggestName"`
	ResourceComplete int     `json:"resourceComplete"`
	CreateTime       int64   `json:"createTime"`
	UpdateTime       int64   `json:"updateTime"`
}
