package extract

import (
	"bytes"
	"io"
	"os"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

// ExtractFileImages extracts all embedded images from the PDF at path.
func ExtractFileImages(path string) ([]model.ExtractedImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return extractImages(f)
}

// ExtractReaderImages extracts all embedded images from a PDF supplied as raw bytes.
func ExtractReaderImages(data []byte) ([]model.ExtractedImage, error) {
	return extractImages(bytes.NewReader(data))
}

func extractImages(rs io.ReadSeeker) ([]model.ExtractedImage, error) {
	pages, err := pdfcpuapi.ExtractImagesRaw(rs, nil, nil)
	if err != nil {
		return nil, err
	}

	var out []model.ExtractedImage
	for _, pageMap := range pages {
		for _, img := range pageMap {
			if img.Reader == nil {
				continue
			}
			data, err := io.ReadAll(img.Reader)
			if err != nil {
				return nil, err
			}
			if len(data) == 0 {
				continue
			}
			out = append(out, model.ExtractedImage{
				PageNum: img.PageNr,
				Index:   img.ObjNr,
				Data:    data,
				Format:  normalizeFormat(img.FileType),
			})
		}
	}
	return out, nil
}

// normalizeFormat maps pdfcpu file type strings to the canonical format names
// used by model.ExtractedImage.
func normalizeFormat(ft string) string {
	switch ft {
	case "tif":
		return "tiff"
	case "jpg", "jpeg":
		return "jpg"
	case "png":
		return "png"
	default:
		if ft == "" {
			return "png"
		}
		return ft
	}
}
