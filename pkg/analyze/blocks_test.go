package analyze

import (
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func TestClassifyBlocks_HeadingByFontSize(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "Title", Bold: true}}, Y: 720, X: 72, FontSize: 24},
		{Spans: []model.Span{{Text: "Body text goes here."}}, Y: 690, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "More body text."}}, Y: 675, X: 72, FontSize: 12},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Type != model.BlockHeading {
		t.Errorf("block[0].Type = %v, want BlockHeading", blocks[0].Type)
	}
	if blocks[0].Level != 1 {
		t.Errorf("block[0].Level = %d, want 1", blocks[0].Level)
	}
	if blocks[1].Type != model.BlockParagraph {
		t.Errorf("block[1].Type = %v, want BlockParagraph", blocks[1].Type)
	}
	if len(blocks[1].Lines) != 2 {
		t.Errorf("paragraph has %d lines, want 2", len(blocks[1].Lines))
	}
}

func TestClassifyBlocks_HeadingLevels(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "H1"}}, Y: 700, X: 72, FontSize: 24},   // >= 12*2.0
		{Spans: []model.Span{{Text: "H2"}}, Y: 680, X: 72, FontSize: 20.4}, // >= 12*1.7
		{Spans: []model.Span{{Text: "H3"}}, Y: 660, X: 72, FontSize: 16.8}, // >= 12*1.4
		{Spans: []model.Span{{Text: "H4"}}, Y: 640, X: 72, FontSize: 14.4}, // >= 12*1.2
	}

	blocks := ClassifyBlocks(lines, 12)

	expected := []int{1, 2, 3, 4}
	for i, want := range expected {
		if i >= len(blocks) {
			t.Fatalf("only got %d blocks, expected at least %d", len(blocks), i+1)
		}
		if blocks[i].Type != model.BlockHeading {
			t.Errorf("block[%d] not a heading", i)
			continue
		}
		if blocks[i].Level != want {
			t.Errorf("block[%d].Level = %d, want %d", i, blocks[i].Level, want)
		}
	}
}

func TestClassifyBlocks_ParagraphBreakByGap(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "First paragraph line 1."}}, Y: 700, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "First paragraph line 2."}}, Y: 686, X: 72, FontSize: 12},
		// Large gap (Y jump from 686 to 650 = 36pt gap, > 1.5*12=18)
		{Spans: []model.Span{{Text: "Second paragraph."}}, Y: 650, X: 72, FontSize: 12},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Type != model.BlockParagraph || blocks[1].Type != model.BlockParagraph {
		t.Error("both blocks should be paragraphs")
	}
	if len(blocks[0].Lines) != 2 {
		t.Errorf("first paragraph: %d lines, want 2", len(blocks[0].Lines))
	}
}

func TestClassifyBlocks_BoldBodyAsH5(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "Bold Heading", Bold: true}}, Y: 700, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "Normal body text follows."}}, Y: 680, X: 72, FontSize: 12},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Type != model.BlockHeading || blocks[0].Level != 5 {
		t.Errorf("block[0] = type %v level %d, want heading level 5", blocks[0].Type, blocks[0].Level)
	}
}

func TestClassifyBlocks_Empty(t *testing.T) {
	blocks := ClassifyBlocks(nil, 12)
	if len(blocks) != 0 {
		t.Errorf("got %d blocks for nil input, want 0", len(blocks))
	}
}

func TestClassifyBlocks_BulletList(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "Introduction."}}, Y: 720, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "- First item"}}, Y: 700, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "- Second item"}}, Y: 686, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "- Third item"}}, Y: 672, X: 72, FontSize: 12},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[1].Type != model.BlockList {
		t.Errorf("block[1].Type = %v, want BlockList", blocks[1].Type)
	}
	if blocks[1].Ordered {
		t.Error("bullet list should not be ordered")
	}
	if len(blocks[1].Lines) != 3 {
		t.Errorf("list has %d items, want 3", len(blocks[1].Lines))
	}
}

func TestClassifyBlocks_OrderedList(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "1. First"}}, Y: 700, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "2. Second"}}, Y: 686, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "3. Third"}}, Y: 672, X: 72, FontSize: 12},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].Type != model.BlockList {
		t.Errorf("type = %v, want BlockList", blocks[0].Type)
	}
	if !blocks[0].Ordered {
		t.Error("should be ordered")
	}
}

func TestClassifyBlocks_CodeBlock(t *testing.T) {
	lines := []model.Line{
		{Spans: []model.Span{{Text: "Some text."}}, Y: 720, X: 72, FontSize: 12},
		{Spans: []model.Span{{Text: "func main() {", Mono: true}}, Y: 700, X: 72, FontSize: 10},
		{Spans: []model.Span{{Text: "    fmt.Println()", Mono: true}}, Y: 688, X: 72, FontSize: 10},
		{Spans: []model.Span{{Text: "}", Mono: true}}, Y: 676, X: 72, FontSize: 10},
	}

	blocks := ClassifyBlocks(lines, 12)

	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[1].Type != model.BlockCode {
		t.Errorf("block[1].Type = %v, want BlockCode", blocks[1].Type)
	}
	if len(blocks[1].Lines) != 3 {
		t.Errorf("code block has %d lines, want 3", len(blocks[1].Lines))
	}
}
