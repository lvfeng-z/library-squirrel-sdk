package dto

// PluginDTO 插件
type PluginDTO struct {
	ID             int64   `json:"id"`
	PublicID       *string `json:"publicId"`
	Author         *string `json:"author"`
	Name           *string `json:"name"`
	Version        *string `json:"version"`
	Description    *string `json:"description"`
	Changelog      *string `json:"changelog"`
	EntryPath      *string `json:"entryPath"`
	RootPath       *string `json:"rootPath"`
	BackupID       *int64  `json:"backupId"`
	SortNum        *int64  `json:"sortNum"`
	PluginData     *string `json:"pluginData"`
	Uninstalled    *bool   `json:"uninstalled"`
	ActivationType *string `json:"activationType"`
	CreateTime     int64   `json:"createTime"`
	UpdateTime     int64   `json:"updateTime"`
}
