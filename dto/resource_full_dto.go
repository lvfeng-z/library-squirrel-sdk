package dto

// ResourceFullDTO 资源完整 DTO（包含作品资源和封面的 PersistentStore 信息）
type ResourceFullDTO struct {
	ID               int64              `json:"id"`
	WorkID           int64              `json:"workId"`
	TaskID           int64              `json:"taskId"`
	Enabled          bool               `json:"enabled"`
	SuggestName      *string            `json:"suggestName"`
	ResourceComplete int                `json:"resourceComplete"`
	WorkStoreID      *int64             `json:"workStoreId"`
	ThumbnailStoreID *int64             `json:"thumbnailStoreId"`
	WorkStore        *PersistentStoreDTO `json:"workStore,omitempty"`       // 作品资源文件
	ThumbnailStore   *PersistentStoreDTO `json:"thumbnailStore,omitempty"`  // 封面/缩略图
	CreateTime       int64              `json:"createTime"`
	UpdateTime       int64              `json:"updateTime"`
}
