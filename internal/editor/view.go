package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"vilefetch/internal/layout"
)

type frame struct {
	canvasX, canvasY int
	canvasW, canvasH int
	palW             int
	appY, appH       int
}

func (m Model) helpString() string {
	h := m.help
	if m.width > 0 {
		h.Width = m.width
	}
	h.ShowAll = m.showHelp
	return h.View(m.activeHelp())
}

func (m Model) frame() frame {
	helpLines := 1
	if hs := m.helpString(); hs != "" {
		helpLines = strings.Count(hs, "\n") + 1
	}
	extra := 0
	if m.mode == modePrompt || m.mode == modeProps || m.mode == modeQuit {
		extra = 1
	}
	appH := 0
	if (m.mode == modeAppearance || m.mode == modePlace) && len(m.apps) > 0 {
		appH = len(m.apps)
		if appH > 6 {
			appH = 6
		}
	}
	palW := 0
	if m.mode == modePalette || m.mode == modeAppearance || m.mode == modePlace {
		palW = 22
	}
	bottom := 1 + helpLines + appH + extra
	canvasY := 1
	canvasH := m.height - canvasY - bottom
	if canvasH < 1 {
		canvasH = 1
	}
	canvasX := 0
	if palW > 0 {
		canvasX = palW + 1
	}
	canvasW := m.width - canvasX
	if canvasW < 1 {
		canvasW = 1
	}
	return frame{
		canvasX: canvasX, canvasY: canvasY,
		canvasW: canvasW, canvasH: canvasH,
		palW: palW, appY: canvasY + canvasH, appH: appH,
	}
}

func (m *Model) reveal() {
	if m.height < 4 {
		return
	}
	f := m.frame()
	if m.cursorX < m.camX {
		m.camX = m.cursorX
	}
	if m.cursorY < m.camY {
		m.camY = m.cursorY
	}
	if m.cursorX >= m.camX+f.canvasW {
		m.camX = m.cursorX - f.canvasW + 1
	}
	if m.cursorY >= m.camY+f.canvasH {
		m.camY = m.cursorY - f.canvasH + 1
	}
	if m.camX < 0 {
		m.camX = 0
	}
	if m.camY < 0 {
		m.camY = 0
	}
}

func (m Model) View() string {
	if m.mode == modePreview {
		body := layout.Render(m.layout, m.fields)
		return body + "\n" + m.helpString()
	}
	if m.height < 4 || m.width < 10 {
		return "vilefetch"
	}
	f := m.frame()
	var lines []string
	lines = append(lines, m.titleLine(f))
	lines = append(lines, m.bodyLines(f)...)
	lines = append(lines, m.appearanceLines(f)...)
	if m.mode == modePrompt || m.mode == modeProps || m.mode == modeQuit {
		lines = append(lines, pad(m.promptLine(), m.width))
	}
	lines = append(lines, pad(m.statusLine(), m.width))
	lines = append(lines, m.helpString())
	return strings.Join(lines, "\n")
}

func (m Model) titleLine(f frame) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Render(" vilefetch")
	name := "  " + m.layout.Name
	if m.dirty {
		name += "*"
	}
	rest := ""
	if m.mode == modeColor {
		rest = "  " + swatches(m.colorIdx)
	}
	if m.mode == modePalette && m.filter != "" {
		rest = "  /" + m.filter
	}
	return title + name + rest + strings.Repeat(" ", max(0, m.width-lipgloss.Width(title+name+rest)))
}

