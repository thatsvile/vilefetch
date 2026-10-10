package layout

// FieldDef is one placeable fetched line.
type FieldDef struct {
	Key      string
	Label    string
	Format   string
	Glyph    string
	Bar      bool
	BarField string
}

// BuiltinFields are the info rows available on every machine.
func BuiltinFields() []FieldDef {
	return []FieldDef{
		{Key: "title", Label: "Title", Format: "{title}", Glyph: "▸"},
		{Key: "os", Label: "OS", Format: "{os.pretty}", Glyph: "▣"},
		{Key: "host", Label: "Host", Format: "{host.model}", Glyph: "▮"},
		{Key: "kernel", Label: "Kernel", Format: "{kernel.release}", Glyph: "▸"},
		{Key: "uptime", Label: "Uptime", Format: "{uptime}", Glyph: "◉"},
		{Key: "shell", Label: "Shell", Format: "{shell.name}", Glyph: "▸"},
		{Key: "terminal", Label: "Terminal", Format: "{terminal.name}", Glyph: "▣"},
		{Key: "desktop", Label: "Desktop", Format: "{desktop.name}", Glyph: "▣"},
		{Key: "cpu", Label: "CPU", Format: "{cpu.text}", Glyph: "▮"},
		{Key: "gpu", Label: "GPU", Format: "{gpu.name}", Glyph: "▣"},
		{Key: "memory", Label: "Memory", Format: "{memory.text}", Glyph: "▮", Bar: true, BarField: "memory.percent"},
		{Key: "swap", Label: "Swap", Format: "{swap.text}", Glyph: "▮", Bar: true, BarField: "swap.percent"},
		{Key: "disk", Label: "Disk", Format: "{disk.text}", Glyph: "▮", Bar: true, BarField: "disk.percent"},
		{Key: "battery", Label: "Battery", Format: "{battery.text}", Glyph: "◉", Bar: true, BarField: "battery.percent"},
	}
}

// Fields appends one entry per connected display.
func Fields(info map[string]string) []FieldDef {
	out := BuiltinFields()
	n := 0
	if info != nil {
		if v, ok := info["display.count"]; ok {
			for _, r := range v {
				if r < '0' || r > '9' {
					n = 0
					break
				}
				n = n*10 + int(r-'0')
			}
		}
	}
	if n == 0 {
		out = append(out, FieldDef{Key: "display.0", Label: "Display", Format: "{display.0.text}", Glyph: "▣"})
		return out
	}
	for i := 0; i < n; i++ {
		key := "display." + itoa(i)
		out = append(out, FieldDef{
			Key:    key,
			Label:  "Display " + itoa(i+1),
			Format: "{" + key + ".text}",
			Glyph:  "▣",
		})
	}
	return out
}

// Appearance is a way to draw an info row.
type Appearance struct {
	Name      string
	ShowLabel bool
	Mark      string
	Sep       string
	Bar       bool
}

// AppearancesFor lists the looks a field can take, each meant to be previewed live.
func AppearancesFor(f FieldDef) []Appearance {
	apps := []Appearance{
		{Name: "plain", ShowLabel: true, Mark: "none", Sep: ": "},
		{Name: "bullet", ShowLabel: true, Mark: "bullet", Sep: "  "},
		{Name: "icon", ShowLabel: false, Mark: "icon", Sep: " "},
		{Name: "arrow", ShowLabel: true, Mark: "none", Sep: " → "},
		{Name: "value", ShowLabel: false, Mark: "none", Sep: ""},
	}
	if f.Bar {
		apps = append(apps, Appearance{Name: "bar", ShowLabel: true, Mark: "icon", Sep: "  ", Bar: true})
	}
	return apps
}

// SpecFrom builds an info row from a field and an appearance.
func SpecFrom(f FieldDef, a Appearance) InfoRowSpec {
	glyph := ""
	if a.Mark == "bullet" {
		glyph = "•"
	}
	if a.Mark == "icon" {
		glyph = f.Glyph
	}
	spec := InfoRowSpec{
		Format:     f.Format,
		Label:      f.Label,
		ShowLabel:  a.ShowLabel,
		Mark:       a.Mark,
		Glyph:      glyph,
		Sep:        a.Sep,
		MarkColor:  &Color{Index: 6},
		LabelColor: &Color{Index: 4},
	}
	if a.Bar && f.Bar {
		spec.Bar = true
		spec.BarField = f.BarField
	}
	return spec
}

// InfoWidget places a live-sized info row. W and H match the rendered line.
func InfoWidget(id string, f FieldDef, a Appearance, x, y int, fields map[string]string) Widget {
	spec := SpecFrom(f, a)
	w := GlyphWidth(InfoGlyphs(spec, fields))
	if w < 1 {
		w = 1
	}
	return Widget{
		ID: id, Kind: KindInfo, X: x, Y: y, W: w, H: 1,
		Fit: FitHug, Push: PushNone, Info: &spec,
	}
}
