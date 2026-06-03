package dto

// ResourceDTO 资源数据传输对象
type ResourceDTO struct {
	ID                int64   `json:"id"`
	WorkID            int64   `json:"workId"`
	TaskID            int64   `json:"taskId"`
	Enabled           bool    `json:"enabled"`
	FilePath          *string `json:"filePath"`
	FileName          *string `json:"fileName"`
	FilenameExtension *string `json:"filenameExtension"`
	SuggestName       *string `json:"suggestName"`
	ResourceSize      *int64  `json:"resourceSize"`
	Workdir           *string `json:"workdir"`
	ResourceComplete  int     `json:"resourceComplete"`
	CreateTime        int64   `json:"createTime"`
	UpdateTime        int64   `json:"updateTime"`
}
