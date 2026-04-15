package analyze

import (
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func TestReconstructLines_BasicGrouping(t *testing.T) {
	// Two lines of text: "Hello World" at Y=720, "Body text here" at Y=700
	texts := []model.TextElement{
		{Text: "Hello", Font: "Helvetica-Bold", FontSize: 24, X: 72, Y: 720, W: 60, Bold: true},
		{Text: "World", Font: "Helvetica-Bold", FontSize: 24, X: 140, Y: 720, W: 62, Bold: true},
		{Text: "Body", Font: "Helvetica", FontSize: 12, X: 72, Y: 700, W: 30},
		{Text: "text", Font: "Helvetica", FontSize: 12, X: 108, Y: 700, W: 24},
		{Text: "here", Font: "Helvetica", FontSize: 12, X: 138, Y: 700, W: 26},
	}

	lines := ReconstructLines(texts)

	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	// First line (higher Y = top of page in PDF coords)
	if lines[0].FontSize != 24 {
		t.Errorf("line[0].FontSize = %v, want 24", lines[0].FontSize)
	}
	got := fullLineText(lines[0])
	if got != "Hello World" {
		t.Errorf("line[0] text = %q, want %q", got, "Hello World")
	}

	// Second line
	if lines[1].FontSize != 12 {
		t.Errorf("line[1].FontSize = %v, want 12", lines[1].FontSize)
	}
	got = fullLineText(lines[1])
	if got != "Body text here" {
		t.Errorf("line[1] text = %q, want %q", got, "Body text here")
	}
}

func TestReconstructLines_FontChangeCreatesNewSpan(t *testing.T) {
	texts := []model.TextElement{
		{Text: "Normal", Font: "Helvetica", FontSize: 12, X: 72, Y: 700, W: 42},
		{Text: "bold", Font: "Helvetica-Bold", FontSize: 12, X: 120, Y: 700, W: 28, Bold: true},
		{Text: "text", Font: "Helvetica", FontSize: 12, X: 154, Y: 700, W: 24},
	}

	lines := ReconstructLines(texts)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if len(lines[0].Spans) != 3 {
		t.Fatalf("got %d spans, want 3", len(lines[0].Spans))
	}
	if lines[0].Spans[0].Bold || !lines[0].Spans[1].Bold || lines[0].Spans[2].Bold {
		t.Error("span bold flags incorrect")
	}
}

func TestReconstructLines_LargeGapInsertsSpace(t *testing.T) {
	texts := []model.TextElement{
		{Text: "Word1", Font: "Helvetica", FontSize: 12, X: 72, Y: 700, W: 35},
		// gap of 13 (> 0.3*12=3.6 but < 1.5*12=18) → insert space
		{Text: "Word2", Font: "Helvetica", FontSize: 12, X: 120, Y: 700, W: 35},
	}

	lines := ReconstructLines(texts)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	got := fullLineText(lines[0])
	if got != "Word1 Word2" {
		t.Errorf("text = %q, want %q", got, "Word1 Word2")
	}
}

func TestReconstructLines_EmptyInput(t *testing.T) {
	lines := ReconstructLines(nil)
	if len(lines) != 0 {
		t.Errorf("got %d lines for nil input, want 0", len(lines))
	}
}

func TestReconstructLines_PerGlyphPDF(t *testing.T) {
	// Simulates a PDF that emits one character per TextElement with W=0
	// and nearly identical X positions. Document order = reading order.
	// The X values are NOT monotonic (mimics real per-glyph PDFs).
	texts := []model.TextElement{
		{Text: "H", FontSize: 10, X: 100.08, Y: 500, W: 0},
		{Text: "e", FontSize: 10, X: 100.35, Y: 500, W: 0},
		{Text: "l", FontSize: 10, X: 100.36, Y: 500, W: 0},
		{Text: "l", FontSize: 10, X: 100.30, Y: 500, W: 0}, // X goes backward
		{Text: "o", FontSize: 10, X: 100.42, Y: 500, W: 0},
		{Text: " ", FontSize: 10, X: 100.33, Y: 500, W: 0}, // space
		{Text: "W", FontSize: 10, X: 100.50, Y: 500, W: 0},
		{Text: "o", FontSize: 10, X: 100.45, Y: 500, W: 0}, // X goes backward again
		{Text: "r", FontSize: 10, X: 100.52, Y: 500, W: 0},
		{Text: "l", FontSize: 10, X: 100.53, Y: 500, W: 0},
		{Text: "d", FontSize: 10, X: 100.54, Y: 500, W: 0},
	}

	lines := ReconstructLines(texts)

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	got := fullLineText(lines[0])
	if got != "Hello World" {
		t.Errorf("got %q, want %q", got, "Hello World")
	}
}

func TestReconstructLines_PerGlyphMultiColumn(t *testing.T) {
	// Per-glyph PDF with two columns: "Nom" at X~72 and "Age" at X~200.
	// Each column's glyphs cluster together in document order.
	texts := []model.TextElement{
		{Text: "N", FontSize: 10, X: 72.01, Y: 500, W: 0},
		{Text: "o", FontSize: 10, X: 72.02, Y: 500, W: 0},
		{Text: "m", FontSize: 10, X: 72.03, Y: 500, W: 0},
		{Text: "A", FontSize: 10, X: 200.01, Y: 500, W: 0},
		{Text: "g", FontSize: 10, X: 200.02, Y: 500, W: 0},
		{Text: "e", FontSize: 10, X: 200.03, Y: 500, W: 0},
	}

	lines := ReconstructLines(texts)

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	got := fullLineText(lines[0])
	if got != "Nom Age" {
		t.Errorf("got %q, want %q", got, "Nom Age")
	}
	// First column should come first
	if lines[0].X > 100 {
		t.Errorf("line X = %v, want ~72 (first column)", lines[0].X)
	}
}

func TestReconstructLines_DoesNotMutateInput(t *testing.T) {
	texts := []model.TextElement{
		{Text: "B", FontSize: 12, X: 200, Y: 500, W: 10},
		{Text: "A", FontSize: 12, X: 72, Y: 700, W: 10},
	}
	origFirst := texts[0].Text

	ReconstructLines(texts)

	if texts[0].Text != origFirst {
		t.Errorf("input was mutated: texts[0].Text changed from %q to %q", origFirst, texts[0].Text)
	}
}

func TestReconstructLines_SingleElement(t *testing.T) {
	texts := []model.TextElement{
		{Text: "alone", FontSize: 12, X: 72, Y: 700, W: 40},
	}

	lines := ReconstructLines(texts)

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if fullLineText(lines[0]) != "alone" {
		t.Errorf("got %q, want %q", fullLineText(lines[0]), "alone")
	}
}

// fullLineText concatenates all span text to get the full readable line.
func fullLineText(line model.Line) string {
	var s string
	for _, sp := range line.Spans {
		s += sp.Text
	}
	return s
}
