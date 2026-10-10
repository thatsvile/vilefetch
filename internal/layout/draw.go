package layout

import "strings"

func draw(buf *Buffer, w Widget, p Placement, fields map[string]string) {
	switch w.Kind {
	case KindArt:
		text := ""
		if w.Art != nil {
			text = w.Art.Text
		}
		drawLiteral(buf, p, text, w.Style, false)
	case KindText:
		text := ""
		if w.Text != nil {
			text = w.Text.Text
		}
		drawLiteral(buf, p, text, w.Style, w.Fit == FitFixed || w.Fit == FitHugHeight)
	case KindInfo:
		if w.Info != nil {
			wrap := w.Fit == FitFixed || w.Fit == FitHugHeight
			drawGlyphBlock(buf, p, [][]Glyph{InfoGlyphs(*w.Info, fields)}, wrap)
		}
	case KindTable:
		if w.Table != nil {
			drawTable(buf, p, *w.Table, fields, w.Style)
		}
	case KindBox:
		if w.Box != nil {
			drawBox(buf, p, *w.Box, fields, w.Style)
		}
	case KindSeparator:
		drawSep(buf, p, w)
	case KindColors:
		spec := ColorsSpec{Style: "blocks", Which: "both"}
		if w.Swatch != nil {
			spec = *w.Swatch
		}
		drawGlyphBlock(buf, p, [][]Glyph{colorGlyphs(spec)}, false)
	default:
	}
	buf.applyPaint(p.X, p.Y, p.W, p.H, w.Paint)
}

func drawLiteral(buf *Buffer, p Placement, text string, style Style, wrap bool) {
	lines := linesOf(text)
	var blocks [][]Glyph
	for _, ln := range lines {
		g := GlyphsOf(ln, style.Fg)
		for i := range g {
			if style.Bg != nil {
				g[i].Bg = style.Bg
			}
		}
		blocks = append(blocks, g)
	}
	drawGlyphBlock(buf, p, blocks, wrap)
}

func drawGlyphBlock(buf *Buffer, p Placement, lines [][]Glyph, wrap bool) {
	y := p.Y
	for _, line := range lines {
		if y >= p.Y+p.H {
			break
		}
		if wrap && GlyphWidth(line) > p.W {
			for _, wl := range WrapGlyphs(line, p.W) {
				if y >= p.Y+p.H {
					break
				}
				buf.WriteGlyphs(p.X, y, wl)
				y++
			}
			continue
		}
		buf.WriteGlyphs(p.X, y, clipGlyphs(line, p.W))
		y++
	}
}

func clipGlyphs(g []Glyph, width int) []Glyph {
	var out []Glyph
	used := 0
	for _, gl := range g {
		if used+gl.W > width {
			break
		}
		out = append(out, gl)
		used += gl.W
	}
	return out
}

type boxSet struct {
	tl, tr, bl, br, h, v, lj, rj, tj, bj, cross rune
}

func borderSet(name string) (boxSet, bool) {
	switch name {
	case "", "none":
		return boxSet{}, false
	case "thick":
		return boxSet{'┏', '┓', '┗', '┛', '━', '┃', '┣', '┫', '┳', '┻', '╋'}, true
	case "double":
		return boxSet{'╔', '╗', '╚', '╝', '═', '║', '╠', '╣', '╦', '╩', '╬'}, true
	case "normal":
		return boxSet{'┌', '┐', '└', '┘', '─', '│', '├', '┤', '┬', '┴', '┼'}, true
	default: // rounded
		return boxSet{'╭', '╮', '╰', '╯', '─', '│', '├', '┤', '┬', '┴', '┼'}, true
	}
}

