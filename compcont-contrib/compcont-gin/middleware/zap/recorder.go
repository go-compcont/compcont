package zap

import (
	"bytes"
	"encoding/base64"
	"io"
	"unicode"

	"github.com/gin-gonic/gin"
)

func bytes2String(data []byte) string {
	// isPrintableChar 判断字节是否为可打印字符
	isPrintableChar := func(b byte) bool {
		return unicode.IsPrint(rune(b)) || b == '\n' || b == '\r' || b == '\t'
	}
	// isBinary 判断内容是否为二进制
	isBinary := func(data []byte) bool {
		for _, b := range data {
			if b < 32 && !isPrintableChar(b) {
				return true
			}
		}
		return false
	}
	if isBinary(data) {
		return base64.StdEncoding.EncodeToString(data)
	} else {
		return string(data)
	}
}

// readCloserRecorder 是一个自定义的 ReadCloser，用于记录读取的前 n 个字节
type readCloserRecorder struct {
	reader io.ReadCloser
	buffer *bytes.Buffer
	limit  int
}

// newReadCloserRecorder 创建一个新的 readCloserRecorder，并在创建时读取前 n 个字节
func newReadCloserRecorder(rdc io.ReadCloser, limit int) *readCloserRecorder {
	buffer := bytes.NewBuffer(make([]byte, 0, limit))
	limitedReader := io.LimitReader(rdc, int64(limit))
	io.Copy(buffer, limitedReader)
	return &readCloserRecorder{
		reader: rdc,
		buffer: buffer,
		limit:  limit,
	}
}

func (r *readCloserRecorder) LimitedBody() []byte {
	return r.buffer.Bytes()
}

// Read 实现 io.Reader 接口
func (r *readCloserRecorder) Read(p []byte) (n int, err error) {
	if r.buffer.Len() > 0 {
		return r.buffer.Read(p)
	}
	return r.reader.Read(p)
}

// Close 实现 io.Closer 接口
func (r *readCloserRecorder) Close() error {
	return r.reader.Close()
}

// writeCloserRecorder 是一个自定义的 WriteCloser，用于记录写入的前 n 个字节
type writeCloserRecorder struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
	limit  int
}

// newWriteCloserRecorder 创建一个新的 writeCloserRecorder
func newWriteCloserRecorder(rwr gin.ResponseWriter, limit int) *writeCloserRecorder {
	return &writeCloserRecorder{
		ResponseWriter: rwr,
		buffer:         bytes.NewBuffer(make([]byte, 0, limit)),
		limit:          limit,
	}
}

func (w *writeCloserRecorder) LimitedBody() []byte {
	return w.buffer.Bytes()
}

// Write 实现 io.Writer 接口
func (w *writeCloserRecorder) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	if w.buffer.Len() < w.limit {
		remaining := w.limit - w.buffer.Len()
		if n > remaining {
			w.buffer.Write(p[:remaining])
		} else {
			w.buffer.Write(p[:n])
		}
	}
	return n, err
}
