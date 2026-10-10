// Package layout is the document model, measurement, and cell painter for vilefetch.
package layout

// Fit controls how a widget relates to its content size.
type Fit string

const (
	FitFixed     Fit = "fixed"
	FitHugWidth  Fit = "hugWidth"
	FitHugHeight Fit = "hugHeight"
	FitHug       Fit = "hug"
)

// Push shifts later widgets when a hugging widget grows past its minimum box.
type Push string

const (
	PushNone  Push = "none"
	PushRight Push = "right"
	PushDown  Push = "down"
)

const (
	KindArt       = "art"
	KindText      = "text"
	KindInfo      = "inforow"
	KindTable     = "table"
	KindBox       = "box"
	KindSeparator = "separator"
	KindColors    = "colors"
	KindSpacer    = "spacer"
)

// Color is a terminal palette index, 0 through 15.
type Color struct {
	Index int `json:"index"`
}

// CellPaint is a color stamped onto one cell of a widget.
// Coordinates are columns and rows from the widget's top-left.
// A nil Fg or Bg leaves that channel alone.
type CellPaint struct {
	X  int    `json:"x"`
	Y  int    `json:"y"`
	Fg *Color `json:"fg,omitempty"`
	Bg *Color `json:"bg,omitempty"`
}

// Style is the default color for a widget's glyphs.
type Style struct {
	Fg *Color `json:"fg,omitempty"`
	Bg *Color `json:"bg,omitempty"`
}

// Canvas is the editor grid size. Printing uses the bounding box of widgets.
type Canvas struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// Layout is one saved design.
type Layout struct {
	Name    string   `json:"name"`
	Canvas  Canvas   `json:"canvas"`
	Widgets []Widget `json:"widgets"`
}

// Widget is one placed object.
type Widget struct {
	ID    string      `json:"id"`
	Kind  string      `json:"kind"`
	X     int         `json:"x"`
	Y     int         `json:"y"`
	W     int         `json:"w"`
	H     int         `json:"h"`
	Z     int         `json:"z"`
	Fit   Fit         `json:"fit"`
	Push  Push        `json:"push"`
	Style Style       `json:"style,omitempty"`
	Paint []CellPaint `json:"paint,omitempty"`

	Art    *ArtSpec       `json:"art,omitempty"`
	Text   *TextSpec      `json:"text,omitempty"`
	Info   *InfoRowSpec   `json:"info,omitempty"`
	Table  *TableSpec     `json:"table,omitempty"`
	Box    *BoxSpec       `json:"box,omitempty"`
	Sep    *SeparatorSpec `json:"sep,omitempty"`
	Swatch *ColorsSpec    `json:"swatch,omitempty"`
}

// ArtSpec is imported or pasted ASCII.
type ArtSpec struct {
	Text string `json:"text"`
}

// TextSpec is a short literal label.
type TextSpec struct {
	Text string `json:"text"`
}

// InfoRowSpec is one fetched line: mark, label, separator, value, optional bar.
type InfoRowSpec struct {
	Format     string `json:"format"`
	Label      string `json:"label,omitempty"`
	ShowLabel  bool   `json:"showLabel"`
	Mark       string `json:"mark,omitempty"` // none, bullet, icon
	Glyph      string `json:"glyph,omitempty"`
	Sep        string `json:"sep,omitempty"`
	Bar        bool   `json:"bar,omitempty"`
	BarField   string `json:"barField,omitempty"`
	MarkColor  *Color `json:"markColor,omitempty"`
	LabelColor *Color `json:"labelColor,omitempty"`
	ValueColor *Color `json:"valueColor,omitempty"`
}

// TableSpec aligns several info rows inside a border.
type TableSpec struct {
	Border     string        `json:"border,omitempty"` // none, normal, rounded, thick, double
	RowLines   bool          `json:"rowLines,omitempty"`
	LabelWidth int           `json:"labelWidth,omitempty"`
	Gap        string        `json:"gap,omitempty"`
	Rows       []InfoRowSpec `json:"rows"`
}

// BoxSpec is a bordered block of text or one format string.
type BoxSpec struct {
	Border string `json:"border,omitempty"`
	Text   string `json:"text,omitempty"`
	Format string `json:"format,omitempty"`
}

// SeparatorSpec is a horizontal or vertical rule.
type SeparatorSpec struct {
	Dir  string `json:"dir"` // h or v
	Char string `json:"char,omitempty"`
}

// ColorsSpec draws the terminal palette.
type ColorsSpec struct {
	Style string `json:"style"` // dots, blocks, circles, bar
	Which string `json:"which"` // normal, bright, both
}

// Placement is where a widget lands after fit and push.
type Placement struct {
	X, Y, W, H int
}
