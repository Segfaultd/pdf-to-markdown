package analyze

import (
	"slices"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// ReconstructLines groups TextElements into lines by Y position,
// sorts left-to-right, and merges fragments into spans.
func ReconstructLines(texts []model.TextElement) []model.Line {
	if len(texts) == 0 {
		return nil
	}

	// Sort by Y descending (top of page first in PDF coordinates)
	slices.SortFunc(texts, func(a, b model.TextElement) int {
		if a.Y > b.Y {
			return -1
		}
		if a.Y < b.Y {
			return 1
		}
		return 0
	})

	// Bucket into lines by Y proximity
	buckets := [][]model.TextElement{{texts[0]}}
	for i := 1; i < len(texts); i++ {
		cur := texts[i]
		lastBucket := buckets[len(buckets)-1]
		refY := lastBucket[0].Y
		tolerance := lastBucket[0].FontSize * 0.5
		if tolerance < 1 {
			tolerance = 1
		}

		if refY-cur.Y <= tolerance {
			buckets[len(buckets)-1] = append(lastBucket, cur)
		} else {
			buckets = append(buckets, []model.TextElement{cur})
		}
	}

	// Process each bucket into a Line
	lines := make([]model.Line, 0, len(buckets))
	for _, bucket := range buckets {
		line := buildLine(bucket)
		lines = append(lines, line)
	}

	return lines
}

func buildLine(elems []model.TextElement) model.Line {
	// Sort left-to-right by X
	slices.SortFunc(elems, func(a, b model.TextElement) int {
		if a.X < b.X {
			return -1
		}
		if a.X > b.X {
			return 1
		}
		return 0
	})

	var spans []model.Span
	var maxFontSize float64

	for i, elem := range elems {
		if elem.FontSize > maxFontSize {
			maxFontSize = elem.FontSize
		}

		text := elem.Text

		// Check gap from previous element
		if i > 0 {
			prev := elems[i-1]
			gap := elem.X - (prev.X + prev.W)
			needsSpace := gap >= prev.FontSize*0.3
			sameStyle := elem.Bold == prev.Bold && elem.Italic == prev.Italic && elem.Mono == prev.Mono

			if sameStyle && len(spans) > 0 {
				last := &spans[len(spans)-1]
				if needsSpace {
					last.Text += " " + text
				} else {
					last.Text += text
				}
				continue
			}

			// Different style: prepend space to new span if needed
			if needsSpace {
				text = " " + text
			}
		}

		spans = append(spans, model.Span{
			Text:   text,
			Bold:   elem.Bold,
			Italic: elem.Italic,
			Mono:   elem.Mono,
		})
	}

	return model.Line{
		Spans:    spans,
		Y:        elems[0].Y,
		X:        elems[0].X,
		FontSize: maxFontSize,
	}
}
