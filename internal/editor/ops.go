package editor

import (
	"encoding/json"
	"os"
	"strings"

	"vilefetch/internal/layout"
	"vilefetch/internal/store"
)

func cloneWidgets(ws []layout.Widget) []layout.Widget {
	b, err := json.Marshal(ws)
	if err != nil {
		return append([]layout.Widget{}, ws...)
	}
	var out []layout.Widget
	if json.Unmarshal(b, &out) != nil {
		return append([]layout.Widget{}, ws...)
	}
	return out
}

func (m *Model) pushUndo() {
	m.undo = append(m.undo, cloneWidgets(m.layout.Widgets))
	if len(m.undo) > 64 {
		m.undo = m.undo[len(m.undo)-64:]
	}
	m.redo = nil
}

func (m *Model) undoOnce() {
	if len(m.undo) == 0 {
		m.status = "nothing to undo"
		return
	}
	m.redo = append(m.redo, cloneWidgets(m.layout.Widgets))
	m.layout.Widgets = m.undo[len(m.undo)-1]
	m.undo = m.undo[:len(m.undo)-1]
	m.clampSelection()
	m.dirty = true
}

func (m *Model) redoOnce() {
	if len(m.redo) == 0 {
		m.status = "nothing to redo"
		return
	}
	m.undo = append(m.undo, cloneWidgets(m.layout.Widgets))
	m.layout.Widgets = m.redo[len(m.redo)-1]
	m.redo = m.redo[:len(m.redo)-1]
	m.clampSelection()
	m.dirty = true
}

func (m *Model) clampSelection() {
	if m.selected >= len(m.layout.Widgets) {
		m.selected = len(m.layout.Widgets) - 1
	}
}

func (m *Model) placements() []layout.Placement {
	return layout.Resolve(m.layout.Widgets, m.fields)
}

func (m *Model) hit(x, y int) int {
	ps := m.placements()
	best, bestZ := -1, -1<<30
	for i, p := range ps {
		if x >= p.X && y >= p.Y && x < p.X+p.W && y < p.Y+p.H {
			if m.layout.Widgets[i].Z >= bestZ {
				best = i
				bestZ = m.layout.Widgets[i].Z
			}
		}
	}
	return best
}

func (m *Model) nudge(dx, dy int) {
	if m.mode == modePlace && m.ghost != nil {
		m.ghost.X = max(0, m.ghost.X+dx)
		m.ghost.Y = max(0, m.ghost.Y+dy)
		m.cursorX, m.cursorY = m.ghost.X, m.ghost.Y
		m.reveal()
		return
	}
	if m.mode == modeColor || m.selected < 0 {
		m.cursorX = max(0, m.cursorX+dx)
		m.cursorY = max(0, m.cursorY+dy)
		m.reveal()
		return
	}
	m.pushUndo()
	w := &m.layout.Widgets[m.selected]
	w.X = max(0, w.X+dx)
	w.Y = max(0, w.Y+dy)
	m.cursorX, m.cursorY = w.X, w.Y
	m.dirty = true
	m.reveal()
}

func (m *Model) resizeBy(dw, dh int) {
	if m.selected < 0 || m.selected >= len(m.layout.Widgets) {
		return
	}
	ps := m.placements()
	m.pushUndo()
	w := &m.layout.Widgets[m.selected]
	w.W = max(1, ps[m.selected].W+dw)
	w.H = max(1, ps[m.selected].H+dh)
	w.Fit = layout.FitFixed
	m.dirty = true
}

func (m *Model) cycleFit() {
	if m.selected < 0 {
		return
	}
	order := []layout.Fit{layout.FitFixed, layout.FitHugWidth, layout.FitHugHeight, layout.FitHug}
	w := &m.layout.Widgets[m.selected]
	next := order[0]
	for i, f := range order {
		if w.Fit == f || (w.Fit == "" && f == layout.FitHug) {
			next = order[(i+1)%len(order)]
			break
		}
	}
	m.pushUndo()
	w.Fit = next
	m.dirty = true
	m.status = "fit " + string(next)
}

