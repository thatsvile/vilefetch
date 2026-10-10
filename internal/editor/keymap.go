package editor

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	short []key.Binding
	full  [][]key.Binding
}

func (k keyMap) ShortHelp() []key.Binding { return k.short }

func (k keyMap) FullHelp() [][]key.Binding { return k.full }

func bind(keys, helpKey, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(keys), key.WithHelp(helpKey, desc))
}

func normalKeys() keyMap {
	move := bind("left", "hjkl", "move")
	resize := bind("shift+left", "S-arrows", "resize")
	pal := bind("p", "p", "palette")
	color := bind("c", "c", "color")
	save := bind("ctrl+s", "ctrl+s", "save")
	quit := bind("q", "q", "quit")
	return keyMap{
		short: []key.Binding{move, resize, pal, color, save, quit},
		full: [][]key.Binding{
			{move, resize, bind("tab", "tab", "cycle"), bind("enter", "enter", "edit")},
			{pal, bind("i", "i", "import"), bind("a", "a", "add row"), color},
			{bind("f", "f", "fit"), bind("g", "g", "push"), bind("v", "v", "look"), bind("[", "[ ]", "layer")},
			{bind("d", "d", "duplicate"), bind("x", "x", "delete"), bind("u", "u", "undo"), bind("ctrl+r", "ctrl+r", "redo")},
			{save, bind("ctrl+p", "ctrl+p", "preview"), bind("`", "`", "set default"), quit},
		},
	}
}

func paletteKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("up", "↑↓", "pick"),
			bind("enter", "enter", "arm"),
			bind("backspace", "bs", "filter"),
			bind("esc", "esc", "close"),
		},
	}
}

func appearanceKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("up", "↑↓", "look"),
			bind("enter", "enter", "take"),
			bind("esc", "esc", "back"),
		},
	}
}

func placeKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("left", "hjkl", "move"),
			bind("enter", "enter", "drop"),
			bind("esc", "esc", "cancel"),
		},
	}
}

func colorKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("left", "hjkl", "brush"),
			bind("1", "1-8", "color"),
			bind("!", "S-1-8", "bright"),
			bind("f", "f/b", "fg/bg"),
			bind(" ", "space", "paint"),
			bind("x", "x", "erase"),
			bind("esc", "esc", "done"),
		},
	}
}

func promptKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("enter", "enter", "ok"),
			bind("esc", "esc", "cancel"),
		},
	}
}

func propsKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("tab", "tab", "field"),
			bind("up", "↑↓", "cycle"),
			bind("enter", "enter", "done"),
			bind("esc", "esc", "cancel"),
		},
	}
}

func quitKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("y", "y", "save"),
			bind("n", "n", "discard"),
			bind("esc", "esc", "cancel"),
		},
	}
}

func previewKeys() keyMap {
	return keyMap{
		short: []key.Binding{
			bind("esc", "esc", "back"),
			bind("ctrl+p", "ctrl+p", "back"),
		},
	}
}

func (m Model) activeHelp() keyMap {
	switch m.mode {
	case modePalette:
		return paletteKeys()
	case modeAppearance:
		return appearanceKeys()
	case modePlace:
		return placeKeys()
	case modeColor:
		return colorKeys()
	case modePrompt:
		return promptKeys()
	case modeProps:
		return propsKeys()
	case modeQuit:
		return quitKeys()
	case modePreview:
		return previewKeys()
	default:
		return normalKeys()
	}
}
