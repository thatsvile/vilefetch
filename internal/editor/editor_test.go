package editor

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"vilefetch/internal/layout"
)

func TestPlaceLiveInfoRow(t *testing.T) {
	fields := map[string]string{"os.pretty": "Arch Linux"}
	m := New(layout.Layout{Name: "t", Canvas: layout.Canvas{Cols: 80, Rows: 30}}, fields)
	m.width, m.height = 100, 40
	m.openPalette(false)
	found := false
	for i, idx := range m.filtered {
		if m.items[idx].Name == "OS" {
			m.palIdx = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("OS missing from palette")
	}
	m.armPalette()
	if m.mode != modeAppearance {
		t.Fatalf("mode %d", m.mode)
	}
	m.appIdx = 0
	m.armAppearance()
	if m.ghost == nil || m.ghost.Info == nil {
		t.Fatal("missing ghost")
	}
	line := layout.InfoLine(*m.ghost.Info, fields)
	if line == "" || m.ghost.W != layout.GlyphWidth(layout.InfoGlyphs(*m.ghost.Info, fields)) {
		t.Fatalf("ghost %q width %d", line, m.ghost.W)
	}
	m.drop()
	if len(m.layout.Widgets) != 1 || m.layout.Widgets[0].Kind != layout.KindInfo {
		t.Fatalf("%+v", m.layout.Widgets)
	}
}

func TestColorStamp(t *testing.T) {
	m := New(layout.Layout{
		Name:   "t",
		Canvas: layout.Canvas{Cols: 40, Rows: 20},
		Widgets: []layout.Widget{{
			ID: "w1", Kind: layout.KindText, X: 1, Y: 1, W: 5, H: 1,
			Fit: layout.FitFixed, Text: &layout.TextSpec{Text: "Hello"},
		}},
	}, nil)
	m.selected = 0
	m.mode = modeColor
	m.colorIdx = 1
	m.brushFg = true
	m.stampAt(1, 1, false)
	if len(m.layout.Widgets[0].Paint) != 1 || m.layout.Widgets[0].Paint[0].Fg.Index != 1 {
		t.Fatalf("%+v", m.layout.Widgets[0].Paint)
	}
	buf := layout.Paint(m.layout, nil)
	cell, ok := buf.At(1, 1)
	if !ok || cell.R != 'H' || cell.Fg.Index != 1 {
		t.Fatalf("%+v", cell)
	}
}

func TestViewRenders(t *testing.T) {
	m := New(layout.Classic(map[string]string{"os.id": "arch", "os.pretty": "Arch Linux", "title": "a@b"}), nil)
	m.width, m.height = 100, 40
	if !strings.Contains(m.View(), "vilefetch") {
		t.Fatal("view missing title")
	}
}

func TestMouseDragMovesWidget(t *testing.T) {
	m := New(layout.Layout{
		Name:   "t",
		Canvas: layout.Canvas{Cols: 80, Rows: 30},
		Widgets: []layout.Widget{{
			ID: "w1", Kind: layout.KindText, X: 2, Y: 2, W: 5, H: 1,
			Fit: layout.FitFixed, Text: &layout.TextSpec{Text: "Hello"},
		}},
	}, nil)
	m.width, m.height = 80, 24
	m.onMouse(tea.MouseMsg{X: 2, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if m.selected != 0 || m.dragKind != "move" {
		t.Fatalf("selected %d kind %s", m.selected, m.dragKind)
	}
	m.onMouse(tea.MouseMsg{X: 5, Y: 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	w := m.layout.Widgets[0]
	if w.X != 5 || w.Y != 3 {
		t.Fatalf("moved to %d,%d", w.X, w.Y)
	}
}
