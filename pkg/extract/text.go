package extract

import (
	"io"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// ExtractFilePages extracts text, rectangles, and font stats from all pages.
func ExtractFilePages(path string) ([]model.PageResult, error) {
	f, reader, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return extractFromReader(reader)
}

// ExtractReaderPages extracts from an io.ReaderAt with known size.
func ExtractReaderPages(ra io.ReaderAt, size int64) ([]model.PageResult, error) {
	reader, err := pdf.NewReader(ra, size)
	if err != nil {
		return nil, err
	}
	return extractFromReader(reader)
}

func extractFromReader(reader *pdf.Reader) ([]model.PageResult, error) {
	numPages := reader.NumPage()
	results := make([]model.PageResult, numPages)
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		results[i-1] = extractPage(page, i)
	}
	return results, nil
}

func extractPage(page pdf.Page, pageNum int) model.PageResult {
	content := page.Content()
	result := model.PageResult{
		PageNum: pageNum,
		Texts:   make([]model.TextElement, 0, len(content.Text)),
		FontStats: model.FontStats{
			SizeCounts: make(map[float64]int),
			NameCounts: make(map[string]int),
		},
	}
	for _, t := range content.Text {
		if strings.TrimSpace(t.S) == "" {
			continue
		}
		bold, italic, mono := InferStyle(t.Font)
		result.Texts = append(result.Texts, model.TextElement{
			Text:     t.S,
			Font:     t.Font,
			FontSize: t.FontSize,
			X:        t.X,
			Y:        t.Y,
			W:        t.W,
			Bold:     bold,
			Italic:   italic,
			Mono:     mono,
		})
		charCount := utf8.RuneCountInString(t.S)
		result.FontStats.SizeCounts[t.FontSize] += charCount
		result.FontStats.NameCounts[t.Font] += charCount
	}
	for _, r := range content.Rect {
		result.Rects = append(result.Rects, model.Rectangle{
			MinX: r.Min.X,
			MinY: r.Min.Y,
			MaxX: r.Max.X,
			MaxY: r.Max.Y,
		})
	}
	return result
}

// InferStyle determines bold, italic, and monospace from a PDF font name.
func InferStyle(fontName string) (bold, italic, mono bool) {
	// Strip subset prefix (e.g., "ABCDEF+Arial" → "Arial")
	if idx := strings.Index(fontName, "+"); idx >= 0 && idx < 7 {
		fontName = fontName[idx+1:]
	}
	lower := strings.ToLower(fontName)

	bold = strings.Contains(lower, "bold") ||
		strings.HasSuffix(lower, "bd") ||
		strings.Contains(lower, "-bd") ||
		strings.Contains(lower, ",bd")

	italic = strings.Contains(lower, "italic") ||
		strings.Contains(lower, "oblique") ||
		strings.HasSuffix(lower, "it") ||
		strings.Contains(lower, "-it,") ||
		strings.Contains(lower, "-it ")

	mono = strings.Contains(lower, "courier") ||
		strings.Contains(lower, "mono") ||
		strings.Contains(lower, "consolas") ||
		strings.Contains(lower, "fixed") ||
		strings.Contains(lower, "menlo") ||
		strings.Contains(lower, "monaco")

	return
}
