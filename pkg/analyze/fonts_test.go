package analyze

import (
	"testing"

	"github.com/segfaultd/pdf-to-markdown/pkg/model"
)

func TestMergeFontStats(t *testing.T) {
	s1 := model.FontStats{
		SizeCounts: map[float64]int{12: 500, 24: 50},
		NameCounts: map[string]int{"Helvetica": 500, "Helvetica-Bold": 50},
	}
	s2 := model.FontStats{
		SizeCounts: map[float64]int{12: 300, 18: 30},
		NameCounts: map[string]int{"Helvetica": 300, "Helvetica-Bold": 30},
	}
	merged := MergeFontStats([]model.FontStats{s1, s2})

	if merged.SizeCounts[12] != 800 {
		t.Errorf("SizeCounts[12] = %d, want 800", merged.SizeCounts[12])
	}
	if merged.SizeCounts[24] != 50 {
		t.Errorf("SizeCounts[24] = %d, want 50", merged.SizeCounts[24])
	}
	if merged.SizeCounts[18] != 30 {
		t.Errorf("SizeCounts[18] = %d, want 30", merged.SizeCounts[18])
	}
}

func TestBodyFontSize(t *testing.T) {
	stats := model.FontStats{
		SizeCounts: map[float64]int{12: 800, 24: 50, 18: 30},
	}
	body := BodyFontSize(stats)
	if body != 12 {
		t.Errorf("BodyFontSize = %v, want 12", body)
	}
}

func TestBodyFontName(t *testing.T) {
	stats := model.FontStats{
		NameCounts: map[string]int{"Helvetica": 800, "Helvetica-Bold": 80},
	}
	name := BodyFontName(stats)
	if name != "Helvetica" {
		t.Errorf("BodyFontName = %q, want Helvetica", name)
	}
}

func TestBodyFontSizeEmpty(t *testing.T) {
	stats := model.FontStats{}
	body := BodyFontSize(stats)
	if body != 12 {
		t.Errorf("BodyFontSize(empty) = %v, want 12 (default)", body)
	}
}

func TestBodyFontSizeTieBreaking(t *testing.T) {
	stats := model.FontStats{
		SizeCounts: map[float64]int{12: 500, 24: 500},
	}
	body := BodyFontSize(stats)
	if body != 12 {
		t.Errorf("BodyFontSize tie = %v, want 12 (smaller wins)", body)
	}
}

func TestMergeFontStatsEmpty(t *testing.T) {
	merged := MergeFontStats(nil)
	if len(merged.SizeCounts) != 0 {
		t.Error("expected empty SizeCounts for nil input")
	}
}

func TestBodyFontNameEmpty(t *testing.T) {
	name := BodyFontName(model.FontStats{})
	if name != "" {
		t.Errorf("BodyFontName(empty) = %q, want empty", name)
	}
}
