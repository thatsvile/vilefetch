package layout

import (
	"regexp"
	"strconv"
	"strings"
)

var fieldRE = regexp.MustCompile(`\{([A-Za-z0-9_.]+)\}`)

// Fill replaces {field} tokens. Unknown fields become empty.
func Fill(format string, fields map[string]string) string {
	if format == "" || fields == nil {
		return format
	}
	return fieldRE.ReplaceAllStringFunc(format, func(tok string) string {
		key := tok[1 : len(tok)-1]
		return fields[key]
	})
}

// Bar renders a 10-cell meter from a percent field ("42" or "42%").
func Bar(percent string, width int) string {
	if width < 1 {
		width = 10
	}
	percent = strings.TrimSpace(strings.TrimSuffix(percent, "%"))
	p, err := strconv.ParseFloat(percent, 64)
	if err != nil || p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	filled := int(p / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
