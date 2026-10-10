package layout

import "sort"

func minSize(v, fallback int) int {
	if v < 1 {
		return fallback
	}
	return v
}

func vOverlap(y1, h1, y2, h2 int) bool {
	if h1 < 1 {
		h1 = 1
	}
	if h2 < 1 {
		h2 = 1
	}
	return y1 < y2+h2 && y2 < y1+h1
}

func hOverlap(x1, w1, x2, w2 int) bool {
	if w1 < 1 {
		w1 = 1
	}
	if w2 < 1 {
		w2 = 1
	}
	return x1 < x2+w2 && x2 < x1+w1
}

// Natural is the content size before the minimum box is applied.
func Natural(w Widget, fields map[string]string) (int, int) {
	switch w.Kind {
	case KindArt:
		if w.Art == nil {
			return 1, 1
		}
		lines := linesOf(w.Art.Text)
		return max(1, maxTextWidth(lines)), max(1, len(lines))
	case KindText:
		text := ""
		if w.Text != nil {
			text = w.Text.Text
		}
		lines := linesOf(text)
		return max(1, maxTextWidth(lines)), max(1, len(lines))
	case KindInfo:
		if w.Info == nil {
			return 1, 1
		}
		return max(1, GlyphWidth(InfoGlyphs(*w.Info, fields))), 1
	case KindTable:
		if w.Table == nil {
			return 2, 2
		}
		return tableNatural(*w.Table, fields)
	case KindBox:
		if w.Box == nil {
			return 4, 3
		}
		return boxNatural(*w.Box, fields)
	case KindSeparator:
		dw, dh := w.W, w.H
		if w.Sep != nil && w.Sep.Dir == "v" {
			if dw < 1 {
				dw = 1
			}
			if dh < 1 {
				dh = 8
			}
		} else {
			if dw < 1 {
				dw = 20
			}
			if dh < 1 {
				dh = 1
			}
		}
		return dw, dh
	case KindColors:
		spec := ColorsSpec{Style: "blocks", Which: "both"}
		if w.Swatch != nil {
			spec = *w.Swatch
		}
		g := colorGlyphs(spec)
		return max(1, GlyphWidth(g)), 1
	case KindSpacer:
		return minSize(w.W, 1), minSize(w.H, 1)
	default:
		return minSize(w.W, 1), minSize(w.H, 1)
	}
}

func resolvedSize(w Widget, fields map[string]string) (int, int) {
	nw, nh := Natural(w, fields)
	rw, rh := w.W, w.H
	if rw < 1 {
		rw = nw
	}
	if rh < 1 {
		rh = nh
	}
	fit := w.Fit
	if fit == "" {
		fit = FitHug
	}
	switch fit {
	case FitHug:
		if nw > rw {
			rw = nw
		}
		if nh > rh {
			rh = nh
		}
	case FitHugWidth:
		if nw > rw {
			rw = nw
		}
	case FitHugHeight:
		if nh > rh {
			rh = nh
		}
	case FitFixed:
	}
	if rw < 1 {
		rw = 1
	}
	if rh < 1 {
		rh = 1
	}
	return rw, rh
}

// Resolve applies fit and push. Returned slices match widgets by index.
func Resolve(widgets []Widget, fields map[string]string) []Placement {
	p := make([]Placement, len(widgets))
	for i, w := range widgets {
		rw, rh := resolvedSize(w, fields)
		p[i] = Placement{X: w.X, Y: w.Y, W: rw, H: rh}
	}
	order := make([]int, len(widgets))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		ia, ib := order[a], order[b]
		if widgets[ia].Y != widgets[ib].Y {
			return widgets[ia].Y < widgets[ib].Y
		}
		return widgets[ia].X < widgets[ib].X
	})
	for _, i := range order {
		w := widgets[i]
		minW := w.W
		minH := w.H
		if minW < 1 {
			minW = p[i].W
		}
		if minH < 1 {
			minH = p[i].H
		}
		if w.Push == PushRight && p[i].W > minW {
			grow := p[i].W - minW
			right := w.X + minW
			for j := range widgets {
				if j == i {
					continue
				}
				if widgets[j].X >= right && vOverlap(p[i].Y, p[i].H, p[j].Y, p[j].H) {
					p[j].X += grow
				}
			}
		}
		if w.Push == PushDown && p[i].H > minH {
			grow := p[i].H - minH
			bottom := w.Y + minH
			for j := range widgets {
				if j == i {
					continue
				}
				if widgets[j].Y >= bottom && hOverlap(p[i].X, p[i].W, p[j].X, p[j].W) {
					p[j].Y += grow
				}
			}
		}
	}
	return p
}

func zOrder(widgets []Widget) []int {
	order := make([]int, len(widgets))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return widgets[order[a]].Z < widgets[order[b]].Z
	})
	return order
}

// Paint draws the layout into a buffer sized to the bounding box.
func Paint(l Layout, fields map[string]string) Buffer {
	ps := Resolve(l.Widgets, fields)
	maxX, maxY := 0, 0
	for _, p := range ps {
		if p.X+p.W > maxX {
			maxX = p.X + p.W
		}
		if p.Y+p.H > maxY {
			maxY = p.Y + p.H
		}
	}
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	buf := NewBuffer(maxX, maxY)
	for _, i := range zOrder(l.Widgets) {
		draw(&buf, l.Widgets[i], ps[i], fields)
	}
	return buf
}

// Render paints the layout and returns ANSI text.
func Render(l Layout, fields map[string]string) string {
	if len(l.Widgets) == 0 {
		return ""
	}
	return Paint(l, fields).String()
}

// PaintOnto draws widgets onto an existing buffer, then applies editor overlays.
func PaintOnto(buf *Buffer, widgets []Widget, fields map[string]string) []Placement {
	ps := Resolve(widgets, fields)
	for _, i := range zOrder(widgets) {
		draw(buf, widgets[i], ps[i], fields)
	}
	return ps
}
