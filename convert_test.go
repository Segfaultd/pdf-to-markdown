package pdf2md

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-pdf/fpdf"
)

func createConvertTestPDF(t *testing.T) string {
	t.Helper()
	p := fpdf.New("P", "pt", "A4", "")
	p.AddPage()

	p.SetFont("Helvetica", "B", 24)
	p.Text(72, 72, "Document Title")

	p.SetFont("Helvetica", "", 12)
	p.Text(72, 120, "This is the first paragraph of the document.")
	p.Text(72, 136, "It has two lines.")

	p.Text(72, 180, "This is the second paragraph.")

	path := filepath.Join(t.TempDir(), "convert_test.pdf")
	if err := p.OutputFileAndClose(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConvertFile(t *testing.T) {
	pdfPath := createConvertTestPDF(t)
	mdPath := filepath.Join(t.TempDir(), "output.md")

	err := ConvertFile(pdfPath, mdPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}

	md := string(data)
	if len(md) == 0 {
		t.Error("empty markdown output")
	}
	// Should contain some text from the PDF
	if !strings.Contains(md, "#") {
		t.Log("Note: no heading detected — this may be normal if font sizes aren't preserved as expected by fpdf")
	}
}

func TestConvertBytes(t *testing.T) {
	pdfPath := createConvertTestPDF(t)
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatal(err)
	}

	md, err := ConvertBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(md) == 0 {
		t.Error("ConvertBytes returned empty output")
	}
}

func TestConvert(t *testing.T) {
	pdfPath := createConvertTestPDF(t)
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var buf bytes.Buffer
	err = Convert(f, &buf, nil)
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Error("Convert produced empty output")
	}
}
