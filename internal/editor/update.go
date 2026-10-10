package editor

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"vilefetch/internal/layout"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.MouseMsg:
		m.onMouse(msg)
		return m, nil
	case tea.KeyMsg:
		if msg.Paste {
			if strings.Contains(string(msg.Runes), "\n") || m.mode == modeNormal || m.mode == modePlace {
				m.placeArt(string(msg.Runes))
				return m, nil
			}
		}
		return m.onKey(msg)
	}
	return m, nil
}

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "ctrl+c" {
		return m.requestQuit()
	}
	switch m.mode {
	case modePreview:
		if s == "esc" || s == "ctrl+p" || s == "q" {
			m.mode = modeNormal
		}
		return m, nil
	case modeQuit:
		switch s {
		case "y":
			m.save()
			if !m.dirty {
				return m, tea.Quit
			}
			return m, nil
		case "n":
			return m, tea.Quit
		case "esc":
			m.mode = modeNormal
		}
		return m, nil
	case modePrompt:
		return m.promptKey(msg)
	case modeProps:
		return m.propsKey(msg)
	case modePalette:
		return m.paletteKey(s)
	case modeAppearance:
		return m.appearanceKey(s)
	case modePlace:
		return m.placeKey(s)
	case modeColor:
		return m.colorKey(s)
	default:
		return m.normalKey(s)
	}
}

func (m Model) requestQuit() (tea.Model, tea.Cmd) {
	if m.dirty {
		m.mode = modeQuit
		m.status = "save changes? y / n"
		return m, nil
	}
	return m, tea.Quit
}

func (m Model) normalKey(s string) (tea.Model, tea.Cmd) {
	switch s {
	case "q":
		return m.requestQuit()
	case "up", "k":
		m.nudge(0, -1)
	case "down", "j":
		m.nudge(0, 1)
	case "left", "h":
		m.nudge(-1, 0)
	case "right", "l":
		m.nudge(1, 0)
	case "shift+up":
		m.resizeBy(0, -1)
	case "shift+down":
		m.resizeBy(0, 1)
	case "shift+left":
		m.resizeBy(-1, 0)
	case "shift+right":
		m.resizeBy(1, 0)
	case "tab":
		m.cycleSelect(1)
	case "shift+tab":
		m.cycleSelect(-1)
	case "p":
		m.openPalette(false)
	case "a":
		if m.selected >= 0 && m.layout.Widgets[m.selected].Kind == layout.KindTable {
			m.openPalette(true)
		} else {
			m.status = "select a table to add a row"
		}
	case "c":
		if m.selected < 0 {
			m.selected = m.hit(m.cursorX, m.cursorY)
		}
		if m.selected < 0 {
			m.status = "select something to paint"
			return m, nil
		}
		m.mode = modeColor
		m.stroking = false
	case "i":
		m.openImport()
	case "enter":
		m.openProps()
	case "f":
		m.cycleFit()
	case "g":
		m.cyclePush()
	case "v":
		m.cycleAppearance()
	case "[":
		m.layer(-1)
	case "]":
		m.layer(1)
	case "d":
		m.duplicate()
	case "x":
		m.deleteSelected()
	case "u":
		m.undoOnce()
	case "ctrl+r":
		m.redoOnce()
	case "ctrl+s":
		m.save()
	case "ctrl+p":
		m.mode = modePreview
	case "`":
		m.makeDefault()
	case "?":
		m.showHelp = !m.showHelp
	case "esc":
		m.selected = -1
		m.showHelp = false
	}
	return m, nil
}

func (m Model) paletteKey(s string) (tea.Model, tea.Cmd) {
	switch s {
	case "esc":
		m.mode = modeNormal
		m.intent = intentPlace
	case "up":
		m.movePalette(-1)
	case "down":
		m.movePalette(1)
	case "enter":
		m.armPalette()
	case "backspace":
		if m.filter != "" {
			r := []rune(m.filter)
			m.filter = string(r[:len(r)-1])
			m.applyFilter()
		}
	default:
		if len(s) == 1 && s[0] >= 32 && s[0] < 127 {
			m.filter += s
			m.palIdx = 0
			m.applyFilter()
		}
	}
	return m, nil
}