func drawBorder(buf *Buffer, x, y, w, h int, name string, fg *Color) {
	set, ok := borderSet(name)
	if !ok || w < 2 || h < 2 {
		return
	}
	put := func(cx, cy int, r rune) {
		buf.WriteGlyphs(cx, cy, GlyphsOf(string(r), fg))
	}
	put(x, y, set.tl)
	put(x+w-1, y, set.tr)
	put(x, y+h-1, set.bl)
	put(x+w-1, y+h-1, set.br)
	for i := 1; i < w-1; i++ {
		put(x+i, y, set.h)
		put(x+i, y+h-1, set.h)
	}
	for i := 1; i < h-1; i++ {
		put(x, y+i, set.v)
		put(x+w-1, y+i, set.v)
	}
}

func drawHLine(buf *Buffer, x, y, w int, r rune, fg *Color) {
	for i := 0; i < w; i++ {
		buf.WriteGlyphs(x+i, y, GlyphsOf(string(r), fg))
	}
}

func tableNatural(spec TableSpec, fields map[string]string) (int, int) {
	labelW, valueW := columnWidths(spec, fields, 0)
	gap := spec.Gap
	if gap == "" {
		gap = "  "
	}
	innerW := labelW + textWidth(gap) + valueW
	innerH := len(spec.Rows)
	if spec.RowLines && innerH > 1 {
		innerH += innerH - 1
	}
	if innerH < 1 {
		innerH = 1
	}
	bw, pad := 0, 0
	if spec.Border != "none" {
		bw = 2
		pad = 2
	}
	return max(1, innerW+bw+pad), max(1, innerH+bw)
}

func columnWidths(spec TableSpec, fields map[string]string, valueLimit int) (int, int) {
	labelW, valueW := 0, 0
	for _, row := range spec.Rows {
		lw := GlyphWidth(LabelGlyphs(row))
		if spec.LabelWidth > 0 {
			lw = spec.LabelWidth
		}
		vw := GlyphWidth(ValueGlyphs(row, fields))
		if lw > labelW {
			labelW = lw
		}
		if vw > valueW {
			valueW = vw
		}
	}
	if valueLimit > 0 && valueW > valueLimit {
		valueW = valueLimit
	}
	if labelW < 1 && len(spec.Rows) > 0 {
		labelW = 0
	}
	return labelW, valueW
}

func drawTable(buf *Buffer, p Placement, spec TableSpec, fields map[string]string, style Style) {
	border := spec.Border
	if border == "" {
		border = "rounded"
	}
	inset := 0
	if border != "none" {
		inset = 1
		drawBorder(buf, p.X, p.Y, p.W, p.H, border, style.Fg)
	}
	pad := 0
	if border != "none" {
		pad = 1
	}
	innerW := p.W - inset*2 - pad*2
	innerH := p.H - inset*2
	if innerW < 1 || innerH < 1 {
		return
	}
	gap := spec.Gap
	if gap == "" {
		gap = "  "
	}
	gapW := textWidth(gap)
	labelW, _ := columnWidths(spec, fields, 0)
	if spec.LabelWidth > 0 {
		labelW = spec.LabelWidth
	}
	valueW := innerW - labelW - gapW
	if valueW < 1 {
		valueW = 1
		if labelW+gapW+valueW > innerW {
			labelW = max(0, innerW-gapW-valueW)
		}
	}
	set, _ := borderSet(border)
	y := p.Y + inset
	limit := p.Y + inset + innerH
	for i, row := range spec.Rows {
		if y >= limit {
			break
		}
		label := padGlyphs(clipGlyphs(LabelGlyphs(row), labelW), labelW, row.LabelColor)
		values := WrapGlyphs(ValueGlyphs(row, fields), valueW)
		if len(values) == 0 {
			values = [][]Glyph{{}}
		}
		for li, line := range values {
			if y >= limit {
				break
			}
			buf.WriteGlyphs(p.X+inset+pad, y, label)
			if li > 0 {
				buf.WriteGlyphs(p.X+inset+pad, y, padGlyphs(nil, labelW, nil))
			}
			buf.WriteGlyphs(p.X+inset+pad+labelW, y, GlyphsOf(gap, row.LabelColor))
			buf.WriteGlyphs(p.X+inset+pad+labelW+gapW, y, clipGlyphs(line, valueW))
			y++
			label = padGlyphs(nil, labelW, nil)
		}
		if spec.RowLines && i < len(spec.Rows)-1 && y < limit && border != "none" {
			drawHLine(buf, p.X+1, y, p.W-2, set.h, style.Fg)
			buf.WriteGlyphs(p.X, y, GlyphsOf(string(set.lj), style.Fg))
			buf.WriteGlyphs(p.X+p.W-1, y, GlyphsOf(string(set.rj), style.Fg))
			y++
		}
	}
}

