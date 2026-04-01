package analyze

import (
	"strings"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// ClassifyBlocks takes reconstructed lines and a body font size and returns
// a slice of Blocks with type and level assigned.
func ClassifyBlocks(lines []model.Line, bodyFontSize float64) []model.Block {
	if len(lines) == 0 {
		return nil
	}

	var blocks []model.Block
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Check for code block (monospaced lines)
		if allMono(line) {
			block, consumed := consumeCode(lines, i)
			blocks = append(blocks, block)
			i += consumed
			continue
		}

		// Check for list item
		if isListItem(line) {
			block, consumed := consumeList(lines, i, bodyFontSize)
			blocks = append(blocks, block)
			i += consumed
			continue
		}

		// Check for heading by font size
		if level := headingLevel(line.FontSize, bodyFontSize); level > 0 {
			blocks = append(blocks, model.Block{
				Type:  model.BlockHeading,
				Level: level,
				Lines: []model.Line{line},
			})
			i++
			continue
		}

		// Check for bold body-size line as H5
		if isH5Candidate(line, bodyFontSize) && i+1 < len(lines) && !lineAllBold(lines[i+1]) {
			blocks = append(blocks, model.Block{
				Type:  model.BlockHeading,
				Level: 5,
				Lines: []model.Line{line},
			})
			i++
			continue
		}

		// Accumulate paragraph lines
		para := model.Block{
			Type:  model.BlockParagraph,
			Lines: []model.Line{line},
		}
		i++

		for i < len(lines) {
			next := lines[i]

			// Stop if next line is a heading (by font size)
			if headingLevel(next.FontSize, bodyFontSize) > 0 {
				break
			}

			// Stop if next line is a bold body-size H5 candidate
			if isH5Candidate(next, bodyFontSize) {
				break
			}

			// Stop if next line is a list item or code
			if isListItem(next) || allMono(next) {
				break
			}

			// Stop on large Y gap
			prev := lines[i-1]
			gap := prev.Y - next.Y
			if gap > 1.5*bodyFontSize {
				break
			}

			para.Lines = append(para.Lines, next)
			i++
		}

		blocks = append(blocks, para)
	}

	return blocks
}

// headingLevel returns 1-4 based on font size ratio, or 0 if not a heading.
func headingLevel(fontSize, bodyFontSize float64) int {
	ratio := fontSize / bodyFontSize
	switch {
	case ratio >= 2.0:
		return 1
	case ratio >= 1.7:
		return 2
	case ratio >= 1.4:
		return 3
	case ratio >= 1.2:
		return 4
	default:
		return 0
	}
}

// isH5Candidate returns true for a body-size line where all spans are bold and
// the text is short (a heading-like label rather than body prose).
func isH5Candidate(line model.Line, bodyFontSize float64) bool {
	if line.FontSize > bodyFontSize*1.1 {
		return false
	}
	if !lineAllBold(line) {
		return false
	}
	// Short line heuristic: fewer than 80 characters
	return len(strings.TrimSpace(lineText(line))) < 80
}

// lineAllBold returns true when every span in the line is bold.
func lineAllBold(line model.Line) bool {
	if len(line.Spans) == 0 {
		return false
	}
	for _, s := range line.Spans {
		if !s.Bold {
			return false
		}
	}
	return true
}

// lineText concatenates all span texts for a line.
func lineText(line model.Line) string {
	var sb strings.Builder
	for _, s := range line.Spans {
		sb.WriteString(s.Text)
	}
	return sb.String()
}

// allMono returns true when every span in the line uses monospace font.
func allMono(line model.Line) bool {
	if len(line.Spans) == 0 {
		return false
	}
	for _, s := range line.Spans {
		if !s.Mono {
			return false
		}
	}
	return true
}

// isListItem returns true when the line looks like a list item.
// Stub — filled in by Task 6.
func isListItem(line model.Line) bool { return false }

// consumeList collects a run of list-item lines into a Block.
// Stub — filled in by Task 6.
func consumeList(lines []model.Line, start int, bodyFontSize float64) (model.Block, int) {
	return model.Block{}, 1
}

// consumeCode collects a run of monospaced lines into a code Block.
// Stub — filled in by Task 6.
func consumeCode(lines []model.Line, start int) (model.Block, int) { return model.Block{}, 1 }
