package dto

// ResourceStoreDTO Resource 关联的单个 typed store(包装 PersistentStoreDTO + store_type/generation)
type ResourceStoreDTO struct {
	StoreType  string              `json:"storeType"`          // main | thumbnail | videoTrack | audioTrack | merged
	Generation string              `json:"generation"`         // downloaded | derived
	Store      *PersistentStoreDTO `json:"store,omitempty"`    // 对应的 PersistentStore 信息
}

// ResourceFullDTO 资源完整 DTO
// Stores 为 resource_store 关联表全部 store(多轨模型,主数据源);
// WorkStore/ThumbnailStore 为从 Stores 按 storeType 派生的便捷访问器(供前端取主文件/缩略图路径,减少改动面)
type ResourceFullDTO struct {
	ID               int64                `json:"id"`
	WorkID           int64                `json:"workId"`
	TaskID           int64                `json:"taskId"`
	Enabled          bool                 `json:"enabled"`
	SuggestName      *string              `json:"suggestName"`
	ResourceComplete int                  `json:"resourceComplete"`
	Stores           []ResourceStoreDTO   `json:"stores,omitempty"`          // 全部 store(resource_store 关联)
	WorkStore        *PersistentStoreDTO   `json:"workStore,omitempty"`       // 便捷访问器:storeType=main
	ThumbnailStore   *PersistentStoreDTO   `json:"thumbnailStore,omitempty"`  // 便捷访问器:storeType=thumbnail
	CreateTime       int64                `json:"createTime"`
	UpdateTime       int64                `json:"updateTime"`
}
