package emit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func TestEmit_Heading(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockHeading, Level: 1, Lines: []model.Line{
			{Spans: []model.Span{{Text: "Title"}}},
		}},
		{Type: model.BlockHeading, Level: 3, Lines: []model.Line{
			{Spans: []model.Span{{Text: "Subsection"}}},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	want := "# Title\n\n### Subsection\n\n"
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestEmit_Paragraph(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockParagraph, Lines: []model.Line{
			{Spans: []model.Span{{Text: "First line."}}},
			{Spans: []model.Span{{Text: "Second line."}}},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	want := "First line.\nSecond line.\n\n"
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestEmit_InlineFormatting(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockParagraph, Lines: []model.Line{
			{Spans: []model.Span{
				{Text: "Normal "},
				{Text: "bold", Bold: true},
				{Text: " and "},
				{Text: "italic", Italic: true},
				{Text: " and "},
				{Text: "code", Mono: true},
			}},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "**bold**") {
		t.Errorf("missing bold formatting in: %q", got)
	}
	if !strings.Contains(got, "*italic*") {
		t.Errorf("missing italic formatting in: %q", got)
	}
	if !strings.Contains(got, "`code`") {
		t.Errorf("missing code formatting in: %q", got)
	}
}

func TestEmit_BoldItalic(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockParagraph, Lines: []model.Line{
			{Spans: []model.Span{
				{Text: "both", Bold: true, Italic: true},
			}},
		}},
	}

	var buf bytes.Buffer
	Emit(blocks, &buf)

	if !strings.Contains(buf.String(), "***both***") {
		t.Errorf("missing bold+italic: %q", buf.String())
	}
}

func TestEmit_Table(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockTable, Table: &model.Table{
			Headers: []string{"Name", "Age"},
			Rows:    [][]string{{"Alice", "30"}, {"Bob", "25"}},
			Aligns:  []model.Align{model.AlignLeft, model.AlignRight},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "| Name") {
		t.Errorf("missing header: %s", got)
	}
	if !strings.Contains(got, "| Alice") {
		t.Errorf("missing data row: %s", got)
	}
	if !strings.Contains(got, "---:") {
		t.Errorf("missing right-align separator: %s", got)
	}
}

func TestEmit_List(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockList, Ordered: false, Lines: []model.Line{
			{Spans: []model.Span{{Text: "- First"}}},
			{Spans: []model.Span{{Text: "- Second"}}},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	want := "- First\n- Second\n\n"
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestEmit_Code(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockCode, Lines: []model.Line{
			{Spans: []model.Span{{Text: "func main() {", Mono: true}}},
			{Spans: []model.Span{{Text: "    fmt.Println()", Mono: true}}},
			{Spans: []model.Span{{Text: "}", Mono: true}}},
		}},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	want := "```\nfunc main() {\n    fmt.Println()\n}\n```\n\n"
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestEmit_Image(t *testing.T) {
	blocks := []model.Block{
		{Type: model.BlockImage, ImageRef: "report_p1_001.png"},
	}

	var buf bytes.Buffer
	if err := Emit(blocks, &buf); err != nil {
		t.Fatal(err)
	}

	want := "![](report_p1_001.png)\n\n"
	if buf.String() != want {
		t.Errorf("got: %q, want: %q", buf.String(), want)
	}
}
