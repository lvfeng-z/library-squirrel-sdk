package dto

// PersistentStoreDTO 文件持久存储数据传输对象
type PersistentStoreDTO struct {
	ID                int64   `json:"id"`
	FilePath          *string `json:"filePath"`
	FileName          *string `json:"fileName"`
	FilenameExtension *string `json:"filenameExtension"`
	Status            int     `json:"status"` // 0=未完成，1=完成
	CreateTime        int64   `json:"createTime"`
	UpdateTime        int64   `json:"updateTime"`
}
