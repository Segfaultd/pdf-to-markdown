package emit

import (
	"fmt"
	"io"
	"strings"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// Emit writes blocks as Markdown to w.
func Emit(blocks []model.Block, w io.Writer) error {
	for _, b := range blocks {
		var err error
		switch b.Type {
		case model.BlockHeading:
			err = emitHeading(b, w)
		case model.BlockParagraph:
			err = emitParagraph(b, w)
		case model.BlockList:
			err = emitList(b, w)
		case model.BlockTable:
			err = emitTable(b, w)
		case model.BlockCode:
			err = emitCode(b, w)
		case model.BlockImage:
			err = emitImage(b, w)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// emitHeading writes: #{level} <line text>\n\n
func emitHeading(b model.Block, w io.Writer) error {
	prefix := strings.Repeat("#", b.Level)
	for _, line := range b.Lines {
		if _, err := fmt.Fprintf(w, "%s %s\n\n", prefix, renderLine(line)); err != nil {
			return err
		}
	}
	return nil
}

// emitParagraph writes each line followed by \n, then an extra \n after the block.
func emitParagraph(b model.Block, w io.Writer) error {
	for _, line := range b.Lines {
		if _, err := fmt.Fprintf(w, "%s\n", renderLine(line)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}

// emitList writes each line as-is (bullet/number prefix already in text) + \n, then \n.
func emitList(b model.Block, w io.Writer) error {
	for _, line := range b.Lines {
		if _, err := fmt.Fprintf(w, "%s\n", renderLine(line)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}

// emitCode writes a fenced code block using raw (unformatted) line text.
func emitCode(b model.Block, w io.Writer) error {
	if _, err := fmt.Fprint(w, "```\n"); err != nil {
		return err
	}
	for _, line := range b.Lines {
		if _, err := fmt.Fprintf(w, "%s\n", rawLineText(line)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "```\n\n")
	return err
}

// emitImage writes: ![](images/{ref})\n\n
func emitImage(b model.Block, w io.Writer) error {
	_, err := fmt.Fprintf(w, "![](images/%s)\n\n", b.ImageRef)
	return err
}

// emitTable writes a GFM table with padded columns and alignment separators.
func emitTable(b model.Block, w io.Writer) error {
	if b.Table == nil {
		return nil
	}
	t := b.Table
	cols := len(t.Headers)
	if cols == 0 {
		return nil
	}

	// Compute max width per column (header vs each row cell).
	widths := make([]int, cols)
	for i, h := range t.Headers {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range t.Rows {
		for i := 0; i < cols && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	// Header row.
	if err := writeTableRow(w, t.Headers, widths); err != nil {
		return err
	}

	// Separator row.
	seps := make([]string, cols)
	for i := 0; i < cols; i++ {
		align := model.AlignLeft
		if i < len(t.Aligns) {
			align = t.Aligns[i]
		}
		dashes := strings.Repeat("-", widths[i])
		switch align {
		case model.AlignLeft:
			seps[i] = ":" + dashes
		case model.AlignRight:
			seps[i] = dashes + ":"
		case model.AlignCenter:
			seps[i] = ":" + dashes + ":"
		default:
			seps[i] = dashes
		}
	}
	if err := writeTableRow(w, seps, nil); err != nil {
		return err
	}

	// Data rows.
	for _, row := range t.Rows {
		// Pad short rows to cols.
		padded := make([]string, cols)
		copy(padded, row)
		if err := writeTableRow(w, padded, widths); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "\n")
	return err
}

// writeTableRow writes a single pipe-delimited row. When widths is non-nil cells
// are padded to the given widths; when widths is nil cells are written as-is
// (used for the separator row where dashes already encode alignment).
func writeTableRow(w io.Writer, cells []string, widths []int) error {
	var sb strings.Builder
	sb.WriteString("|")
	for i, c := range cells {
		sb.WriteString(" ")
		sb.WriteString(c)
		if widths != nil && i < len(widths) {
			// Right-pad with spaces.
			sb.WriteString(strings.Repeat(" ", widths[i]-len(c)))
		}
		sb.WriteString(" |")
	}
	sb.WriteString("\n")
	_, err := fmt.Fprint(w, sb.String())
	return err
}

// renderLine concatenates all spans in a line with inline Markdown formatting.
func renderLine(line model.Line) string {
	var sb strings.Builder
	for _, s := range line.Spans {
		sb.WriteString(renderSpan(s))
	}
	return sb.String()
}

// renderSpan applies bold/italic/mono markers.
func renderSpan(s model.Span) string {
	switch {
	case s.Bold && s.Italic:
		return "***" + s.Text + "***"
	case s.Bold:
		return "**" + s.Text + "**"
	case s.Italic:
		return "*" + s.Text + "*"
	case s.Mono:
		return "`" + s.Text + "`"
	default:
		return s.Text
	}
}

// rawLineText concatenates span text without any formatting markers.
func rawLineText(line model.Line) string {
	var sb strings.Builder
	for _, s := range line.Spans {
		sb.WriteString(s.Text)
	}
	return sb.String()
}
