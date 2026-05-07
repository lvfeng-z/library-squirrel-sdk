package pluginsdk

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSON-RPC 2.0 服务端（供主进程侧使用，接收插件的请求并处理）
type RPCServer struct {
	codec    *FrameCodec
	handlers map[string]Handler // method -> handler function
}

// Handler 是 RPC 方法处理函数的签名
// params 是请求参数（已反序列化的 JSON）
// result 是需要序列化回写入响应结果的指针
type Handler func(params json.RawMessage) (result any, err error)

// NewRPCServer 创建 JSON-RPC 服务端
func NewRPCServer(codec *FrameCodec) *RPCServer {
	return &RPCServer{
		codec:    codec,
		handlers: make(map[string]Handler),
	}
}

// Register 注册 RPC 方法处理函数
func (s *RPCServer) Register(method string, handler Handler) {
	s.handlers[method] = handler
}

// HandleIncomingMessage 处理收到的数据帧
// 由主进程的帧读取循环调用
func (s *RPCServer) HandleIncomingMessage(frameType byte, payload []byte) error {
	if frameType == FrameTypeBinary {
		// 二进制帧不由服务端处理（由 StreamReader 处理）
		return nil
	}
	if frameType != FrameTypeJSON {
		return ErrInvalidFrameType
	}

	// 解析请求
	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		ID     int64           `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("unmarshal request: %w", err)
	}

	if req.JSONRPC != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %s", req.JSONRPC)
	}

	handler, ok := s.handlers[req.Method]
	if !ok {
		// 未知方法，返回错误
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"error": map[string]any{
				"code":    -32601,
				"message": fmt.Sprintf("method not found: %s", req.Method),
			},
		}
		respData, _ := json.Marshal(resp)
		return s.codec.WriteJSON(respData)
	}

	result, err := handler(req.Params)
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req.ID,
	}
	if err != nil {
		resp["error"] = map[string]any{
			"code":    -32000,
			"message": err.Error(),
		}
	} else if result != nil {
		resp["result"] = result
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	return s.codec.WriteJSON(respData)
}

// Serve 启动服务端主循环（阻塞）
// frameReader 返回 (payload, frameType, error)
// 通常实现为从 net.Conn 读取帧
func (s *RPCServer) Serve(frameReader func() ([]byte, byte, error)) error {
	for {
		payload, frameType, err := frameReader()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil // 连接正常关闭
			}
			return fmt.Errorf("read frame: %w", err)
		}

		if err := s.HandleIncomingMessage(frameType, payload); err != nil {
			// 记录错误但继续运行
			fmt.Printf("handle message error: %v\n", err)
		}
	}
}
