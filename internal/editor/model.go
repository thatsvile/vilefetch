package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"vilefetch/internal/layout"
)

type mode int

const (
	modeNormal mode = iota
	modePalette
	modeAppearance
	modePlace
	modeColor
	modePrompt
	modeProps
	modePreview
	modeQuit
)

type intent int

const (
	intentPlace intent = iota
	intentAddRow
)

// Model is the layout editor.
type Model struct {
	layout layout.Layout
	fields map[string]string
	mode   mode

	selected         int
	cursorX, cursorY int
	camX, camY       int
	width, height    int

	items    []palItem
	filtered []int
	palIdx   int
	filter   string
	palOff   int

	field  layout.FieldDef
	apps   []layout.Appearance
	appIdx int
	appOff int
	ghost  *layout.Widget
	intent intent

	brushFg  bool
	colorIdx int
	stroking bool

	undo   [][]layout.Widget
	redo   [][]layout.Widget
	dirty  bool
	status string

	showHelp bool
	help     help.Model

	prompt     textinput.Model
	promptKind string

	props    []propField
	propIdx  int
	propSnap []layout.Widget

	dragging bool
	dragKind string
	dragUndo bool
	originX  int
	originY  int
	originW  int
	originH  int
	pressX   int
	pressY   int

	seq int
}

type propField struct {
	Name string
	Enum []string
	Get  func() string
	Set  func(string)
}

// New builds an editor for an existing or empty layout.
func New(l layout.Layout, fields map[string]string) Model {
	if fields == nil {
		fields = map[string]string{}
	}
	if l.Canvas.Cols < 40 {
		l.Canvas.Cols = 100
	}
	if l.Canvas.Rows < 20 {
		l.Canvas.Rows = 40
	}
	m := Model{
		layout:   l,
		fields:   fields,
		selected: -1,
		brushFg:  true,
		colorIdx: 2,
		help:     help.New(),
		status:   "p palette  c color  ? help",
	}
	m.seq = 1
	for _, w := range l.Widgets {
		if strings.HasPrefix(w.ID, "w") {
			if n, err := strconv.Atoi(strings.TrimPrefix(w.ID, "w")); err == nil && n >= m.seq {
				m.seq = n + 1
			}
		}
	}
	m.rebuildPalette()
	return m
}

// Run opens the editor on the alternate screen.
func Run(l layout.Layout, fields map[string]string) error {
	m := New(l, fields)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) nextID() string {
	id := fmt.Sprintf("w%d", m.seq)
	m.seq++
	return id
}
