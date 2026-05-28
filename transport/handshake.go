package transport

import (
	"os/exec"

	"github.com/hashicorp/go-plugin"
)

// Handshake 握手配置，主程序和插件必须使用相同的值
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "LIBRARY_SQUIRREL_PLUGIN",
	MagicCookieValue: "d9a7f2b1e5c3846021de7baf0953c8e7",
}

// PluginMap 插件类型映射
var PluginMap = map[string]plugin.Plugin{
	"library_squirrel": &LSPlugin{},
}

// NewClientConfig 创建 hashicorp/go-plugin 的 Client 配置
func NewClientConfig(cmd *exec.Cmd) *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}
}
