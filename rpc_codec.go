package pluginsdk

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

// Message types for the frame protocol
const (
	FrameTypeJSON   byte = 0x01 // JSON-RPC message (request/response/notification)
	FrameTypeBinary byte = 0x02 // Binary data chunk (streaming)
)

// CodecErrorCodec errors
var (
	ErrInvalidFrameLength = errors.New("invalid frame: length too short")
	ErrInvalidFrameType   = errors.New("invalid frame: unknown type byte")
	ErrFrameIncomplete    = errors.New("incomplete frame: not enough bytes read")
)

// FrameCodec 基于 net.Conn 的帧读写器
// 支持混合 JSON-RPC 消息和二进制数据帧的全双工传输
type FrameCodec struct {
	conn io.ReadWriter
	buf  []byte
}

// NewFrameCodec 创建帧编解码器
func NewFrameCodec(conn io.ReadWriter) *FrameCodec {
	return &FrameCodec{
		conn: conn,
		buf:  make([]byte, 0, 4096),
	}
}

// WriteJSON 写入 JSON-RPC 消息
func (c *FrameCodec) WriteJSON(data []byte) error {
	return c.writeFrame(FrameTypeJSON, data)
}

// WriteBinary 写入二进制数据帧
// streamID: 流标识符
// data: 数据块（长度为 0 表示 EOF）
func (c *FrameCodec) WriteBinary(streamID string, data []byte) error {
	// 帧格式: [1字节 streamID长度] [N字节 streamID] [数据块]
	frameLen := 1 + len(streamID) + len(data)
	headerLen := 4 + 1 // length prefix + type byte

	buf := make([]byte, headerLen+frameLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(frameLen))
	buf[4] = FrameTypeBinary
	offset := 5

	buf[offset] = byte(len(streamID))
	offset++

	copy(buf[offset:], streamID)
	offset += len(streamID)

	copy(buf[offset:], data)
	offset += len(data)

	_, err := c.conn.Write(buf)
	return err
}

// ReadFrame 读取一帧数据
// 返回 (payload, frameType, error)
func (c *FrameCodec) ReadFrame() ([]byte, byte, error) {
	// 读取帧头：4字节长度 + 1字节类型
	header := make([]byte, 5)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, 0, err
	}

	frameLen := binary.BigEndian.Uint32(header[0:4])
	frameType := header[4]

	if frameLen == 0 {
		return nil, frameType, nil
	}

	payload := make([]byte, frameLen)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return nil, 0, err
	}

	return payload, frameType, nil
}

// ReadJSON 读取一个 JSON-RPC 消息帧
func (c *FrameCodec) ReadJSON() ([]byte, error) {
	payload, frameType, err := c.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frameType != FrameTypeJSON {
		return nil, ErrInvalidFrameType
	}
	return payload, nil
}

// writeFrame 写入一帧（内部方法）
func (c *FrameCodec) writeFrame(frameType byte, data []byte) error {
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header[0:4], uint32(len(data)))
	header[4] = frameType

	var buf []byte
	if len(data) <= 1024 {
		buf = make([]byte, 5+len(data))
		copy(buf, header)
		copy(buf[5:], data)
	} else {
		buf = make([]byte, 5)
		copy(buf, header)
		buf = append(buf, data...)
	}

	_, err := c.conn.Write(buf)
	return err
}

// FrameWriter 线程安全的帧写入器封装
type FrameWriter struct {
	codec *FrameCodec
	mu    sync.Mutex
}

// NewFrameWriter 创建帧写入器
func NewFrameWriter(codec *FrameCodec) *FrameWriter {
	return &FrameWriter{codec: codec}
}

// WriteJSON 线程安全写入 JSON
func (w *FrameWriter) WriteJSON(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.codec.WriteJSON(data)
}

// WriteBinary 线程安全写入二进制帧
func (w *FrameWriter) WriteBinary(streamID string, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.codec.WriteBinary(streamID, data)
}
