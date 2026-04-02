package analyze

import (
	"math"
	"slices"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// TableRegion holds a detected table along with its vertical extent.
type TableRegion struct {
	Table *model.Table
	MinY  float64
	MaxY  float64
}

// DetectTables attempts to find tables in the given text elements.
//
// Strategy 1 (rectangle-based): when >= 4 rectangles are provided, look for a
// grid of cell rectangles and map text into cells.
//
// Strategy 2 (column alignment): when no rectangles exist, look for >= 3 lines
// that each have the same number (>= 2) of elements at consistent X positions.
//
// Returns the detected TableRegions and a slice of indices into texts that
// belong to a table (so the caller can filter them out).
func DetectTables(texts []model.TextElement, rects []model.Rectangle) ([]TableRegion, []int) {
	if len(rects) >= 4 {
		return detectRectTables(texts, rects)
	}
	return detectAlignedTables(texts)
}

// ─── Strategy 1: rectangle-based ─────────────────────────────────────────────

func detectRectTables(texts []model.TextElement, rects []model.Rectangle) ([]TableRegion, []int) {
	const edgeTol = 2.0

	// Bucket rectangle top/bottom edges to find rows, and left/right edges to
	// find columns.
	rowEdges := bucketEdges(func() []float64 {
		es := make([]float64, 0, len(rects)*2)
		for _, r := range rects {
			es = append(es, r.MinY, r.MaxY)
		}
		return es
	}(), edgeTol)

	colEdges := bucketEdges(func() []float64 {
		es := make([]float64, 0, len(rects)*2)
		for _, r := range rects {
			es = append(es, r.MinX, r.MaxX)
		}
		return es
	}(), edgeTol)

	nRows := len(rowEdges) - 1
	nCols := len(colEdges) - 1

	if nRows < 1 || nCols < 2 {
		return nil, nil
	}

	// Sort edges descending for rows (PDF Y increases upward) and ascending for cols.
	slices.SortFunc(rowEdges, func(a, b float64) int {
		if a > b {
			return -1
		}
		if a < b {
			return 1
		}
		return 0
	})
	slices.Sort(colEdges)

	// Build a grid: grid[row][col] = cell text
	grid := make([][]string, nRows)
	for i := range grid {
		grid[i] = make([]string, nCols)
	}

	consumedSet := make(map[int]bool)

	for ti, te := range texts {
		r, c := cellFor(te, rowEdges, colEdges, edgeTol)
		if r < 0 || c < 0 {
			continue
		}
		if grid[r][c] != "" {
			grid[r][c] += " " + te.Text
		} else {
			grid[r][c] = te.Text
		}
		consumedSet[ti] = true
	}

	if len(consumedSet) == 0 {
		return nil, nil
	}

	// Build Table: first row = headers.
	headers := grid[0]
	rows := grid[1:]

	tbl := &model.Table{
		Headers: headers,
		Rows:    rows,
		Aligns:  make([]model.Align, nCols),
	}

	minY := rowEdges[len(rowEdges)-1]
	maxY := rowEdges[0]

	region := TableRegion{Table: tbl, MinY: minY, MaxY: maxY}

	consumed := setToSlice(consumedSet)
	return []TableRegion{region}, consumed
}

// bucketEdges collapses nearby edge values into representative buckets and
// returns the unique bucket values sorted ascending.
func bucketEdges(vals []float64, tol float64) []float64 {
	slices.Sort(vals)
	var buckets []float64
	for _, v := range vals {
		found := false
		for i, b := range buckets {
			if math.Abs(v-b) <= tol {
				// Average into bucket
				buckets[i] = (b + v) / 2
				found = true
				break
			}
		}
		if !found {
			buckets = append(buckets, v)
		}
	}
	return buckets
}

// cellFor returns the (row, col) grid indices for a text element, or (-1,-1)
// if the element doesn't fall inside any cell.
func cellFor(te model.TextElement, rowEdges, colEdges []float64, tol float64) (int, int) {
	row := -1
	for i := 0; i < len(rowEdges)-1; i++ {
		top := rowEdges[i]
		bot := rowEdges[i+1]
		if te.Y <= top+tol && te.Y >= bot-tol {
			row = i
			break
		}
	}
	if row < 0 {
		return -1, -1
	}

	col := -1
	for j := 0; j < len(colEdges)-1; j++ {
		left := colEdges[j]
		right := colEdges[j+1]
		if te.X >= left-tol && te.X < right+tol {
			col = j
			break
		}
	}
	return row, col
}

// ─── Strategy 2: column alignment ────────────────────────────────────────────

const yLineTol = 4.0  // pt tolerance for same-line grouping
const xAlignTol = 5.0 // pt tolerance for column X alignment
const minColCount = 2  // minimum columns for a table line
const minLineCount = 3 // minimum lines (including header) for a table

func detectAlignedTables(texts []model.TextElement) ([]TableRegion, []int) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Group text elements into Y-lines.
	yLines := groupByY(texts, yLineTol)

	// Find runs of lines that look like table rows.
	var tables []TableRegion
	var consumed []int

	i := 0
	for i < len(yLines) {
		run, colCount := longestAlignedRun(yLines, i)
		if run < minLineCount || colCount < minColCount {
			i++
			continue
		}

		// We have a valid table from yLines[i] to yLines[i+run-1].
		tableLines := yLines[i : i+run]
		tbl, idxs := buildAlignedTable(tableLines, colCount, texts)
		if tbl == nil {
			i++
			continue
		}

		maxY := tableLines[0][0].Y
		minY := tableLines[len(tableLines)-1][0].Y

		tables = append(tables, TableRegion{Table: tbl, MinY: minY, MaxY: maxY})
		consumed = append(consumed, idxs...)
		i += run
	}

	return tables, consumed
}

