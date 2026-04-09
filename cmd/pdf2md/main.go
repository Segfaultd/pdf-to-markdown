package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	pdf2md "github.com/segfaultd/pdf-to-markdown"
)

var version = "dev"

func main() {
	var (
		workers   int
		imageDir  string
		pageBreak bool
		quiet     bool
		showVer   bool
	)

	flag.IntVar(&workers, "workers", runtime.NumCPU(), "concurrent page workers")
	flag.IntVar(&workers, "w", runtime.NumCPU(), "concurrent page workers (shorthand)")
	flag.StringVar(&imageDir, "image-dir", "images", "image output directory")
	flag.StringVar(&imageDir, "i", "images", "image output directory (shorthand)")
	flag.BoolVar(&pageBreak, "page-break", false, "insert --- between pages")
	flag.BoolVar(&pageBreak, "p", false, "insert --- between pages (shorthand)")
	flag.BoolVar(&quiet, "quiet", false, "suppress stderr progress")
	flag.BoolVar(&quiet, "q", false, "suppress stderr progress (shorthand)")
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.BoolVar(&showVer, "v", false, "print version and exit (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pdf2md [flags] <input.pdf> [output.md]\n\n")
		fmt.Fprintf(os.Stderr, "Convert PDF to Markdown.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  input.pdf    PDF file to convert (use \"-\" for stdin)\n")
		fmt.Fprintf(os.Stderr, "  output.md    Output file (defaults to stdout)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -w, --workers int       concurrent page workers (default: %d)\n", runtime.NumCPU())
		fmt.Fprintf(os.Stderr, "  -i, --image-dir string  image output directory (default: \"images\")\n")
		fmt.Fprintf(os.Stderr, "  -p, --page-break        insert --- between pages\n")
		fmt.Fprintf(os.Stderr, "  -q, --quiet             suppress stderr progress\n")
		fmt.Fprintf(os.Stderr, "  -v, --version           print version and exit\n")
	}

	flag.Parse()

	if showVer {
		fmt.Printf("pdf2md %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := args[0]
	var outputPath string
	if len(args) >= 2 {
		outputPath = args[1]
	}

	opts := &pdf2md.Options{
		Workers:   workers,
		ImageDir:  imageDir,
		PageBreak: pageBreak,
		Quiet:     quiet,
	}

	start := time.Now()

	var err error
	if outputPath != "" {
		err = pdf2md.ConvertFile(inputPath, outputPath, opts)
	} else {
		var r io.ReadSeeker
		if inputPath == "-" {
			data, readErr := io.ReadAll(os.Stdin)
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "pdf2md: error reading stdin: %v\n", readErr)
				os.Exit(1)
			}
			r = bytes.NewReader(data)
		} else {
			f, openErr := os.Open(inputPath)
			if openErr != nil {
				fmt.Fprintf(os.Stderr, "pdf2md: %v\n", openErr)
				os.Exit(1)
			}
			defer f.Close()
			r = f
		}
		err = pdf2md.Convert(r, os.Stdout, opts)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "pdf2md: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		elapsed := time.Since(start)
		fmt.Fprintf(os.Stderr, "pdf2md: %s, %s\n", inputPath, elapsed.Round(time.Millisecond))
	}
}
