package extract

import (
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func BenchmarkTextElementPool(b *testing.B) {
	b.Run("with_pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := AcquireTextElements()
			for j := 0; j < 100; j++ {
				*s = append(*s, model.TextElement{Text: "test"})
			}
			ReleaseTextElements(s)
		}
	})

	b.Run("without_pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := make([]model.TextElement, 0, 256)
			for j := 0; j < 100; j++ {
				s = append(s, model.TextElement{Text: "test"})
			}
		}
	})
}

func BenchmarkBufferPool(b *testing.B) {
	b.Run("with_pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := AcquireBuffer()
			for j := 0; j < 100; j++ {
				buf.WriteString("test data here ")
			}
			ReleaseBuffer(buf)
		}
	})

	b.Run("without_pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 0, 4096)
			for j := 0; j < 100; j++ {
				buf = append(buf, "test data here "...)
			}
		}
	})
}