// groupByY groups text elements into lines by proximity of their Y coordinates.
// Returns slices of elements per line, sorted top-to-bottom (descending Y).
func groupByY(texts []model.TextElement, tol float64) [][]model.TextElement {
	// Sort by Y descending
	sorted := make([]model.TextElement, len(texts))
	copy(sorted, texts)
	slices.SortFunc(sorted, func(a, b model.TextElement) int {
		if a.Y > b.Y {
			return -1
		}
		if a.Y < b.Y {
			return 1
		}
		return 0
	})

	var lines [][]model.TextElement
	for _, te := range sorted {
		if len(lines) == 0 {
			lines = append(lines, []model.TextElement{te})
			continue
		}
		last := lines[len(lines)-1]
		refY := last[0].Y
		if math.Abs(te.Y-refY) <= tol {
			lines[len(lines)-1] = append(last, te)
		} else {
			lines = append(lines, []model.TextElement{te})
		}
	}

	// Sort each line left-to-right
	for li := range lines {
		slices.SortFunc(lines[li], func(a, b model.TextElement) int {
			if a.X < b.X {
				return -1
			}
			if a.X > b.X {
				return 1
			}
			return 0
		})
	}

	return lines
}

// longestAlignedRun returns how many consecutive lines starting at yLines[start]
// form a column-aligned run, plus the column count of that run.
func longestAlignedRun(yLines [][]model.TextElement, start int) (int, int) {
	if start >= len(yLines) {
		return 0, 0
	}

	firstLine := yLines[start]
	colCount := len(firstLine)
	if colCount < minColCount {
		return 0, 0
	}

	// Derive expected X positions from the first line
	xPositions := make([]float64, colCount)
	for j, te := range firstLine {
		xPositions[j] = te.X
	}

	count := 1
	for i := start + 1; i < len(yLines); i++ {
		line := yLines[i]
		if len(line) != colCount {
			break
		}
		if !xAligned(line, xPositions) {
			break
		}
		count++
	}

	if count < minLineCount {
		return 0, 0
	}
	return count, colCount
}

// xAligned returns true when each element in line has its X within xAlignTol
// of the corresponding expected X position.
func xAligned(line []model.TextElement, xPositions []float64) bool {
	if len(line) != len(xPositions) {
		return false
	}
	for j, te := range line {
		if math.Abs(te.X-xPositions[j]) > xAlignTol {
			return false
		}
	}
	return true
}

// buildAlignedTable constructs a Table from aligned yLines and returns the
// original text element indices that were consumed.
func buildAlignedTable(tableLines [][]model.TextElement, colCount int, allTexts []model.TextElement) (*model.Table, []int) {
	// Determine alignment per column by comparing consistency of left vs right edges.
	aligns := make([]model.Align, colCount)
	for j := 0; j < colCount; j++ {
		leftVars := columnVariance(tableLines, j, false)
		rightVars := columnVariance(tableLines, j, true)
		if rightVars < leftVars {
			aligns[j] = model.AlignRight
		} else {
			aligns[j] = model.AlignLeft
		}
	}

	// First line = headers, rest = data rows.
	headers := make([]string, colCount)
	for j, te := range tableLines[0] {
		headers[j] = te.Text
	}

	rows := make([][]string, len(tableLines)-1)
	for i, line := range tableLines[1:] {
		row := make([]string, colCount)
		for j, te := range line {
			row[j] = te.Text
		}
		rows[i] = row
	}

	tbl := &model.Table{
		Headers: headers,
		Rows:    rows,
		Aligns:  aligns,
	}

	// Build consumed index list by matching table elements back to allTexts.
	consumedSet := make(map[int]bool)
	for _, line := range tableLines {
		for _, te := range line {
			for idx, orig := range allTexts {
				if orig.Text == te.Text && orig.X == te.X && orig.Y == te.Y {
					consumedSet[idx] = true
					break
				}
			}
		}
	}

	return tbl, setToSlice(consumedSet)
}

// columnVariance computes the variance of X (or X+W when rightEdge=true) for
// column j across all table lines.
func columnVariance(lines [][]model.TextElement, col int, rightEdge bool) float64 {
	if len(lines) == 0 {
		return 0
	}
	var sum float64
	for _, line := range lines {
		if col >= len(line) {
			continue
		}
		te := line[col]
		x := te.X
		if rightEdge {
			x = te.X + te.W
		}
		sum += x
	}
	mean := sum / float64(len(lines))

	var variance float64
	for _, line := range lines {
		if col >= len(line) {
			continue
		}
		te := line[col]
		x := te.X
		if rightEdge {
			x = te.X + te.W
		}
		d := x - mean
		variance += d * d
	}
	return variance / float64(len(lines))
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func setToSlice(s map[int]bool) []int {
	out := make([]int, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
