package dto

import "encoding/json"

// InstallType 安装类型
type InstallType int

const (
	InstallTypeManual InstallType = 0
	InstallTypeAuto   InstallType = 1
)

// ActivationType 激活类型
type ActivationType int

const (
	ActivationTypeManual  ActivationType = 0
	ActivationTypeStartup ActivationType = 1
)

// PluginActivation 插件激活配置
type PluginActivation struct {
	Type ActivationType `json:"type"`
}

// PluginManifest 插件清单（从 plugin.json 解析）
type PluginManifest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description,omitempty"`
	Extensions  *PluginExtensions `json:"extensions"`
	Activation  PluginActivation  `json:"activation"`
	EntryFile   string            `json:"entryFile"`
}

// PluginInstallDTO 插件安装数据传输对象
type PluginInstallDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description,omitempty"`
	Extensions  *PluginExtensions `json:"extensions"`
	Activation  PluginActivation  `json:"activation"`
	EntryFile   string            `json:"entryFile"`
	PackagePath string            `json:"packagePath,omitempty"`
	PublicID    string            `json:"publicId,omitempty"`
}

// GetPublicID 获取插件公开ID（格式：作者/名称）
func (p *PluginManifest) GetPublicID() string {
	return p.Author + "/" + p.ID
}

// ToPluginInstallDTO 转换为安装DTO
func (p *PluginManifest) ToPluginInstallDTO(packagePath string) *PluginInstallDTO {
	return &PluginInstallDTO{
		ID:          p.ID,
		Name:        p.Name,
		Version:     p.Version,
		Author:      p.Author,
		Description: p.Description,
		Extensions:  p.Extensions,
		Activation:  p.Activation,
		EntryFile:   p.EntryFile,
		PackagePath: packagePath,
		PublicID:    p.GetPublicID(),
	}
}

// NewPluginManifest 创建插件清单
func NewPluginManifest() *PluginManifest {
	return &PluginManifest{}
}

// PluginExtensions 插件扩展点集合
type PluginExtensions struct {
	TaskHandlers    []TaskHandlerDeclaration `json:"taskHandlers,omitempty"`
	SiteBrowsers    []SiteBrowserDeclaration `json:"siteBrowsers,omitempty"`
	FrontendExtensions []FrontendExtensionDeclaration `json:"frontendExtensions,omitempty"`
	StaticResources *StaticResourcesConfig   `json:"staticResources,omitempty"`
}

// TaskHandlerDeclaration 任务处理器声明
type TaskHandlerDeclaration struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SiteBrowserDeclaration 站点浏览器声明
type SiteBrowserDeclaration struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// FrontendExtensionDeclaration 前端扩展声明
type FrontendExtensionDeclaration struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Kind        string          `json:"kind"`
	Order       int             `json:"order,omitempty"`
	Content     json.RawMessage `json:"content"`
}

// EmbedContent embed 类型前端扩展配置
type EmbedContent struct {
	ContentType    string          `json:"contentType"`
	Source         json.RawMessage `json:"source"`
	Position       string          `json:"position"`
	ExtensionId string          `json:"extensionId,omitempty"`
	Props          json.RawMessage `json:"props,omitempty"`
}

// ViewContent view 类型前端扩展配置
type ViewContent struct {
	ContentType string          `json:"contentType"`
	Source      json.RawMessage `json:"source"`
	Title       string          `json:"title,omitempty"`
	Props       json.RawMessage `json:"props,omitempty"`
}

// MenuContent menu 类型前端扩展配置
type MenuContent struct {
	Icon     string                          `json:"icon,omitempty"`
	ViewId   string                          `json:"viewId,omitempty"`
	Children []FrontendExtensionDeclaration  `json:"children,omitempty"`
}

// SiteBrowserListContent siteBrowserList 类型前端扩展配置
type SiteBrowserListContent struct {
	Icon           string `json:"icon,omitempty"`
	ExtensionId string `json:"extensionId"`
}

// StaticResourcesConfig 静态资源配置
type StaticResourcesConfig struct {
	Directories []string `json:"directories"`
}