func (m *Model) cyclePush() {
	if m.selected < 0 {
		return
	}
	order := []layout.Push{layout.PushNone, layout.PushRight, layout.PushDown}
	w := &m.layout.Widgets[m.selected]
	next := order[0]
	for i, p := range order {
		if w.Push == p || (w.Push == "" && p == layout.PushNone) {
			next = order[(i+1)%len(order)]
			break
		}
	}
	m.pushUndo()
	w.Push = next
	m.dirty = true
	m.status = "push " + string(next)
}

func (m *Model) cycleSelect(dir int) {
	n := len(m.layout.Widgets)
	if n == 0 {
		return
	}
	if m.selected < 0 {
		m.selected = 0
	} else {
		m.selected = (m.selected + dir + n) % n
	}
	w := m.layout.Widgets[m.selected]
	m.cursorX, m.cursorY = w.X, w.Y
	m.reveal()
}

func (m *Model) deleteSelected() {
	if m.selected < 0 || m.selected >= len(m.layout.Widgets) {
		return
	}
	m.pushUndo()
	m.layout.Widgets = append(m.layout.Widgets[:m.selected], m.layout.Widgets[m.selected+1:]...)
	m.clampSelection()
	m.dirty = true
	m.status = "deleted"
}

func (m *Model) duplicate() {
	if m.selected < 0 || m.selected >= len(m.layout.Widgets) {
		return
	}
	m.pushUndo()
	cp := cloneWidgets([]layout.Widget{m.layout.Widgets[m.selected]})[0]
	cp.ID = m.nextID()
	cp.X += 1
	cp.Y += 1
	cp.Z++
	m.layout.Widgets = append(m.layout.Widgets, cp)
	m.selected = len(m.layout.Widgets) - 1
	m.dirty = true
}

func (m *Model) layer(dir int) {
	if m.selected < 0 {
		return
	}
	m.pushUndo()
	m.layout.Widgets[m.selected].Z += dir
	m.dirty = true
}

func (m *Model) remeasure(i int) {
	if i < 0 || i >= len(m.layout.Widgets) {
		return
	}
	w := &m.layout.Widgets[i]
	nw, nh := layout.Natural(*w, m.fields)
	switch w.Fit {
	case layout.FitFixed:
		return
	case layout.FitHugWidth:
		w.W = nw
	case layout.FitHugHeight:
		w.H = nh
	default:
		w.W, w.H = nw, nh
	}
}

func (m *Model) drop() {
	if m.ghost == nil {
		return
	}
	if m.intent == intentAddRow && m.selected >= 0 && m.selected < len(m.layout.Widgets) &&
		m.layout.Widgets[m.selected].Kind == layout.KindTable && m.ghost.Info != nil &&
		m.layout.Widgets[m.selected].Table != nil {
		m.pushUndo()
		w := &m.layout.Widgets[m.selected]
		w.Table.Rows = append(w.Table.Rows, *m.ghost.Info)
		m.remeasure(m.selected)
		m.dirty = true
		m.ghost = nil
		m.mode = modeNormal
		m.status = "row added"
		return
	}
	m.pushUndo()
	g := *m.ghost
	g.ID = m.nextID()
	m.layout.Widgets = append(m.layout.Widgets, g)
	m.selected = len(m.layout.Widgets) - 1
	m.cursorX, m.cursorY = g.X, g.Y
	m.dirty = true
	m.ghost = nil
	m.mode = modeNormal
	m.status = "placed"
	m.reveal()
}

func (m *Model) placeArt(text string) {
	text = strings.TrimRight(text, "\n")
	if strings.TrimSpace(text) == "" {
		return
	}
	m.pushUndo()
	w := layout.Widget{
		ID: m.nextID(), Kind: layout.KindArt, X: m.cursorX, Y: m.cursorY,
		Fit: layout.FitFixed, Push: layout.PushNone,
		Art: &layout.ArtSpec{Text: text},
	}
	w.W, w.H = layout.Natural(w, m.fields)
	m.layout.Widgets = append(m.layout.Widgets, w)
	m.selected = len(m.layout.Widgets) - 1
	m.dirty = true
	m.mode = modeNormal
	m.status = "pasted art"
}

