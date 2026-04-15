# pdf2md

A high-performance PDF to Markdown converter written in Go. Extracts text, tables, and images from PDF documents and produces well-structured Markdown output.

## Features

- **Structural analysis** -- detects headings, paragraphs, lists, code blocks, and tables from font size, style, and spatial layout
- **Table detection** -- rectangle-based (bordered tables) and column-alignment heuristic (borderless tables)
- **Image extraction** -- extracts embedded images to files and inserts Markdown references
- **Inline formatting** -- preserves bold, italic, and monospace styling
- **Fast** -- processes ~2.7ms per page on Apple Silicon; a 30-page mixed PDF converts in under 80ms
- **Pure Go** -- no CGO, easy cross-compilation

## Install

### From source

Requires Go 1.22+.

```sh
go install github.com/segfaultd/pdf-to-markdown/cmd/pdf2md@latest
```

### Build locally

```sh
git clone https://github.com/Segfaultd/pdf-to-markdown.git
cd pdf-to-markdown
go build -o pdf2md ./cmd/pdf2md/
```

The binary is at `./pdf2md`.

## Usage

### CLI

```sh
# Convert to stdout
pdf2md report.pdf

# Convert to file
pdf2md report.pdf report.md

# Read from stdin
cat report.pdf | pdf2md -

# With options
pdf2md --page-break --image-dir=assets report.pdf output.md
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--workers` | `-w` | NumCPU | Concurrent page workers |
| `--image-dir` | `-i` | `images` | Image output directory |
| `--page-break` | `-p` | false | Insert `---` between pages |
| `--quiet` | `-q` | false | Suppress stderr timing |
| `--version` | `-v` | | Print version and exit |

### Library

```go
package main

import (
    "os"

    pdf2md "github.com/segfaultd/pdf-to-markdown"
)

func main() {
    // File to file
    pdf2md.ConvertFile("input.pdf", "output.md", nil)

    // Reader to writer
    f, _ := os.Open("input.pdf")
    defer f.Close()
    pdf2md.Convert(f, os.Stdout, &pdf2md.Options{
        ImageDir:  "assets",
        PageBreak: true,
    })

    // Bytes in, bytes out
    data, _ := os.ReadFile("input.pdf")
    md, _ := pdf2md.ConvertBytes(data)
    os.Stdout.Write(md)
}
```

## How it works

The converter runs a 3-stage pipeline:

1. **Extraction** -- reads PDF pages via [ledongthuc/pdf](https://github.com/ledongthuc/pdf) for styled text (font name, size, X/Y position per glyph) and [pdfcpu](https://github.com/pdfcpu/pdfcpu) for embedded images
2. **Structural analysis** -- reconstructs lines from positioned glyphs (Y-bucketing, X-sort, fragment merging), then classifies lines into blocks (headings by font-size ratio, paragraphs by line spacing, lists by prefix patterns, code by monospace font, tables by grid rectangles or column alignment)
3. **Emission** -- streams Markdown to an `io.Writer` as each page completes

## Benchmarks

Measured on Apple M5 Pro, single-threaded, real-world PDFs:

| Document | Pages | Size | Time | Per page |
|----------|-------|------|------|----------|
| Bitcoin Whitepaper | 9 | 180K | 37ms | 4.1ms |
| NIST Cybersecurity Framework | 32 | 1.5MB | 165ms | 5.2ms |
| RFC 2616 (HTTP/1.1) | 176 | 249K | 171ms | 0.97ms |
| CMS Medicare Handbook | 128 | 3.9MB | 1.18s | 9.2ms |
| Arxiv 1706.03762 (Attention Is All You Need) | 15 | 2.1MB | 1.13s | 75ms |
| IRS 1040 Instructions | 126 | 4.4MB | 2.71s | 21.5ms |

Typical text-heavy documents convert at **1-5ms/page**. Dense layouts (tax forms, math-heavy papers) are slower due to the high glyph-fragment count per page.

Run benchmarks yourself:

```sh
go test -bench=. -benchmem
```

## License

MIT
