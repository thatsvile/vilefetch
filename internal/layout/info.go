package layout

// LabelGlyphs is the mark and label portion of an info row.
func LabelGlyphs(spec InfoRowSpec) []Glyph {
	var g []Glyph
	switch spec.Mark {
	case "bullet":
		gl := spec.Glyph
		if gl == "" {
			gl = "•"
		}
		g = append(g, GlyphsOf(gl, spec.MarkColor)...)
		if spec.ShowLabel && spec.Label != "" {
			g = append(g, GlyphsOf(" ", nil)...)
		}
	case "icon":
		gl := spec.Glyph
		if gl == "" {
			gl = "▸"
		}
		g = append(g, GlyphsOf(gl, spec.MarkColor)...)
		if spec.ShowLabel && spec.Label != "" {
			g = append(g, GlyphsOf(" ", nil)...)
		}
	}
	if spec.ShowLabel && spec.Label != "" {
		g = append(g, GlyphsOf(spec.Label, spec.LabelColor)...)
	}
	return g
}

// ValueGlyphs is the fetched value and optional percent bar.
func ValueGlyphs(spec InfoRowSpec, fields map[string]string) []Glyph {
	val := Fill(spec.Format, fields)
	g := GlyphsOf(val, spec.ValueColor)
	if spec.Bar && spec.BarField != "" {
		if len(g) > 0 {
			g = append(g, GlyphsOf(" ", nil)...)
		}
		g = append(g, GlyphsOf(Bar(fields[spec.BarField], 10), spec.ValueColor)...)
	}
	return g
}

// InfoGlyphs is the full standalone line the editor shows while placing a row.
func InfoGlyphs(spec InfoRowSpec, fields map[string]string) []Glyph {
	label := LabelGlyphs(spec)
	value := ValueGlyphs(spec, fields)
	if len(label) == 0 {
		return value
	}
	sep := spec.Sep
	if sep == "" {
		sep = ": "
	}
	var g []Glyph
	g = append(g, label...)
	g = append(g, GlyphsOf(sep, spec.LabelColor)...)
	g = append(g, value...)
	return g
}

// InfoLine is InfoGlyphs as a plain string.
func InfoLine(spec InfoRowSpec, fields map[string]string) string {
	var b stringsBuilder
	for _, gl := range InfoGlyphs(spec, fields) {
		b.WriteRune(gl.R)
	}
	return b.String()
}

// stringsBuilder avoids importing strings just for a builder in this file's hot path.
type stringsBuilder struct {
	buf []rune
}

func (b *stringsBuilder) WriteRune(r rune) {
	b.buf = append(b.buf, r)
}

func (b *stringsBuilder) String() string {
	return string(b.buf)
}
