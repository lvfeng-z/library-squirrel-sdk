package dto

// BackupDTO 备份数据传输对象
type BackupDTO struct {
	ID         int64   `json:"id"`
	SourceType *int64  `json:"sourceType"`
	SourceID   *int64  `json:"sourceId"`
	FileName   *string `json:"fileName"`
	FilePath   *string `json:"filePath"`
	Workdir    *string `json:"workdir"`
	CreateTime int64   `json:"createTime"`
	UpdateTime int64   `json:"updateTime"`
}
