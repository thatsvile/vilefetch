package editor

import (
	"strings"

	"vilefetch/internal/layout"
)

type palItem struct {
	Name   string
	Kind   string
	Field  layout.FieldDef
	Swatch string
}

func (m *Model) rebuildPalette() {
	var items []palItem
	if m.intent != intentAddRow {
		items = append(items,
			palItem{Name: "Logo", Kind: "logo"},
			palItem{Name: "Text", Kind: "text"},
			palItem{Name: "Box", Kind: "box"},
			palItem{Name: "Rule", Kind: "sep-h"},
			palItem{Name: "Column", Kind: "sep-v"},
			palItem{Name: "Color blocks", Kind: "colors", Swatch: "blocks"},
			palItem{Name: "Color dots", Kind: "colors", Swatch: "dots"},
			palItem{Name: "Color circles", Kind: "colors", Swatch: "circles"},
			palItem{Name: "Color bar", Kind: "colors", Swatch: "bar"},
		)
	}
	for _, f := range layout.Fields(m.fields) {
		items = append(items, palItem{Name: f.Label, Kind: "info", Field: f})
	}
	m.items = items
	m.applyFilter()
}

func (m *Model) applyFilter() {
	m.filtered = m.filtered[:0]
	q := strings.ToLower(m.filter)
	for i, it := range m.items {
		if q == "" || strings.Contains(strings.ToLower(it.Name), q) {
			m.filtered = append(m.filtered, i)
		}
	}
	if len(m.filtered) == 0 {
		m.palIdx = 0
		return
	}
	if m.palIdx >= len(m.filtered) {
		m.palIdx = len(m.filtered) - 1
	}
}

func (m *Model) movePalette(dir int) {
	if len(m.filtered) == 0 {
		return
	}
	m.palIdx += dir
	if m.palIdx < 0 {
		m.palIdx = 0
	}
	if m.palIdx >= len(m.filtered) {
		m.palIdx = len(m.filtered) - 1
	}
}

func (m *Model) armPalette() {
	if len(m.filtered) == 0 {
		return
	}
	it := m.items[m.filtered[m.palIdx]]
	if it.Kind == "info" {
		m.field = it.Field
		m.apps = layout.AppearancesFor(it.Field)
		m.appIdx = 0
		m.appOff = 0
		m.mode = modeAppearance
		m.status = it.Name
		return
	}
	m.ghost = m.shapeGhost(it)
	m.mode = modePlace
	m.status = "place " + it.Name
}

func (m *Model) moveAppearance(dir int) {
	if len(m.apps) == 0 {
		return
	}
	m.appIdx += dir
	if m.appIdx < 0 {
		m.appIdx = 0
	}
	if m.appIdx >= len(m.apps) {
		m.appIdx = len(m.apps) - 1
	}
}

func (m *Model) armAppearance() {
	if m.appIdx < 0 || m.appIdx >= len(m.apps) {
		return
	}
	w := layout.InfoWidget("ghost", m.field, m.apps[m.appIdx], m.cursorX, m.cursorY, m.fields)
	m.ghost = &w
	m.mode = modePlace
	m.status = m.apps[m.appIdx].Name + "  " + layout.InfoLine(*w.Info, m.fields)
}

func (m *Model) shapeGhost(it palItem) *layout.Widget {
	w := layout.Widget{
		ID: "ghost", X: m.cursorX, Y: m.cursorY,
		Fit: layout.FitHug, Push: layout.PushNone,
	}
	switch it.Kind {
	case "logo":
		art, col := layout.Logo(m.fields["os.id"])
		w.Kind = layout.KindArt
		w.Art = &layout.ArtSpec{Text: art}
		w.Style.Fg = &col
		w.Fit = layout.FitFixed
	case "text":
		w.Kind = layout.KindText
		w.Text = &layout.TextSpec{Text: "Text"}
	case "box":
		w.Kind = layout.KindBox
		w.Box = &layout.BoxSpec{Border: "rounded", Text: "box"}
		w.W, w.H = 16, 5
	case "sep-h":
		w.Kind = layout.KindSeparator
		w.Sep = &layout.SeparatorSpec{Dir: "h", Char: "─"}
		w.W, w.H = 20, 1
		w.Fit = layout.FitFixed
	case "sep-v":
		w.Kind = layout.KindSeparator
		w.Sep = &layout.SeparatorSpec{Dir: "v", Char: "│"}
		w.W, w.H = 1, 8
		w.Fit = layout.FitFixed
	case "colors":
		w.Kind = layout.KindColors
		w.Swatch = &layout.ColorsSpec{Style: it.Swatch, Which: "both"}
	default:
		w.Kind = layout.KindText
		w.Text = &layout.TextSpec{Text: it.Name}
	}
	nw, nh := layout.Natural(w, m.fields)
	if w.W < nw {
		w.W = nw
	}
	if w.H < nh {
		w.H = nh
	}
	return &w
}

func (m *Model) openPalette(add bool) {
	m.intent = intentPlace
	if add {
		m.intent = intentAddRow
	}
	m.filter = ""
	m.palIdx = 0
	m.palOff = 0
	m.rebuildPalette()
	m.mode = modePalette
}
