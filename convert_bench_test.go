package pdf2md

import (
	"bytes"
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/segfaultd/pdf-to-markdown/pkg/analyze"
	"github.com/segfaultd/pdf-to-markdown/pkg/emit"
	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func generateBenchPDF(pages int) []byte {
	p := fpdf.New("P", "pt", "A4", "")

	for i := 0; i < pages; i++ {
		p.AddPage()

		// Title
		p.SetFont("Helvetica", "B", 24)
		p.Text(72, 72, "Chapter Title")

		// Body paragraphs
		p.SetFont("Helvetica", "", 12)
		y := float64(120)
		for j := 0; j < 20; j++ {
			p.Text(72, y, "This is a line of body text that represents typical document content for benchmarking.")
			y += 16
		}
	}

	var buf bytes.Buffer
	p.Output(&buf)
	return buf.Bytes()
}

func BenchmarkConvertSmall(b *testing.B) {
	data := generateBenchPDF(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertBytes(data)
	}
}

func BenchmarkConvertMedium(b *testing.B) {
	data := generateBenchPDF(30)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertBytes(data)
	}
}

func BenchmarkConvertLarge(b *testing.B) {
	data := generateBenchPDF(200)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ConvertBytes(data)
	}
}

func BenchmarkLineReconstruct(b *testing.B) {
	texts := make([]model.TextElement, 100)
	y := 720.0
	for i := range texts {
		texts[i] = model.TextElement{
			Text: "word", Font: "Helvetica", FontSize: 12,
			X: float64(72 + (i%10)*50), Y: y, W: 30,
		}
		if i%10 == 9 {
			y -= 16
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyze.ReconstructLines(texts)
	}
}

func BenchmarkEmitMarkdown(b *testing.B) {
	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Lines: []model.Line{
			{Spans: []model.Span{{Text: "Title"}}},
		}},
		{Type: model.BlockParagraph, Lines: []model.Line{
			{Spans: []model.Span{{Text: "Body text here."}}},
			{Spans: []model.Span{{Text: "More body text."}}},
		}},
		{Type: model.BlockTable, Table: &model.Table{
			Headers: []string{"A", "B", "C"},
			Rows:    [][]string{{"1", "2", "3"}, {"4", "5", "6"}},
			Aligns:  []model.Align{model.AlignLeft, model.AlignCenter, model.AlignRight},
		}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		emit.Emit(blocks, &buf)
	}
}

func BenchmarkAllocsPerPage(b *testing.B) {
	data := generateBenchPDF(1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ConvertBytes(data)
	}
}
