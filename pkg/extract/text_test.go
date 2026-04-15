package extract

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func createTestPDF(t *testing.T) string {
	t.Helper()
	p := fpdf.New("P", "pt", "A4", "")
	p.AddPage()

	// Title (bold, large)
	p.SetFont("Helvetica", "B", 24)
	p.Text(72, 72, "Test Title")

	// Body text
	p.SetFont("Helvetica", "", 12)
	p.Text(72, 120, "This is body text.")
	p.Text(72, 136, "Second line of body.")

	// Italic text
	p.SetFont("Helvetica", "I", 12)
	p.Text(72, 170, "Italic note.")

	path := filepath.Join(t.TempDir(), "test.pdf")
	if err := p.OutputFileAndClose(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractFilePages(t *testing.T) {
	path := createTestPDF(t)

	results, err := ExtractFilePages(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d page results, want 1", len(results))
	}

	r := results[0]
	if len(r.Texts) == 0 {
		t.Fatal("no text elements extracted")
	}

	// Verify we got text content.
	// The ledongthuc/pdf library may return one TextElement per character,
	// so we join all text and strip spaces before checking for expected strings.
	var joined strings.Builder
	for _, te := range r.Texts {
		joined.WriteString(te.Text)
	}
	allText := joined.String()
	allTextNorm := strings.ReplaceAll(allText, " ", "")

	if !strings.Contains(allTextNorm, "TestTitle") {
		t.Errorf("missing title in extracted text: %q", allText)
	}
	if !strings.Contains(allTextNorm, "body") {
		t.Errorf("missing body text: %q", allText)
	}

	// Check font stats were collected
	if len(r.FontStats.SizeCounts) == 0 {
		t.Error("no font size stats collected")
	}
	if len(r.FontStats.NameCounts) == 0 {
		t.Error("no font name stats collected")
	}
}

func TestExtractReaderPages(t *testing.T) {
	path := createTestPDF(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ra := bytes.NewReader(data)
	results, err := ExtractReaderPages(ra, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d pages, want 1", len(results))
	}
	if len(results[0].Texts) == 0 {
		t.Error("no text elements extracted from reader")
	}
}

func TestExtractMultiplePages(t *testing.T) {
	p := fpdf.New("P", "pt", "A4", "")
	for i := 1; i <= 3; i++ {
		p.AddPage()
		p.SetFont("Helvetica", "", 12)
		p.Text(72, 72, "page content")
	}
	path := filepath.Join(t.TempDir(), "multi.pdf")
	if err := p.OutputFileAndClose(path); err != nil {
		t.Fatal(err)
	}

	results, err := ExtractFilePages(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d pages, want 3", len(results))
	}
	for i, r := range results {
		if r.PageNum != i+1 {
			t.Errorf("page %d has PageNum=%d", i, r.PageNum)
		}
		if len(r.Texts) == 0 {
			t.Errorf("page %d has no text", i+1)
		}
	}
}

func TestInferStyle(t *testing.T) {
	tests := []struct {
		font               string
		bold, italic, mono bool
	}{
		{"Helvetica", false, false, false},
		{"Helvetica-Bold", true, false, false},
		{"Helvetica-BoldOblique", true, true, false},
		{"Helvetica-Oblique", false, true, false},
		{"TimesNewRoman-BdIt", true, true, false},
		{"Courier", false, false, true},
		{"CourierNew-Bold", true, false, true},
		{"ABCDEF+Arial-BoldMT", true, false, false},
		{"Monaco", false, false, true},
		{"Menlo-Regular", false, false, true},
		{"Consolas", false, false, true},
		{"Consolas-Bold", true, false, true},
		{"Helvetica-Italic", false, true, false},
		{"DejaVuSansMono", false, false, true},
		{"BCDEFG+MyCustomFont", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.font, func(t *testing.T) {
			bold, italic, mono := InferStyle(tt.font)
			if bold != tt.bold {
				t.Errorf("InferStyle(%q).bold = %v, want %v", tt.font, bold, tt.bold)
			}
			if italic != tt.italic {
				t.Errorf("InferStyle(%q).italic = %v, want %v", tt.font, italic, tt.italic)
			}
			if mono != tt.mono {
				t.Errorf("InferStyle(%q).mono = %v, want %v", tt.font, mono, tt.mono)
			}
		})
	}
}

// Ensure model import is used (PageResult is referenced in function signatures).
var _ model.PageResult
