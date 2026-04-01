package analyze

import (
	"regexp"
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

var orderedListRe = regexp.MustCompile(`^\d+[.)]\s`)

// isListItem returns true when the line looks like a list item.
// Detects bullet prefixes (-, *, unicode bullets) or ordered list pattern (1. / 1)).
func isListItem(line model.Line) bool {
	if len(line.Spans) == 0 {
		return false
	}
	text := strings.TrimSpace(line.Spans[0].Text)
	if len(text) == 0 {
		return false
	}
	// Bullet list prefixes
	switch text[0] {
	case '-', '*':
		return len(text) > 1 && (text[1] == ' ' || text[1] == '\t')
	case '\xe2': // UTF-8 lead byte for many unicode bullets (•, ‣, ▪, etc.)
		return true
	}
	// Unicode bullet • (U+2022) encoded in UTF-8 is 0xE2 0x80 0xA2
	if strings.HasPrefix(text, "•") || strings.HasPrefix(text, "‣") || strings.HasPrefix(text, "▪") {
		return true
	}
	// Ordered list: digits followed by . or )
	return orderedListRe.MatchString(text)
}

// consumeList collects a run of list-item lines into a Block.
// Continuation lines (indented, not themselves a list item) are included.
// Stops when a non-list, non-continuation line is encountered.
func consumeList(lines []model.Line, start int, bodyFontSize float64) (model.Block, int) {
	first := lines[start]
	firstText := strings.TrimSpace(lineText(first))
	ordered := orderedListRe.MatchString(firstText)

	block := model.Block{
		Type:    model.BlockList,
		Ordered: ordered,
		Lines:   []model.Line{first},
	}

	i := start + 1
	for i < len(lines) {
		line := lines[i]
		if isListItem(line) {
			block.Lines = append(block.Lines, line)
			i++
			continue
		}
		// Continuation: indented relative to the list start and not a heading/code
		if line.X > first.X && !allMono(line) {
			block.Lines = append(block.Lines, line)
			i++
			continue
		}
		break
	}

	return block, i - start
}

// consumeCode collects a run of monospaced lines into a code Block.
func consumeCode(lines []model.Line, start int) (model.Block, int) {
	block := model.Block{
		Type:  model.BlockCode,
		Lines: []model.Line{lines[start]},
	}

	i := start + 1
	for i < len(lines) && allMono(lines[i]) {
		block.Lines = append(block.Lines, lines[i])
		i++
	}

	return block, i - start
}
