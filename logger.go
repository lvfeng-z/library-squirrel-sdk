package pluginsdk

// Logger 插件日志接口，可传递给插件子组件使用
// 通过 PluginContext.GetLogger() 获取，支持 Named 创建子模块 Logger
type Logger interface {
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Named(name string) Logger
}

// PluginLogger 插件侧 Logger 实现，通过 RPC 通知发送日志到主进程
type PluginLogger struct {
	client *RPCClient
	name   string
}

// NewPluginLogger 创建插件侧 Logger
func NewPluginLogger(client *RPCClient, name string) *PluginLogger {
	return &PluginLogger{client: client, name: name}
}

func (l *PluginLogger) Debugf(template string, args ...any) { l.log("ctx/debugf", template, args) }
func (l *PluginLogger) Infof(template string, args ...any)  { l.log("ctx/infof", template, args) }
func (l *PluginLogger) Warnf(template string, args ...any)  { l.log("ctx/warnf", template, args) }
func (l *PluginLogger) Errorf(template string, args ...any) { l.log("ctx/errorf", template, args) }

func (l *PluginLogger) Named(name string) Logger {
	newName := name
	if l.name != "" {
		newName = l.name + "." + name
	}
	return NewPluginLogger(l.client, newName)
}

func (l *PluginLogger) log(method, template string, args []any) {
	type params struct {
		Template    string `json:"template"`
		Args        []any  `json:"args"`
		LoggerName  string `json:"loggerName,omitempty"`
	}
	_ = l.client.Notify(method, params{Template: template, Args: args, LoggerName: l.name})
}