func padGlyphs(g []Glyph, width int, fg *Color) []Glyph {
	w := GlyphWidth(g)
	if w >= width {
		return g
	}
	return append(g, GlyphsOf(strings.Repeat(" ", width-w), fg)...)
}

func boxNatural(spec BoxSpec, fields map[string]string) (int, int) {
	text := spec.Text
	if spec.Format != "" {
		text = Fill(spec.Format, fields)
	}
	lines := linesOf(text)
	bw, pad := 2, 2
	if spec.Border == "none" {
		bw, pad = 0, 0
	}
	return max(2, maxTextWidth(lines)+bw+pad), max(2, len(lines)+bw)
}

func drawBox(buf *Buffer, p Placement, spec BoxSpec, fields map[string]string, style Style) {
	border := spec.Border
	if border == "" {
		border = "rounded"
	}
	if border != "none" {
		drawBorder(buf, p.X, p.Y, p.W, p.H, border, style.Fg)
	}
	text := spec.Text
	if spec.Format != "" {
		text = Fill(spec.Format, fields)
	}
	inset := 0
	pad := 0
	if border != "none" {
		inset = 1
		pad = 1
	}
	inner := Placement{X: p.X + inset + pad, Y: p.Y + inset, W: max(1, p.W-inset*2-pad*2), H: max(1, p.H-inset*2)}
	drawLiteral(buf, inner, text, style, true)
}

func drawSep(buf *Buffer, p Placement, w Widget) {
	ch := "─"
	dir := "h"
	if w.Sep != nil {
		dir = w.Sep.Dir
		if w.Sep.Char != "" {
			ch = w.Sep.Char
		} else if dir == "v" {
			ch = "│"
		}
	}
	r := []rune(ch)
	if len(r) == 0 {
		r = []rune{'─'}
	}
	if dir == "v" {
		for y := 0; y < p.H; y++ {
			buf.WriteGlyphs(p.X, p.Y+y, GlyphsOf(string(r[0]), w.Style.Fg))
		}
		return
	}
	for x := 0; x < p.W; x++ {
		buf.WriteGlyphs(p.X+x, p.Y, GlyphsOf(string(r[0]), w.Style.Fg))
	}
}

func colorGlyphs(spec ColorsSpec) []Glyph {
	var indexes []int
	switch spec.Which {
	case "normal":
		indexes = seq(0, 8)
	case "bright":
		indexes = seq(8, 16)
	default:
		indexes = seq(0, 16)
	}
	glyph := "●"
	gap := " "
	switch spec.Style {
	case "blocks":
		glyph = "██"
		gap = ""
	case "circles":
		glyph = "◉"
	case "bar":
		glyph = "█"
		gap = ""
	case "dots", "":
		glyph = "●"
	}
	var g []Glyph
	for i, idx := range indexes {
		if i > 0 && gap != "" {
			g = append(g, GlyphsOf(gap, nil)...)
		}
		col := &Color{Index: idx}
		g = append(g, GlyphsOf(glyph, col)...)
	}
	return g
}

func seq(from, to int) []int {
	out := make([]int, 0, to-from)
	for i := from; i < to; i++ {
		out = append(out, i)
	}
	return out
}
