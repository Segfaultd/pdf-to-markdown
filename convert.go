package pdf2md

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/segfaultd/pdf-to-markdown/pkg/analyze"
	"github.com/segfaultd/pdf-to-markdown/pkg/emit"
	"github.com/segfaultd/pdf-to-markdown/pkg/extract"
	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// Options configures the conversion pipeline.
type Options struct {
	Workers     int
	ImageDir    string
	PageBreak   bool
	ImageWriter ImageWriter
	Quiet       bool
}

// ImageWriter is implemented by callers that want to handle image persistence.
type ImageWriter interface {
	WriteImage(name string, data []byte, format string) (refPath string, err error)
}

// Convert reads a PDF from r and writes Markdown to w.
// opts may be nil to use defaults.
func Convert(r io.ReadSeeker, w io.Writer, opts *Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading PDF: %w", err)
	}
	return convertFromBytes(data, "", w, opts)
}

// ConvertFile reads the PDF at pdfPath and writes Markdown to mdPath.
// opts may be nil to use defaults.
func ConvertFile(pdfPath, mdPath string, opts *Options) error {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("reading PDF file: %w", err)
	}

	out, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer out.Close()

	base := filepath.Base(pdfPath)
	// Strip extension for use in image filenames.
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}

	return convertFromBytes(data, base, out, opts)
}

// ConvertBytes converts a PDF supplied as raw bytes and returns Markdown bytes.
func ConvertBytes(pdf []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := convertFromBytes(pdf, "", &buf, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// convertFromBytes is the shared pipeline implementation.
// baseName is used as a prefix for image filenames; empty string means no PDF
// filename is available (Convert / ConvertBytes).
func convertFromBytes(data []byte, baseName string, w io.Writer, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}

	ra := bytes.NewReader(data)
	size := int64(len(data))

	// Stage 1: extract text from all pages.
	pages, err := extract.ExtractReaderPages(ra, size)
	if err != nil {
		return fmt.Errorf("extracting pages: %w", err)
	}

	// Stage 1b: extract images.
	images, err := extract.ExtractReaderImages(data)
	if err != nil {
		// Non-fatal: some PDFs produce errors here; continue without images.
		images = nil
	}

	// Stage 2: merge font stats across all pages and compute body font size.
	allStats := make([]model.FontStats, len(pages))
	for i, p := range pages {
		allStats[i] = p.FontStats
	}
	merged := analyze.MergeFontStats(allStats)
	bodyFontSize := analyze.BodyFontSize(merged)

	// Index images by page number for O(1) lookup.
	imagesByPage := make(map[int][]model.ExtractedImage)
	for _, img := range images {
		imagesByPage[img.PageNum] = append(imagesByPage[img.PageNum], img)
	}

	// Stage 3: per-page processing.
	for _, page := range pages {
		// 3a. Detect tables and get the indices of consumed text elements.
		tables, consumedIdxs := analyze.DetectTables(page.Texts, page.Rects)

		// 3b. Filter out consumed text elements.
		consumedSet := make(map[int]bool, len(consumedIdxs))
		for _, idx := range consumedIdxs {
			consumedSet[idx] = true
		}
		remaining := make([]model.TextElement, 0, len(page.Texts)-len(consumedIdxs))
		for i, te := range page.Texts {
			if !consumedSet[i] {
				remaining = append(remaining, te)
			}
		}

		// 3c. Reconstruct lines from remaining text.
		lines := analyze.ReconstructLines(remaining)

		// 3d. Classify blocks.
		blocks := analyze.ClassifyBlocks(lines, bodyFontSize)

		// 3e. Insert table blocks at the appropriate position (by Y coordinate).
		for _, tr := range tables {
			tableBlock := model.Block{
				Type:  model.BlockTable,
				Table: tr.Table,
			}
			blocks = insertBlockByY(blocks, tableBlock, tr.MaxY)
		}

		// 3f. Insert image blocks for this page.
		pageImages := imagesByPage[page.PageNum]
		// Sort images by Y so they appear in document order.
		sort.Slice(pageImages, func(i, j int) bool {
			return pageImages[i].Y > pageImages[j].Y
		})
		for imgIdx, img := range pageImages {
			ref := imageFilename(baseName, page.PageNum, imgIdx, img.Format)

			if opts.ImageWriter != nil {
				refPath, werr := opts.ImageWriter.WriteImage(ref, img.Data, img.Format)
				if werr == nil {
					ref = refPath
				}
			} else if opts.ImageDir != "" {
				dest := filepath.Join(opts.ImageDir, ref)
				_ = os.WriteFile(dest, img.Data, 0o644)
			}

			imageBlock := model.Block{
				Type:     model.BlockImage,
				ImageRef: ref,
			}
			blocks = insertBlockByY(blocks, imageBlock, img.Y)
		}

		// 3g. Emit blocks.
		if err := emit.Emit(blocks, w); err != nil {
			return fmt.Errorf("emitting page %d: %w", page.PageNum, err)
		}

		// Optionally write a page break.
		if opts.PageBreak {
			if _, err := fmt.Fprintf(w, "\n---\n\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// imageFilename constructs a filename for an extracted image.
// baseName is empty when no source filename is known.
func imageFilename(baseName string, pageNum, index int, format string) string {
	if baseName == "" {
		return fmt.Sprintf("image_p%d_%d.%s", pageNum, index, format)
	}
	return fmt.Sprintf("%s_p%d_%d.%s", baseName, pageNum, index, format)
}

// insertBlockByY inserts newBlock into blocks at the position that keeps the
// list ordered by the Y position of each block's first line.  For non-text
// blocks (tables, images) the provided refY is used as the sort key.
// This is a best-effort ordering; if no suitable position is found the block
// is appended at the end.
func insertBlockByY(blocks []model.Block, newBlock model.Block, refY float64) []model.Block {
	// Find the first existing block whose Y is less than refY (i.e. below the
	// new block in PDF coordinates where Y grows upward).
	for i, b := range blocks {
		if blockY(b) < refY {
			// Insert before position i.
			result := make([]model.Block, 0, len(blocks)+1)
			result = append(result, blocks[:i]...)
			result = append(result, newBlock)
			result = append(result, blocks[i:]...)
			return result
		}
	}
	return append(blocks, newBlock)
}

// blockY returns the Y coordinate of the first line of a block, or 0 if there
// are no lines (e.g. table / image blocks that carry no Lines).
func blockY(b model.Block) float64 {
	if len(b.Lines) > 0 {
		return b.Lines[0].Y
	}
	return 0
}