func (m Model) appearanceKey(s string) (tea.Model, tea.Cmd) {
	switch s {
	case "esc":
		m.mode = modePalette
	case "up", "k":
		m.moveAppearance(-1)
	case "down", "j":
		m.moveAppearance(1)
	case "enter":
		m.armAppearance()
	}
	return m, nil
}

func (m Model) placeKey(s string) (tea.Model, tea.Cmd) {
	switch s {
	case "esc":
		m.ghost = nil
		if m.field.Key != "" && len(m.apps) > 0 {
			m.mode = modeAppearance
		} else {
			m.mode = modeNormal
		}
	case "enter":
		m.drop()
	case "up", "k":
		m.nudge(0, -1)
	case "down", "j":
		m.nudge(0, 1)
	case "left", "h":
		m.nudge(-1, 0)
	case "right", "l":
		m.nudge(1, 0)
	}
	return m, nil
}

func (m Model) colorKey(s string) (tea.Model, tea.Cmd) {
	switch s {
	case "esc":
		m.mode = modeNormal
		m.stroking = false
	case "up", "k":
		m.nudge(0, -1)
	case "down", "j":
		m.nudge(0, 1)
	case "left", "h":
		m.nudge(-1, 0)
	case "right", "l":
		m.nudge(1, 0)
	case "f":
		m.brushFg = true
		m.status = "foreground"
	case "b":
		m.brushFg = false
		m.status = "background"
	case "0":
		m.colorIdx = -1
		m.status = "default color"
	case "1", "2", "3", "4", "5", "6", "7", "8":
		m.colorIdx = int(s[0] - '1')
	case "!", "@", "#", "$", "%", "^", "&", "*":
		m.colorIdx = brightIndex(s)
	case " ":
		m.stroking = false
		m.stampAt(m.cursorX, m.cursorY, false)
	case "x":
		m.stroking = false
		m.stampAt(m.cursorX, m.cursorY, true)
	case "o":
		m.sampleAt(m.cursorX, m.cursorY)
	}
	return m, nil
}

func brightIndex(s string) int {
	switch s {
	case "!":
		return 8
	case "@":
		return 9
	case "#":
		return 10
	case "$":
		return 11
	case "%":
		return 12
	case "^":
		return 13
	case "&":
		return 14
	case "*":
		return 15
	default:
		return 8
	}
}

func (m *Model) openImport() {
	m.prompt = textinput.New()
	m.prompt.Placeholder = "path to .txt"
	m.prompt.Focus()
	m.prompt.CharLimit = 512
	m.prompt.Width = 48
	m.promptKind = "import"
	m.mode = modePrompt
}

func (m Model) promptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "enter":
		m.importPath(strings.TrimSpace(m.prompt.Value()))
		return m, nil
	default:
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		return m, cmd
	}
}

func (m *Model) openProps() {
	if m.selected < 0 || m.selected >= len(m.layout.Widgets) {
		return
	}
	m.propSnap = cloneWidgets(m.layout.Widgets)
	m.props = m.propFields(m.selected)
	m.propIdx = 0
	m.mode = modeProps
	m.syncPrompt()
}

func (m *Model) syncPrompt() {
	if m.propIdx < 0 || m.propIdx >= len(m.props) {
		return
	}
	f := m.props[m.propIdx]
	if len(f.Enum) > 0 {
		return
	}
	ti := textinput.New()
	ti.SetValue(f.Get())
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 48
	m.prompt = ti
}

func (m *Model) commitProp() {
	if m.propIdx < 0 || m.propIdx >= len(m.props) {
		return
	}
	f := m.props[m.propIdx]
	if len(f.Enum) == 0 {
		f.Set(m.prompt.Value())
	}
}

