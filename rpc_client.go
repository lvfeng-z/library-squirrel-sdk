package pluginsdk

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// JSON-RPC 2.0 客户端（供插件侧使用，向主进程发送请求并等待响应）
type RPCClient struct {
	codec     *FrameCodec
	writer    *FrameWriter
	nextID    atomic.Int64
	pending   map[int64]chan *rpcResponse
	mu        sync.RWMutex
	done      chan struct{}
	closeOnce sync.Once
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewRPCClient 创建 JSON-RPC 客户端
func NewRPCClient(codec *FrameCodec) *RPCClient {
	return &RPCClient{
		codec:   codec,
		writer:  NewFrameWriter(codec),
		pending: make(map[int64]chan *rpcResponse),
		done:    make(chan struct{}),
	}
}

// Close 关闭客户端，解除所有阻塞的 Call
func (c *RPCClient) Close() {
	c.closeOnce.Do(func() { close(c.done) })
}

// Call 发送 JSON-RPC 请求并等待响应（同步）
func (c *RPCClient) Call(method string, params, result any) error {
	id := c.nextID.Add(1)

	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		req.Params = data
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// 先注册 pending channel，再发送请求，避免响应在注册前到达被丢弃
	ch := make(chan *rpcResponse, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	if err := c.writer.WriteJSON(reqData); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return fmt.Errorf("send request: %w", err)
	}

	var resp *rpcResponse
	select {
	case resp = <-ch:
	case <-c.done:
		return ErrPluginCrashed
	}
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}

	return nil
}

// Notify 发送 JSON-RPC 通知（不等待响应）
func (c *RPCClient) Notify(method string, params any) error {
	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		req["params"] = data
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal notify: %w", err)
	}

	return c.writer.WriteJSON(reqData)
}

// HandleResponse 处理收到的响应（由帧读取器调用）
func (c *RPCClient) HandleResponse(respData []byte) error {
	var resp rpcResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	c.mu.RLock()
	ch, ok := c.pending[resp.ID]
	c.mu.RUnlock()
	if !ok {
		return nil // 无关的响应（如已超时的）
	}

	select {
	case ch <- &resp:
	default:
	}

	return nil
}
