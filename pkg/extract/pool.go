package extract

import (
	"bytes"
	"sync"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

var textElementPool = sync.Pool{
	New: func() any {
		s := make([]model.TextElement, 0, 256)
		return &s
	},
}

func AcquireTextElements() *[]model.TextElement {
	return textElementPool.Get().(*[]model.TextElement)
}

func ReleaseTextElements(s *[]model.TextElement) {
	*s = (*s)[:0]
	textElementPool.Put(s)
}

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

func AcquireBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
