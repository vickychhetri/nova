package text

import (
	"strings"
	"unicode/utf8"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
)

// MeasureText calculates the width and height of a single-line or multi-line text string.
func MeasureText(str string, fontSize float64, weight int) geom.Size {
	if str == "" {
		return geom.Sz(0, fontSize*1.2)
	}

	lines := strings.Split(str, "\n")
	maxWidth := 0.0
	lineHeight := fontSize * 1.3

	for _, line := range lines {
		w := 0.0
		for _, r := range line {
			w += font.MeasureCharWidth(r, fontSize, weight)
		}
		if w > maxWidth {
			maxWidth = w
		}
	}

	return geom.Sz(maxWidth, float64(len(lines))*lineHeight)
}

// WrapLines breaks text into multiple lines that fit within maxWidth.
func WrapLines(str string, maxWidth float64, fontSize float64, weight int) []string {
	if str == "" {
		return []string{""}
	}

	var result []string
	rawLines := strings.Split(str, "\n")

	for _, rawLine := range rawLines {
		words := strings.Fields(rawLine)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		currentLine := ""
		currentWidth := 0.0
		spaceWidth := font.MeasureCharWidth(' ', fontSize, weight)

		for _, word := range words {
			wordWidth := 0.0
			for _, r := range word {
				wordWidth += font.MeasureCharWidth(r, fontSize, weight)
			}

			if currentLine == "" {
				currentLine = word
				currentWidth = wordWidth
			} else if currentWidth+spaceWidth+wordWidth <= maxWidth {
				currentLine += " " + word
				currentWidth += spaceWidth + wordWidth
			} else {
				result = append(result, currentLine)
				currentLine = word
				currentWidth = wordWidth
			}
		}

		if currentLine != "" {
			result = append(result, currentLine)
		}
	}

	return result
}

// TruncateWithEllipsis clips a string to fit in maxWidth by appending "...".
func TruncateWithEllipsis(str string, maxWidth float64, fontSize float64, weight int) string {
	sz := MeasureText(str, fontSize, weight)
	if sz.Width <= maxWidth {
		return str
	}

	ellipsis := "..."
	ellipsisWidth := font.MeasureCharWidth('.', fontSize, weight) * 3

	if maxWidth <= ellipsisWidth {
		return ellipsis
	}

	targetWidth := maxWidth - ellipsisWidth
	curWidth := 0.0
	runes := []rune(str)
	var truncated []rune

	for _, r := range runes {
		w := font.MeasureCharWidth(r, fontSize, weight)
		if curWidth+w > targetWidth {
			break
		}
		curWidth += w
		truncated = append(truncated, r)
	}

	if len(truncated) == 0 && utf8.RuneCountInString(str) > 0 {
		return ellipsis
	}

	return string(truncated) + ellipsis
}
