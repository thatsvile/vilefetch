package layout

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Glyph is one column of styled text. Wide runes occupy one glyph with W=2.
type Glyph struct {
	R  rune
	W  int
	Fg *Color
	Bg *Color
}

func glyphWidth(r rune) int {
	w := runewidth.RuneWidth(r)
	if w < 1 {
		return 1
	}
	return w
}

// GlyphsOf splits s into glyphs colored with fg.
func GlyphsOf(s string, fg *Color) []Glyph {
	var out []Glyph
	for _, r := range s {
		if r == '\n' || r == '\r' {
			continue
		}
		w := glyphWidth(r)
		out = append(out, Glyph{R: r, W: w, Fg: fg})
	}
	return out
}

// GlyphWidth is the column width of a glyph slice.
func GlyphWidth(g []Glyph) int {
	n := 0
	for _, gl := range g {
		n += gl.W
	}
	return n
}

func textWidth(s string) int {
	return GlyphWidth(GlyphsOf(s, nil))
}

// WrapGlyphs breaks glyphs into lines of at most width columns.
func WrapGlyphs(g []Glyph, width int) [][]Glyph {
	if width < 1 {
		width = 1
	}
	if len(g) == 0 {
		return [][]Glyph{{}}
	}
	var lines [][]Glyph
	var cur []Glyph
	used := 0
	for _, gl := range g {
		if used > 0 && used+gl.W > width {
			lines = append(lines, cur)
			cur = nil
			used = 0
		}
		cur = append(cur, gl)
		used += gl.W
	}
	lines = append(lines, cur)
	return lines
}

func linesOf(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}

func maxTextWidth(lines []string) int {
	m := 0
	for _, ln := range lines {
		if w := textWidth(ln); w > m {
			m = w
		}
	}
	return m
}