func swatches(active int) string {
	var b strings.Builder
	for i := 0; i < 16; i++ {
		g := "●"
		if i == active {
			g = "◉"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", i))).Render(g))
	}
	return b.String()
}

func (m Model) bodyLines(f frame) []string {
	canvas := m.canvasLines(f)
	var pal []string
	if f.palW > 0 {
		pal = m.paletteLines(f)
	}
	lines := make([]string, f.canvasH)
	for i := 0; i < f.canvasH; i++ {
		row := ""
		if f.palW > 0 {
			row = pal[i] + " "
		}
		if i < len(canvas) {
			row += canvas[i]
		}
		lines[i] = row
	}
	return lines
}

func (m Model) canvasLines(f frame) []string {
	buf, ps := m.scene()
	if m.selected >= 0 && m.selected < len(ps) && m.mode != modeColor {
		markCorners(&buf, ps[m.selected])
	}
	if m.mode == modeColor {
		col := m.colorIdx
		if col < 0 {
			col = 7
		}
		buf.SetBg(m.cursorX, m.cursorY, layout.Color{Index: col})
	} else if m.selected < 0 && m.ghost == nil {
		buf.SetBg(m.cursorX, m.cursorY, layout.Color{Index: 8})
	}
	raw := buf.Window(m.camX, m.camY, f.canvasW, f.canvasH)
	lines := strings.Split(raw, "\n")
	for len(lines) < f.canvasH {
		lines = append(lines, "")
	}
	if len(lines) > f.canvasH {
		lines = lines[:f.canvasH]
	}
	return lines
}

func (m Model) scene() (layout.Buffer, []layout.Placement) {
	widgets := append([]layout.Widget{}, m.layout.Widgets...)
	if m.ghost != nil {
		g := *m.ghost
		g.Z = 1 << 20
		widgets = append(widgets, g)
	}
	ps := layout.Resolve(widgets, m.fields)
	maxX, maxY := m.layout.Canvas.Cols, m.layout.Canvas.Rows
	for _, p := range ps {
		if p.X+p.W+1 > maxX {
			maxX = p.X + p.W + 1
		}
		if p.Y+p.H+1 > maxY {
			maxY = p.Y + p.H + 1
		}
	}
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	buf := layout.NewBuffer(maxX, maxY)
	dot := layout.Color{Index: 8}
	buf.Fill('·', &dot)
	layout.PaintOnto(&buf, widgets, m.fields)
	return buf, layout.Resolve(m.layout.Widgets, m.fields)
}

func markCorners(buf *layout.Buffer, p layout.Placement) {
	col := layout.Color{Index: 6}
	pts := [][2]int{{0, 0}, {p.W - 1, 0}, {0, p.H - 1}, {p.W - 1, p.H - 1}}
	for _, pt := range pts {
		buf.SetFg(p.X+pt[0], p.Y+pt[1], col)
	}
}

func (m Model) paletteLines(f frame) []string {
	off := windowOff(m.palIdx, len(m.filtered), f.canvasH)
	lines := make([]string, f.canvasH)
	for i := 0; i < f.canvasH; i++ {
		idx := off + i
		label := ""
		on := false
		if idx >= 0 && idx < len(m.filtered) {
			label = " " + m.items[m.filtered[idx]].Name
			on = idx == m.palIdx
		}
		plain := pad(label, f.palW)
		if on {
			lines[i] = lipgloss.NewStyle().Reverse(true).Render(plain)
		} else {
			lines[i] = plain
		}
	}
	return lines
}

func (m Model) appearanceLines(f frame) []string {
	if f.appH == 0 {
		return nil
	}
	off := windowOff(m.appIdx, len(m.apps), f.appH)
	lines := make([]string, 0, f.appH)
	for i := 0; i < f.appH; i++ {
		idx := off + i
		if idx < 0 || idx >= len(m.apps) {
			lines = append(lines, "")
			continue
		}
		spec := layout.SpecFrom(m.field, m.apps[idx])
		line := " " + m.apps[idx].Name + "  " + layout.InfoLine(spec, m.fields)
		plain := pad(line, m.width)
		if idx == m.appIdx {
			lines = append(lines, lipgloss.NewStyle().Reverse(true).Render(plain))
		} else {
			lines = append(lines, plain)
		}
	}
	return lines
}

func (m Model) promptLine() string {
	switch m.mode {
	case modeQuit:
		return " save changes?  y save   n discard   esc cancel"
	case modePrompt:
		return " import " + m.prompt.View()
	case modeProps:
		if m.propIdx < 0 || m.propIdx >= len(m.props) {
			return ""
		}
		f := m.props[m.propIdx]
		val := f.Get()
		if len(f.Enum) == 0 {
			val = m.prompt.View()
		}
		return " " + f.Name + "  " + val
	default:
		return ""
	}
}

func (m Model) statusLine() string {
	name := m.layout.Name
	if m.dirty {
		name += "*"
	}
	sel := "cursor"
	if m.selected >= 0 && m.selected < len(m.layout.Widgets) {
		w := m.layout.Widgets[m.selected]
		ps := m.placements()
		p := ps[m.selected]
		fit := w.Fit
		if fit == "" {
			fit = layout.FitHug
		}
		push := w.Push
		if push == "" {
			push = layout.PushNone
		}
		sel = fmt.Sprintf("%s %d,%d %dx%d %s push:%s", w.Kind, w.X, w.Y, p.W, p.H, fit, push)
	}
	if m.mode == modeColor {
		ch := "fg"
		if !m.brushFg {
			ch = "bg"
		}
		sel = fmt.Sprintf("paint %s %d", ch, m.colorIdx)
	}
	line := " " + name + "  " + sel
	if m.status != "" && m.mode != modeQuit {
		line += "  ·  " + m.status
	}
	return line
}

func windowOff(idx, n, view int) int {
	if view < 1 || n <= view {
		return 0
	}
	off := idx - view/2
	if off < 0 {
		off = 0
	}
	if off > n-view {
		off = n - view
	}
	return off
}

func pad(s string, n int) string {
	if n < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w == n {
		return s
	}
	if w > n {
		r := []rune(s)
		for len(r) > 0 && lipgloss.Width(string(r)) > n {
			r = r[:len(r)-1]
		}
		return string(r)
	}
	return s + strings.Repeat(" ", n-w)
}
