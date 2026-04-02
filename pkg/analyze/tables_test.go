package analyze

import (
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func TestDetectTables_WithRectangles(t *testing.T) {
	// 2x2 table with cell rectangles
	rects := []model.Rectangle{
		{MinX: 72, MinY: 700, MaxX: 200, MaxY: 720},
		{MinX: 200, MinY: 700, MaxX: 350, MaxY: 720},
		{MinX: 72, MinY: 680, MaxX: 200, MaxY: 700},
		{MinX: 200, MinY: 680, MaxX: 350, MaxY: 700},
	}

	texts := []model.TextElement{
		{Text: "Name", X: 80, Y: 710, FontSize: 12, Bold: true},
		{Text: "Age", X: 210, Y: 710, FontSize: 12, Bold: true},
		{Text: "Alice", X: 80, Y: 690, FontSize: 12},
		{Text: "30", X: 210, Y: 690, FontSize: 12},
	}

	tables, consumed := DetectTables(texts, rects)

	if len(tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(tables))
	}
	if len(consumed) != 4 {
		t.Errorf("consumed %d elements, want 4", len(consumed))
	}

	tbl := tables[0].Table
	if len(tbl.Headers) != 2 || tbl.Headers[0] != "Name" || tbl.Headers[1] != "Age" {
		t.Errorf("headers = %v, want [Name Age]", tbl.Headers)
	}
	if len(tbl.Rows) != 1 || tbl.Rows[0][0] != "Alice" || tbl.Rows[0][1] != "30" {
		t.Errorf("rows = %v, want [[Alice 30]]", tbl.Rows)
	}
}

func TestDetectTables_ColumnAlignment(t *testing.T) {
	// No rectangles, but columnar text alignment
	texts := []model.TextElement{
		{Text: "Name", X: 72, Y: 720, W: 40, FontSize: 12, Bold: true},
		{Text: "Score", X: 200, Y: 720, W: 40, FontSize: 12, Bold: true},
		{Text: "Alice", X: 72, Y: 706, W: 36, FontSize: 12},
		{Text: "95", X: 200, Y: 706, W: 16, FontSize: 12},
		{Text: "Bob", X: 72, Y: 692, W: 26, FontSize: 12},
		{Text: "87", X: 200, Y: 692, W: 16, FontSize: 12},
		{Text: "Carol", X: 72, Y: 678, W: 38, FontSize: 12},
		{Text: "92", X: 200, Y: 678, W: 16, FontSize: 12},
	}

	tables, consumed := DetectTables(texts, nil)

	if len(tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(tables))
	}
	if len(consumed) != 8 {
		t.Errorf("consumed %d, want 8", len(consumed))
	}

	tbl := tables[0].Table
	if len(tbl.Rows) != 3 {
		t.Errorf("got %d rows, want 3", len(tbl.Rows))
	}
}

func TestDetectTables_NoTable(t *testing.T) {
	texts := []model.TextElement{
		{Text: "Just a normal paragraph.", X: 72, Y: 720, W: 150, FontSize: 12},
		{Text: "Another line of text.", X: 72, Y: 706, W: 130, FontSize: 12},
	}

	tables, consumed := DetectTables(texts, nil)

	if len(tables) != 0 {
		t.Errorf("got %d tables, want 0", len(tables))
	}
	if len(consumed) != 0 {
		t.Errorf("consumed %d, want 0", len(consumed))
	}
}
