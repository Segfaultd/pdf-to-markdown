# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build -o pdf2md ./cmd/pdf2md/       # Build CLI binary
go test ./...                           # Run all tests
go test -v ./pkg/analyze/ -run TestName # Run a single test by name
go test -bench=. -benchmem ./...        # Run benchmarks with memory profiling
```

Tests generate PDFs on-the-fly using `go-pdf/fpdf` — no fixture files needed.

## Architecture

Pure Go PDF-to-Markdown converter. Three-stage pipeline orchestrated in `convert.go`:

1. **Extract** (`pkg/extract/`) — Pulls text elements (with font name, size, X/Y position) via `ledongthuc/pdf` and images via `pdfcpu`. Bold/italic inferred from font name patterns.

2. **Analyze** (`pkg/analyze/`) — Structural reconstruction:
   - `lines.go`: Y-proximity bucketing + X-sorting to merge glyphs into `Line`/`Span`
   - `fonts.go`: Character-weighted font frequency → body font size detection
   - `tables.go`: Rectangle-based grid detection, fallback to column-alignment heuristic for borderless tables
   - `blocks.go`: Classifies lines into heading/list/code/table/paragraph blocks. Heading levels determined by font-size ratio to body (H1 >= 2.0x, H2 >= 1.7x, etc.). Code blocks require 2+ consecutive monospace lines.

3. **Emit** (`pkg/emit/`) — Streams GFM markdown per block type.

**Data types** live in `pkg/model/types.go`: `TextElement` → `Line`/`Span` → `Block` is the core progression. Tables and images are inserted by Y-position to preserve document order.

**Public API** (`convert.go`): `Convert(io.ReadSeeker, io.Writer, *Options)`, `ConvertFile(in, out, opts)`, `ConvertBytes([]byte)`.

**CLI** (`cmd/pdf2md/main.go`): Wraps the library API with flag parsing. Version injected via `-ldflags`.
