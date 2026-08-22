package text_test

import (
	"testing"

	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/text"
)

func TestTextMeasurement(t *testing.T) {
	sz := text.MeasureText("Hello", 16, font.WeightRegular)
	if sz.Width <= 0 || sz.Height <= 0 {
		t.Fatalf("expected positive dimensions for 'Hello', got %s", sz)
	}

	szEmpty := text.MeasureText("", 16, font.WeightRegular)
	if szEmpty.Width != 0 {
		t.Fatalf("expected 0 width for empty string, got %f", szEmpty.Width)
	}
}

func TestTextWrap(t *testing.T) {
	longStr := "The quick brown fox jumps over the lazy dog"
	lines := text.WrapLines(longStr, 100, 14, font.WeightRegular)
	if len(lines) <= 1 {
		t.Fatalf("expected multiple wrapped lines, got %d lines: %+v", len(lines), lines)
	}
}

func TestTextTruncate(t *testing.T) {
	longStr := "Supercalifragilisticexpialidocious"
	truncated := text.TruncateWithEllipsis(longStr, 60, 14, font.WeightRegular)
	if len(truncated) >= len(longStr) || truncated[len(truncated)-3:] != "..." {
		t.Fatalf("expected truncated string with ellipsis, got: %s", truncated)
	}
}
