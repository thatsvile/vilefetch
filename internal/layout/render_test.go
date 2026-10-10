package layout

import (
	"strings"
	"testing"
)

func TestHugGrowsForLongDisplay(t *testing.T) {
	fields := map[string]string{
		"display.0.name":       "SuperUltraWide Monitor",
		"display.0.resolution": "3440x1440",
	}
	spec := InfoRowSpec{
		Format: "{display.0.name} {display.0.resolution}", Label: "Display",
		ShowLabel: true, Mark: "none", Sep: ": ",
	}
	w := Widget{ID: "d", Kind: KindInfo, X: 0, Y: 0, W: 8, H: 1, Fit: FitHug, Info: &spec}
	ps := Resolve([]Widget{w}, fields)
	want := GlyphWidth(InfoGlyphs(spec, fields))
	if ps[0].W != want {
		t.Fatalf("width %d, want %d (%q)", ps[0].W, want, InfoLine(spec, fields))
	}
	if !strings.Contains(Paint(Layout{Widgets: []Widget{w}}, fields).Plain(), "3440x1440") {
		t.Fatalf("painted %q", Paint(Layout{Widgets: []Widget{w}}, fields).Plain())
	}
}

func TestPushRightMovesNeighbor(t *testing.T) {
	long := Widget{
		ID: "a", Kind: KindText, X: 0, Y: 0, W: 5, H: 1,
		Fit: FitHug, Push: PushRight, Text: &TextSpec{Text: "0123456789"},
	}
	swatch := Widget{
		ID: "b", Kind: KindColors, X: 6, Y: 0, W: 3, H: 1, Fit: FitFixed,
		Swatch: &ColorsSpec{Style: "dots", Which: "normal"},
	}
	ps := Resolve([]Widget{long, swatch}, nil)
	if ps[0].W != 10 {
		t.Fatalf("hug width %d", ps[0].W)
	}
	if ps[1].X != 11 {
		t.Fatalf("swatch x %d, want 11", ps[1].X)
	}
}

func TestPushNoneOverlaps(t *testing.T) {
	long := Widget{
		ID: "a", Kind: KindText, X: 0, Y: 0, W: 5, H: 1,
		Fit: FitHug, Push: PushNone, Text: &TextSpec{Text: "0123456789"},
	}
	other := Widget{
		ID: "b", Kind: KindText, X: 6, Y: 0, W: 3, H: 1, Fit: FitFixed,
		Text: &TextSpec{Text: "xyz"},
	}
	ps := Resolve([]Widget{long, other}, nil)
	if ps[1].X != 6 {
		t.Fatalf("neighbor moved to %d", ps[1].X)
	}
	plain := Paint(Layout{Widgets: []Widget{long, other}}, nil).Plain()
	if !strings.Contains(plain, "xyz") {
		t.Fatalf("overlap paint lost neighbor: %q", plain)
	}
}

func TestPushIgnoresVerticalNeighbor(t *testing.T) {
	long := Widget{
		ID: "a", Kind: KindText, X: 0, Y: 0, W: 5, H: 1,
		Fit: FitHug, Push: PushRight, Text: &TextSpec{Text: "0123456789"},
	}
	below := Widget{
		ID: "b", Kind: KindText, X: 6, Y: 5, W: 3, H: 1, Fit: FitFixed,
		Text: &TextSpec{Text: "xyz"},
	}
	ps := Resolve([]Widget{long, below}, nil)
	if ps[1].X != 6 {
		t.Fatalf("below widget shifted to %d", ps[1].X)
	}
}

func TestFixedWrapsInsideWidth(t *testing.T) {
	spec := InfoRowSpec{
		Format: "SuperUltraWide Monitor 3440x1440", ShowLabel: false, Mark: "none",
	}
	w := Widget{ID: "d", Kind: KindInfo, X: 0, Y: 0, W: 10, H: 6, Fit: FitFixed, Info: &spec}
	plain := Paint(Layout{Widgets: []Widget{w}}, nil).Plain()
	lines := strings.Split(plain, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrap, got %q", plain)
	}
	for _, ln := range lines {
		if textWidth(ln) > 10 {
			t.Fatalf("line wider than box: %q", ln)
		}
	}
	joined := strings.Join(lines, "")
	if !strings.Contains(joined, "3440x1440") {
		t.Fatalf("wrapped text dropped the value: %q", plain)
	}
}

func TestPaintStampColorsGlyph(t *testing.T) {
	w := Widget{
		ID: "t", Kind: KindText, X: 2, Y: 1, W: 5, H: 1, Fit: FitFixed,
		Text:  &TextSpec{Text: "Hello"},
		Paint: []CellPaint{{X: 0, Y: 0, Fg: &Color{Index: 1}}},
	}
	buf := Paint(Layout{Widgets: []Widget{w}}, nil)
	cell, ok := buf.At(2, 1)
	if !ok || cell.R != 'H' || !cell.FgSet || cell.Fg.Index != 1 {
		t.Fatalf("cell %+v ok %v", cell, ok)
	}
	if !strings.Contains(buf.String(), "\x1b[38;5;1m") {
		t.Fatalf("missing ansi: %q", buf.String())
	}
}

func TestMinimumDoesNotShrink(t *testing.T) {
	w := Widget{
		ID: "t", Kind: KindText, X: 0, Y: 0, W: 12, H: 1, Fit: FitHug,
		Text: &TextSpec{Text: "short"},
	}
	ps := Resolve([]Widget{w}, nil)
	if ps[0].W != 12 {
		t.Fatalf("hug shrank to %d", ps[0].W)
	}
}
