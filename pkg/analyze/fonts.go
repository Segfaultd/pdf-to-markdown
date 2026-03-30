package analyze

import "github.com/segfaultd/pdf-to-markdown/pkg/model"

const defaultBodyFontSize = 12.0

func MergeFontStats(stats []model.FontStats) model.FontStats {
	merged := model.FontStats{
		SizeCounts: make(map[float64]int),
		NameCounts: make(map[string]int),
	}
	for _, s := range stats {
		for size, count := range s.SizeCounts {
			merged.SizeCounts[size] += count
		}
		for name, count := range s.NameCounts {
			merged.NameCounts[name] += count
		}
	}
	return merged
}

func BodyFontSize(stats model.FontStats) float64 {
	if len(stats.SizeCounts) == 0 {
		return defaultBodyFontSize
	}
	var maxCount int
	var maxSize float64
	for size, count := range stats.SizeCounts {
		if count > maxCount {
			maxCount = count
			maxSize = size
		}
	}
	return maxSize
}

func BodyFontName(stats model.FontStats) string {
	if len(stats.NameCounts) == 0 {
		return ""
	}
	var maxCount int
	var maxName string
	for name, count := range stats.NameCounts {
		if count > maxCount {
			maxCount = count
			maxName = name
		}
	}
	return maxName
}
