package dto

// Logger 插件日志接口，可传递给插件子组件使用
type Logger interface {
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Named(name string) Logger
}
