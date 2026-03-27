package model

// TextElement is the raw unit from PDF extraction.
type TextElement struct {
	Text     string
	Font     string
	FontSize float64
	X, Y     float64
	W        float64 // width of the text run
	Bold     bool
	Italic   bool
	Mono     bool
}

// Rectangle represents a drawn rectangle on a PDF page.
type Rectangle struct {
	MinX, MinY, MaxX, MaxY float64
}

// ExtractedImage holds raw image data extracted from a page.
type ExtractedImage struct {
	PageNum int
	Index   int
	Data    []byte
	Format  string // "png", "jpg", "tiff"
	Y       float64 // top edge Y for placement ordering
}

// PageResult holds all extraction output for one page.
type PageResult struct {
	PageNum   int
	Texts     []TextElement
	Rects     []Rectangle
	Images    []ExtractedImage
	FontStats FontStats
}

// FontStats tracks frequency of font sizes and names by character count.
type FontStats struct {
	SizeCounts map[float64]int // fontSize → character count
	NameCounts map[string]int  // fontName → character count
}

// BlockType identifies the structural role of a block.
type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockList
	BlockTable
	BlockCode
	BlockImage
)

// Block is the structural unit fed to the markdown emitter.
type Block struct {
	Type     BlockType
	Level    int    // heading level 1-6, list nesting depth
	Ordered  bool   // true for ordered lists
	Lines    []Line
	Table    *Table
	ImageRef string // relative path for BlockImage
}

// Line is a reconstructed line of text with position metadata.
type Line struct {
	Spans    []Span
	Y        float64 // Y position (used by analysis)
	X        float64 // leftmost X position (used by analysis)
	FontSize float64 // dominant font size on this line
}

// Span is an inline run with consistent formatting.
type Span struct {
	Text   string
	Bold   bool
	Italic bool
	Mono   bool
}

// Table holds parsed table data.
type Table struct {
	Headers []string
	Rows    [][]string
	Aligns  []Align
}

// Align represents column alignment.
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)
