package layout

import "strings"

const (
	logoArch = `
      /\
     /  \
    /\   \
   /      \
  /   ,,   \
 /   |  |  -\
/_-''    ''-_\`

	logoDebian = `
  ____
 / __ \
| /  \ |
| \__/ |
 \____/`

	logoUbuntu = `
         _
     ---(_)
 _/  ---  \
(_) |   |
  \  --- _/
     ---(_)`

	logoFedora = `
    ____
   / ___)
  | |
 _| |
/____)`

	logoLinux = `
   .--.
  |o_o |
  |:_/ |
 //   \ \
(|     | )
/ \_   _/ \
\___)=(___/`
)

// Logo returns builtin ASCII and a palette color for an os id.
func Logo(osID string) (string, Color) {
	switch strings.ToLower(osID) {
	case "arch", "archlinux", "endeavouros", "manjaro":
		return trimLogo(logoArch), Color{Index: 6}
	case "debian":
		return trimLogo(logoDebian), Color{Index: 1}
	case "ubuntu", "pop":
		return trimLogo(logoUbuntu), Color{Index: 3}
	case "fedora":
		return trimLogo(logoFedora), Color{Index: 4}
	default:
		return trimLogo(logoLinux), Color{Index: 7}
	}
}

func trimLogo(s string) string {
	s = strings.Trim(s, "\n")
	return s
}

// Classic is the layout seeded on first run.
func Classic(fields map[string]string) Layout {
	osID := ""
	if fields != nil {
		osID = fields["os.id"]
	}
	art, col := Logo(osID)
	lines := linesOf(art)
	lw := maxTextWidth(lines)
	lh := len(lines)
	logo := Widget{
		ID: "logo", Kind: KindArt, X: 0, Y: 0, W: lw, H: lh,
		Fit: FitFixed, Push: PushNone,
		Style: Style{Fg: &col},
		Art:   &ArtSpec{Text: art},
	}
	defs := []FieldDef{
		mustField("title"), mustField("os"), mustField("host"), mustField("kernel"),
		mustField("uptime"), mustField("shell"), mustField("cpu"), mustField("gpu"),
		mustField("memory"), mustField("disk"),
	}
	if fields != nil && fields["display.0.text"] != "" {
		for _, f := range Fields(fields) {
			if strings.HasPrefix(f.Key, "display.") {
				defs = append(defs, f)
			}
		}
	}
	var rows []InfoRowSpec
	for _, d := range defs {
		rows = append(rows, SpecFrom(d, Appearance{Name: "icon", ShowLabel: true, Mark: "icon", Sep: "  "}))
	}
	tableH := len(rows) + 2
	table := Widget{
		ID: "info", Kind: KindTable, X: lw + 2, Y: 0, W: 24, H: tableH,
		Fit: FitHug, Push: PushNone,
		Table: &TableSpec{Border: "rounded", Gap: "  ", Rows: rows},
	}
	colors := Widget{
		ID: "colors", Kind: KindColors, X: 0, Y: max(lh, tableH) + 1, W: 16, H: 1,
		Fit: FitHug, Push: PushNone,
		Swatch: &ColorsSpec{Style: "blocks", Which: "both"},
	}
	return Layout{
		Name:    "classic",
		Canvas:  Canvas{Cols: 100, Rows: 40},
		Widgets: []Widget{logo, table, colors},
	}
}

func mustField(key string) FieldDef {
	for _, f := range BuiltinFields() {
		if f.Key == key {
			return f
		}
	}
	return FieldDef{Key: key, Label: key, Format: "{" + key + "}"}
}