func (m Model) propsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.layout.Widgets = m.propSnap
		m.mode = modeNormal
		return m, nil
	case "enter":
		m.commitProp()
		m.undo = append(m.undo, m.propSnap)
		m.redo = nil
		m.dirty = true
		m.mode = modeNormal
		return m, nil
	case "tab", "shift+tab":
		m.commitProp()
		if n := len(m.props); n > 0 {
			if msg.String() == "tab" {
				m.propIdx = (m.propIdx + 1) % n
			} else {
				m.propIdx = (m.propIdx - 1 + n) % n
			}
			m.syncPrompt()
		}
		return m, nil
	case "up", "down":
		m.cycleEnum(msg.String() == "down")
		return m, nil
	default:
		if m.propIdx >= 0 && m.propIdx < len(m.props) && len(m.props[m.propIdx].Enum) == 0 {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

func (m *Model) cycleEnum(next bool) {
	if m.propIdx < 0 || m.propIdx >= len(m.props) {
		return
	}
	f := m.props[m.propIdx]
	if len(f.Enum) == 0 {
		return
	}
	cur := f.Get()
	idx := 0
	for i, opt := range f.Enum {
		if opt == cur {
			idx = i
			break
		}
	}
	if next {
		idx = (idx + 1) % len(f.Enum)
	} else {
		idx = (idx - 1 + len(f.Enum)) % len(f.Enum)
	}
	f.Set(f.Enum[idx])
	m.dirty = true
}

func (m *Model) propFields(i int) []propField {
	var fs []propField
	text := func(name string, get func() string, set func(string)) {
		fs = append(fs, propField{Name: name, Get: get, Set: set})
	}
	enum := func(name string, opts []string, get func() string, set func(string)) {
		fs = append(fs, propField{Name: name, Enum: opts, Get: get, Set: set})
	}
	w := m.layout.Widgets[i]
	switch w.Kind {
	case layout.KindInfo:
		if w.Info == nil {
			break
		}
		text("label", func() string { return m.layout.Widgets[i].Info.Label }, func(s string) {
			m.layout.Widgets[i].Info.Label = s
			m.remeasure(i)
		})
		text("sep", func() string { return m.layout.Widgets[i].Info.Sep }, func(s string) {
			m.layout.Widgets[i].Info.Sep = s
			m.remeasure(i)
		})
		text("glyph", func() string { return m.layout.Widgets[i].Info.Glyph }, func(s string) {
			m.layout.Widgets[i].Info.Glyph = s
			if s != "" && m.layout.Widgets[i].Info.Mark == "none" {
				m.layout.Widgets[i].Info.Mark = "icon"
			}
			m.remeasure(i)
		})
		text("format", func() string { return m.layout.Widgets[i].Info.Format }, func(s string) {
			m.layout.Widgets[i].Info.Format = s
			m.remeasure(i)
		})
	case layout.KindText:
		if w.Text == nil {
			break
		}
		text("text", func() string { return m.layout.Widgets[i].Text.Text }, func(s string) {
			m.layout.Widgets[i].Text.Text = s
			m.remeasure(i)
		})
	case layout.KindBox:
		if w.Box == nil {
			break
		}
		text("text", func() string { return m.layout.Widgets[i].Box.Text }, func(s string) {
			m.layout.Widgets[i].Box.Text = s
		})
		enum("border", []string{"rounded", "normal", "thick", "double", "none"},
			func() string { return m.layout.Widgets[i].Box.Border },
			func(s string) { m.layout.Widgets[i].Box.Border = s })
	case layout.KindTable:
		if w.Table == nil {
			break
		}
		enum("border", []string{"rounded", "normal", "thick", "double", "none"},
			func() string {
				if m.layout.Widgets[i].Table.Border == "" {
					return "rounded"
				}
				return m.layout.Widgets[i].Table.Border
			},
			func(s string) { m.layout.Widgets[i].Table.Border = s })
		text("gap", func() string { return m.layout.Widgets[i].Table.Gap }, func(s string) {
			m.layout.Widgets[i].Table.Gap = s
			m.remeasure(i)
		})
	case layout.KindSeparator:
		if w.Sep == nil {
			break
		}
		text("char", func() string { return m.layout.Widgets[i].Sep.Char }, func(s string) {
			m.layout.Widgets[i].Sep.Char = s
		})
		enum("dir", []string{"h", "v"},
			func() string { return m.layout.Widgets[i].Sep.Dir },
			func(s string) { m.layout.Widgets[i].Sep.Dir = s })
	case layout.KindColors:
		if w.Swatch == nil {
			break
		}
		enum("style", []string{"dots", "blocks", "circles", "bar"},
			func() string { return m.layout.Widgets[i].Swatch.Style },
			func(s string) {
				m.layout.Widgets[i].Swatch.Style = s
				m.remeasure(i)
			})
		enum("which", []string{"normal", "bright", "both"},
			func() string { return m.layout.Widgets[i].Swatch.Which },
			func(s string) {
				m.layout.Widgets[i].Swatch.Which = s
				m.remeasure(i)
			})
	}
	enum("fit", []string{"hug", "hugWidth", "hugHeight", "fixed"},
		func() string {
			if m.layout.Widgets[i].Fit == "" {
				return "hug"
			}
			return string(m.layout.Widgets[i].Fit)
		},
		func(s string) { m.layout.Widgets[i].Fit = layout.Fit(s) })
	enum("push", []string{"none", "right", "down"},
		func() string {
			if m.layout.Widgets[i].Push == "" {
				return "none"
			}
			return string(m.layout.Widgets[i].Push)
		},
		func(s string) { m.layout.Widgets[i].Push = layout.Push(s) })
	return fs
}

func (m *Model) onMouse(msg tea.MouseMsg) {
	if m.width == 0 || m.mode == modePreview || m.mode == modeQuit || m.mode == modePrompt || m.mode == modeProps {
		return
	}
	f := m.frame()
	if msg.Button == tea.MouseButtonWheelUp {
		m.camY = max(0, m.camY-3)
		return
	}
	if msg.Button == tea.MouseButtonWheelDown {
		m.camY += 3
		return
	}
	if msg.Action == tea.MouseActionRelease {
		m.dragging = false
		m.dragUndo = false
		m.stroking = false
		return
	}
	if f.palW > 0 && msg.X < f.palW && msg.Y >= f.canvasY && msg.Y < f.canvasY+f.canvasH {
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			row := msg.Y - f.canvasY + m.palOff
			if row >= 0 && row < len(m.filtered) {
				m.palIdx = row
				m.armPalette()
			}
		}
		return
	}
	if f.appH > 0 && msg.Y >= f.appY && msg.Y < f.appY+f.appH {
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			idx := msg.Y - f.appY + m.appOff
			if idx >= 0 && idx < len(m.apps) {
				m.appIdx = idx
				m.armAppearance()
			}
		}
		return
	}
	if msg.X < f.canvasX || msg.Y < f.canvasY || msg.Y >= f.canvasY+f.canvasH || msg.X >= f.canvasX+f.canvasW {
		return
	}
	cx := msg.X - f.canvasX + m.camX
	cy := msg.Y - f.canvasY + m.camY
	if cx < 0 || cy < 0 {
		return
	}
	if m.mode == modeColor && (msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion) && msg.Button == tea.MouseButtonLeft {
		m.cursorX, m.cursorY = cx, cy
		m.stampAt(cx, cy, false)
		return
	}
	if m.mode == modePlace && msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && m.ghost != nil {
		m.ghost.X, m.ghost.Y = cx, cy
		m.drop()
		return
	}
	if m.mode != modeNormal && m.mode != modePlace {
		return
	}
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		i := m.hit(cx, cy)
		m.cursorX, m.cursorY = cx, cy
		if i < 0 {
			m.selected = -1
			return
		}
		m.selected = i
		ps := m.placements()
		p := ps[i]
		w := m.layout.Widgets[i]
		m.dragging = true
		m.dragUndo = false
		m.pressX, m.pressY = cx, cy
		m.originX, m.originY = w.X, w.Y
		m.originW, m.originH = p.W, p.H
		m.dragKind = "move"
		if (p.W > 1 && cx == p.X+p.W-1) || (p.H > 1 && cy == p.Y+p.H-1) {
			m.dragKind = "resize"
		}
		return
	}
	if msg.Action == tea.MouseActionMotion && m.dragging && m.selected >= 0 {
		if !m.dragUndo {
			m.pushUndo()
			m.dragUndo = true
		}
		w := &m.layout.Widgets[m.selected]
		dx, dy := cx-m.pressX, cy-m.pressY
		if m.dragKind == "resize" {
			w.W = max(1, m.originW+dx)
			w.H = max(1, m.originH+dy)
			w.Fit = layout.FitFixed
		} else {
			w.X = max(0, m.originX+dx)
			w.Y = max(0, m.originY+dy)
			m.cursorX, m.cursorY = w.X, w.Y
		}
		m.dirty = true
	}
}
