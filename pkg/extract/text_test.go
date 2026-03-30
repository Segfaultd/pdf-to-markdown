package extract

import "testing"

func TestInferStyle(t *testing.T) {
	tests := []struct {
		font               string
		bold, italic, mono bool
	}{
		{"Helvetica", false, false, false},
		{"Helvetica-Bold", true, false, false},
		{"Helvetica-BoldOblique", true, true, false},
		{"Helvetica-Oblique", false, true, false},
		{"TimesNewRoman-BdIt", true, true, false},
		{"Courier", false, false, true},
		{"CourierNew-Bold", true, false, true},
		{"ABCDEF+Arial-BoldMT", true, false, false},
		{"Monaco", false, false, true},
		{"Menlo-Regular", false, false, true},
		{"Consolas", false, false, true},
		{"Consolas-Bold", true, false, true},
		{"Helvetica-Italic", false, true, false},
		{"DejaVuSansMono", false, false, true},
		{"BCDEFG+MyCustomFont", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.font, func(t *testing.T) {
			bold, italic, mono := InferStyle(tt.font)
			if bold != tt.bold {
				t.Errorf("InferStyle(%q).bold = %v, want %v", tt.font, bold, tt.bold)
			}
			if italic != tt.italic {
				t.Errorf("InferStyle(%q).italic = %v, want %v", tt.font, italic, tt.italic)
			}
			if mono != tt.mono {
				t.Errorf("InferStyle(%q).mono = %v, want %v", tt.font, mono, tt.mono)
			}
		})
	}
}
