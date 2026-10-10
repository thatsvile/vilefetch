package layout

import "strings"

// Cell is one terminal column.
type Cell struct {
	R     rune
	W     int
	Fg    Color
	Bg    Color
	FgSet bool
	BgSet bool
	Drawn bool
	Skip  bool
}

// Buffer is a cell grid painted back to front.
type Buffer struct {
	W, H  int
	Cells []Cell
}

// NewBuffer allocates a cleared grid.
func NewBuffer(w, h int) Buffer {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Buffer{W: w, H: h, Cells: make([]Cell, w*h)}
}

func (b *Buffer) index(x, y int) int {
	return y*b.W + x
}

// At returns the cell at x,y. The bool is false when out of range.
func (b *Buffer) At(x, y int) (Cell, bool) {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return Cell{}, false
	}
	return b.Cells[b.index(x, y)], true
}

func (b *Buffer) set(x, y int, c Cell) {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return
	}
	b.Cells[b.index(x, y)] = c
}

// Fill paints every cell with r.
func (b *Buffer) Fill(r rune, fg *Color) {
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			c := Cell{R: r, W: 1, Drawn: true}
			if fg != nil {
				c.Fg = *fg
				c.FgSet = true
			}
			b.set(x, y, c)
		}
	}
}

// WriteGlyphs paints glyphs at x,y and returns the column after the last glyph.
func (b *Buffer) WriteGlyphs(x, y int, glyphs []Glyph) int {
	for _, gl := range glyphs {
		if gl.R == 0 {
			continue
		}
		c := Cell{R: gl.R, W: gl.W, Drawn: true}
		if gl.Fg != nil {
			c.Fg = *gl.Fg
			c.FgSet = true
		}
		if gl.Bg != nil {
			c.Bg = *gl.Bg
			c.BgSet = true
		}
		b.set(x, y, c)
		for i := 1; i < gl.W; i++ {
			b.set(x+i, y, Cell{Skip: true, Drawn: true, W: 1})
		}
		x += gl.W
	}
	return x
}

// SetFg colors an existing cell.
func (b *Buffer) SetFg(x, y int, col Color) {
	c, ok := b.At(x, y)
	if !ok || !c.Drawn || c.Skip {
		return
	}
	c.Fg = col
	c.FgSet = true
	b.set(x, y, c)
}

// SetBg colors an existing cell's background.
func (b *Buffer) SetBg(x, y int, col Color) {
	c, ok := b.At(x, y)
	if !ok || !c.Drawn || c.Skip {
		return
	}
	c.Bg = col
	c.BgSet = true
	b.set(x, y, c)
}

func (b *Buffer) applyPaint(originX, originY, w, h int, paints []CellPaint) {
	for _, p := range paints {
		if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
			continue
		}
		x, y := originX+p.X, originY+p.Y
		c, ok := b.At(x, y)
		if !ok {
			continue
		}
		if c.Skip && x > originX {
			x--
			c, ok = b.At(x, y)
			if !ok {
				continue
			}
		}
		if !c.Drawn {
			continue
		}
		if p.Fg != nil {
			c.Fg = *p.Fg
			c.FgSet = true
		}
		if p.Bg != nil {
			c.Bg = *p.Bg
			c.BgSet = true
		}
		b.set(x, y, c)
	}
}

// Plain returns the grid without ANSI, for tests.
func (b Buffer) Plain() string {
	var rows []string
	for y := 0; y < b.H; y++ {
		var line []rune
		for x := 0; x < b.W; x++ {
			c := b.Cells[b.index(x, y)]
			if c.Skip {
				continue
			}
			if !c.Drawn || c.R == 0 {
				line = append(line, ' ')
				continue
			}
			line = append(line, c.R)
		}
		rows = append(rows, strings.TrimRight(string(line), " "))
	}
	for len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	return strings.Join(rows, "\n")
}

// String renders ANSI SGR for palette indexes 0-15.
func (b Buffer) String() string {
	var bld strings.Builder
	for y := 0; y < b.H; y++ {
		last := -1
		for x := b.W - 1; x >= 0; x-- {
			c := b.Cells[b.index(x, y)]
			if c.Drawn && !c.Skip && (c.R != 0 && c.R != ' ' || c.BgSet) {
				last = x
				break
			}
		}
		fgOn, bgOn := false, false
		var fg, bg Color
		for x := 0; x <= last; x++ {
			c := b.Cells[b.index(x, y)]
			if c.Skip {
				continue
			}
			r := c.R
			if r == 0 {
				r = ' '
			}
			if c.FgSet != fgOn || (c.FgSet && c.Fg != fg) || c.BgSet != bgOn || (c.BgSet && c.Bg != bg) {
				bld.WriteString("\x1b[0m")
				fgOn, bgOn = false, false
				if c.FgSet {
					bld.WriteString(sgrFG(c.Fg.Index))
					fg, fgOn = c.Fg, true
				}
				if c.BgSet {
					bld.WriteString(sgrBG(c.Bg.Index))
					bg, bgOn = c.Bg, true
				}
			}
			bld.WriteRune(r)
		}
		if fgOn || bgOn {
			bld.WriteString("\x1b[0m")
		}
		if y < b.H-1 {
			bld.WriteByte('\n')
		}
	}
	return strings.TrimRight(bld.String(), "\n")
}

// Window returns the visible rectangle as ANSI.
func (b Buffer) Window(x, y, w, h int) string {
	if w < 1 || h < 1 {
		return ""
	}
	sub := NewBuffer(w, h)
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			c, ok := b.At(x+col, y+row)
			if ok {
				sub.set(col, row, c)
			}
		}
	}
	return sub.String()
}

func sgrFG(index int) string {
	if index < 0 {
		index = 0
	}
	if index > 15 {
		index = 15
	}
	return "\x1b[38;5;" + itoa(index) + "m"
}

func sgrBG(index int) string {
	if index < 0 {
		index = 0
	}
	if index > 15 {
		index = 15
	}
	return "\x1b[48;5;" + itoa(index) + "m"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d [4]byte
	i := len(d)
	for n > 0 {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
	}
	return string(d[i:])
}
