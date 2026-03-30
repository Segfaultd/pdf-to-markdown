package extract

import "strings"

// InferStyle determines bold, italic, and monospace from a PDF font name.
func InferStyle(fontName string) (bold, italic, mono bool) {
	// Strip subset prefix (e.g., "ABCDEF+Arial" → "Arial")
	if idx := strings.Index(fontName, "+"); idx >= 0 && idx < 7 {
		fontName = fontName[idx+1:]
	}
	lower := strings.ToLower(fontName)

	bold = strings.Contains(lower, "bold") ||
		strings.HasSuffix(lower, "bd") ||
		strings.Contains(lower, "-bd") ||
		strings.Contains(lower, ",bd")

	italic = strings.Contains(lower, "italic") ||
		strings.Contains(lower, "oblique") ||
		strings.HasSuffix(lower, "it") ||
		strings.Contains(lower, "-it,") ||
		strings.Contains(lower, "-it ")

	mono = strings.Contains(lower, "courier") ||
		strings.Contains(lower, "mono") ||
		strings.Contains(lower, "consolas") ||
		strings.Contains(lower, "fixed") ||
		strings.Contains(lower, "menlo") ||
		strings.Contains(lower, "monaco")

	return
}
