package analyze

import (
	"math"
	"slices"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// ReconstructLines groups TextElements into lines by Y position,
// sorts left-to-right, and merges fragments into spans.
func ReconstructLines(texts []model.TextElement) []model.Line {
	if len(texts) == 0 {
		return nil
	}

	// Copy to avoid mutating the caller's slice.
	sorted := make([]model.TextElement, len(texts))
	copy(sorted, texts)

	// Stable sort by Y descending. Stable preserves document order for
	// elements at the same Y — critical for per-glyph PDFs where the
	// content-stream order IS the reading order.
	slices.SortStableFunc(sorted, func(a, b model.TextElement) int {
		if a.Y > b.Y {
			return -1
		}
		if a.Y < b.Y {
			return 1
		}
		return 0
	})

	// Bucket into lines by Y proximity
	buckets := [][]model.TextElement{{sorted[0]}}
	for i := 1; i < len(sorted); i++ {
		cur := sorted[i]
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
	// Cluster elements by X proximity, preserving document order within
	// each cluster. This handles PDFs that report per-glyph elements
	// with W=0 and nearly identical X positions — pure X-sort would
	// scramble them because the tiny X differences aren't monotonic.
	//
	// For normal PDFs (proper W values, distinct X per element) each
	// element becomes its own cluster and the cluster sort is equivalent
	// to a regular X-sort.
	clusters := clusterByX(elems)

	// Sort clusters left-to-right by their reference X.
	slices.SortFunc(clusters, func(a, b []model.TextElement) int {
		if a[0].X < b[0].X {
			return -1
		}
		if a[0].X > b[0].X {
			return 1
		}
		return 0
	})

	// Flatten clusters back into ordered elements.
	ordered := make([]model.TextElement, 0, len(elems))
	for _, c := range clusters {
		ordered = append(ordered, c...)
	}

	// Build spans from ordered elements.
	var spans []model.Span
	var maxFontSize float64

	for i, elem := range ordered {
		if elem.FontSize > maxFontSize {
			maxFontSize = elem.FontSize
		}

		text := elem.Text

		if i > 0 {
			prev := ordered[i-1]
			gap := elem.X - (prev.X + prev.W)
			// For W=0 glyphs in the same cluster, gap ≈ 0 → no space.
			// Between clusters the X jump is large → space inserted.
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
		Y:        ordered[0].Y,
		X:        ordered[0].X,
		FontSize: maxFontSize,
	}
}

// clusterByX groups elements that are at nearly the same X position,
// preserving their original (document) order within each cluster.
// Elements whose X is within threshold of the cluster's reference X
// stay together; a larger jump starts a new cluster.
func clusterByX(elems []model.TextElement) [][]model.TextElement {
	if len(elems) == 0 {
		return nil
	}

	// Threshold: half the dominant font size, minimum 2pt.
	fontSize := elems[0].FontSize
	for _, e := range elems {
		if e.FontSize > fontSize {
			fontSize = e.FontSize
		}
	}
	threshold := fontSize * 0.5
	if threshold < 2 {
		threshold = 2
	}

	var clusters [][]model.TextElement
	current := []model.TextElement{elems[0]}
	refX := elems[0].X

	for i := 1; i < len(elems); i++ {
		if math.Abs(elems[i].X-refX) <= threshold {
			current = append(current, elems[i])
		} else {
			clusters = append(clusters, current)
			current = []model.TextElement{elems[i]}
			refX = elems[i].X
		}
	}
	clusters = append(clusters, current)
	return clusters
}