func (m *Model) save() {
	if err := store.Save(m.layout); err != nil {
		m.status = err.Error()
		return
	}
	m.dirty = false
	m.status = "saved " + m.layout.Name
}

func (m *Model) makeDefault() {
	m.save()
	if m.dirty {
		return
	}
	if err := store.SetDefault(m.layout.Name); err != nil {
		m.status = err.Error()
		return
	}
	m.status = "default is " + m.layout.Name
}

func (m *Model) importPath(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		m.status = err.Error()
		return
	}
	if len(b) > 64*1024 {
		m.status = "file too large"
		return
	}
	m.placeArt(string(b))
}

func (m *Model) cycleAppearance() {
	i := m.selected
	if i < 0 || m.layout.Widgets[i].Kind != layout.KindInfo || m.layout.Widgets[i].Info == nil {
		return
	}
	info := m.layout.Widgets[i].Info
	var def layout.FieldDef
	found := false
	for _, d := range layout.Fields(m.fields) {
		if d.Format == info.Format {
			def = d
			found = true
			break
		}
	}
	if !found {
		return
	}
	apps := layout.AppearancesFor(def)
	cur := 0
	for n, a := range apps {
		if a.Mark == info.Mark && a.Sep == info.Sep && a.ShowLabel == info.ShowLabel && a.Bar == info.Bar {
			cur = n
			break
		}
	}
	next := apps[(cur+1)%len(apps)]
	m.pushUndo()
	spec := layout.SpecFrom(def, next)
	m.layout.Widgets[i].Info = &spec
	m.remeasure(i)
	m.dirty = true
	m.status = next.Name
}

func (m *Model) beginStroke() {
	if !m.stroking {
		m.pushUndo()
		m.stroking = true
	}
}

func (m *Model) stampAt(x, y int, clear bool) {
	i := m.hit(x, y)
	if i < 0 {
		return
	}
	ps := m.placements()
	p := ps[i]
	lx, ly := x-p.X, y-p.Y
	m.beginStroke()
	m.selected = i
	w := &m.layout.Widgets[i]
	if clear || m.colorIdx < 0 {
		w.Paint = removePaint(w.Paint, lx, ly)
	} else {
		cp := layout.CellPaint{X: lx, Y: ly}
		if m.brushFg {
			cp.Fg = &layout.Color{Index: m.colorIdx}
		} else {
			cp.Bg = &layout.Color{Index: m.colorIdx}
		}
		w.Paint = upsertPaint(w.Paint, cp)
	}
	m.dirty = true
}

func removePaint(ps []layout.CellPaint, x, y int) []layout.CellPaint {
	out := ps[:0]
	for _, p := range ps {
		if p.X == x && p.Y == y {
			continue
		}
		out = append(out, p)
	}
	return out
}

func upsertPaint(ps []layout.CellPaint, cp layout.CellPaint) []layout.CellPaint {
	for i := range ps {
		if ps[i].X == cp.X && ps[i].Y == cp.Y {
			if cp.Fg != nil {
				ps[i].Fg = cp.Fg
			}
			if cp.Bg != nil {
				ps[i].Bg = cp.Bg
			}
			return ps
		}
	}
	return append(ps, cp)
}

func (m *Model) sampleAt(x, y int) {
	buf := layout.Paint(m.layout, m.fields)
	cell, ok := buf.At(x, y)
	if !ok || !cell.Drawn {
		return
	}
	if m.brushFg && cell.FgSet {
		m.colorIdx = cell.Fg.Index
		m.status = "sampled fg " + itoa(m.colorIdx)
	}
	if !m.brushFg && cell.BgSet {
		m.colorIdx = cell.Bg.Index
		m.status = "sampled bg " + itoa(m.colorIdx)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
