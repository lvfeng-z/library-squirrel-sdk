package dto

// ResourceStoreDTO Resource 关联的单个 typed store(包装 PersistentStoreDTO + store_type/generation)
type ResourceStoreDTO struct {
	StoreType  string              `json:"storeType"`       // image | document | thumbnail | videoTrack | audioTrack | videoMain
	Generation string              `json:"generation"`      // downloaded | derived
	Store      *PersistentStoreDTO `json:"store,omitempty"` // 对应的 PersistentStore 信息
}

// ResourceFullDTO 资源完整 DTO
// Stores 为 resource_store 关联表全部 store(多轨模型,主数据源);
// WorkStore 为展示主体,由后端按资源类型的 PrimaryRoles 优先级链派生(ResolvePrimaryStore),前端纯消费;
// ThumbnailStore 为从 Stores 按 storeType=thumbnail 派生的便捷访问器。
type ResourceFullDTO struct {
	ID               int64                `json:"id"`
	WorkID           int64                `json:"workId"`
	TaskID           int64                `json:"taskId"`
	Enabled          bool                 `json:"enabled"`
	SuggestName      *string              `json:"suggestName"`
	ResourceType     string               `json:"resourceType"`
	ResourceComplete int                  `json:"resourceComplete"`
	Stores           []ResourceStoreDTO   `json:"stores,omitempty"`
	WorkStore        *PersistentStoreDTO  `json:"workStore,omitempty"`
	ThumbnailStore   *PersistentStoreDTO  `json:"thumbnailStore,omitempty"`
	CreateTime       int64                `json:"createTime"`
	UpdateTime       int64                `json:"updateTime"`
}
