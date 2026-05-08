package pluginsdk

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// StreamWriter 将 io.ReadCloser 的数据以二进制帧流式传输到连接
// 用于 TaskHandler.Start() 的结果流
type StreamWriter struct {
	codec    *FrameCodec
	mu       sync.Mutex
	closed   atomic.Bool
	streamID string
}

// NewStreamWriter 创建 StreamWriter
func NewStreamWriter(codec *FrameCodec, streamID string) *StreamWriter {
	return &StreamWriter{
		codec:    codec,
		streamID: streamID,
	}
}

// Stream 将 reader 的数据流式传输到连接
// 使用 32KB 块进行传输
// 传输完成后自动发送 EOF 帧（零长度数据块）
func (w *StreamWriter) Stream(reader io.Reader) error {
	buf := make([]byte, 32*1024)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if err := w.codec.WriteBinary(w.streamID, buf[:n]); err != nil {
				return fmt.Errorf("write binary frame: %w", err)
			}
		}

		if err != nil {
			// 发送 EOF（零长度数据块）
			if err := w.codec.WriteBinary(w.streamID, nil); err != nil {
				return fmt.Errorf("write EOF frame: %w", err)
			}
			return err
		}
	}
}

// StreamWithProgress 流式传输并报告进度
// progressCallback 在每个数据块传输完成后调用
func (w *StreamWriter) StreamWithProgress(reader io.Reader, progressCallback func(bytesRead int64)) error {
	buf := make([]byte, 32*1024)
	var totalRead int64

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			totalRead += int64(n)
			if err := w.codec.WriteBinary(w.streamID, buf[:n]); err != nil {
				return fmt.Errorf("write binary frame: %w", err)
			}
			if progressCallback != nil {
				progressCallback(totalRead)
			}
		}

		if err != nil {
			if err := w.codec.WriteBinary(w.streamID, nil); err != nil {
				return fmt.Errorf("write EOF frame: %w", err)
			}
			return err
		}
	}
}

// Close 标记流为已关闭（不实际关闭底层的 codec）
func (w *StreamWriter) Close() error {
	w.closed.Store(true)
	return nil
}
