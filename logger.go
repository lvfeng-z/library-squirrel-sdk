package pluginsdk

import (
	"context"
	"fmt"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/gen"
)

// Logger 插件日志接口，可传递给插件子组件使用
type Logger interface {
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Named(name string) Logger
}

// grpcLogger 通过 HostService gRPC 发送日志
type grpcLogger struct {
	client gen.HostServiceClient
	name   string
}

// NewGRPCLogger 创建基于 gRPC 的 Logger
func NewGRPCLogger(client gen.HostServiceClient) *grpcLogger {
	return &grpcLogger{client: client}
}

func (l *grpcLogger) Debugf(template string, args ...any) { l.log(0, template, args) }
func (l *grpcLogger) Infof(template string, args ...any)  { l.log(1, template, args) }
func (l *grpcLogger) Warnf(template string, args ...any)  { l.log(2, template, args) }
func (l *grpcLogger) Errorf(template string, args ...any) { l.log(3, template, args) }

func (l *grpcLogger) Named(name string) Logger {
	newName := name
	if l.name != "" {
		newName = l.name + "." + name
	}
	return &grpcLogger{client: l.client, name: newName}
}

func (l *grpcLogger) log(level int32, template string, args []any) {
	strArgs := make([]string, len(args))
	for i, a := range args {
		strArgs[i] = fmt.Sprintf("%v", a)
	}
	_, _ = l.client.Log(context.Background(), &gen.LogRequest{
		Level:      level,
		Template:   template,
		Args:       strArgs,
		LoggerName: l.name,
	})
}
