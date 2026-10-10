package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"vilefetch/internal/editor"
	"vilefetch/internal/info"
	"vilefetch/internal/layout"
	"vilefetch/internal/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "vilefetch:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return printLayout("")
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage()
		return nil
	case "layouts":
		return listLayouts()
	case "use":
		if len(args) != 2 {
			usage()
			os.Exit(2)
		}
		return useLayout(args[1])
	case "edit":
		name := ""
		if len(args) > 2 {
			usage()
			os.Exit(2)
		}
		if len(args) == 2 {
			name = args[1]
		}
		return editLayout(name)
	case "--layout", "-l":
		if len(args) != 2 {
			usage()
			os.Exit(2)
		}
		return printLayout(args[1])
	default:
		usage()
		os.Exit(2)
		return nil
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `vilefetch prints a saved layout. vilefetch edit opens the layout editor.

Usage:
  vilefetch                  print the default layout
  vilefetch --layout NAME    print NAME
  vilefetch edit [NAME]      edit NAME, or the default
  vilefetch layouts          list layouts
  vilefetch use NAME         set the default layout

Editor (keyboard first, mouse works too):
  hjkl or arrows     move the selection
  shift+arrows       resize
  p                  palette of info rows and shapes
  enter              pick a look, then drop the live line
  c                  paint glyphs with the terminal palette
  i                  import a .txt  (paste ASCII any time)
  ctrl+s             save
  `+"`"+`                  set this layout as the default
  ?                  all bindings
`)
}

func collect() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return info.Collect(ctx)
}

func printLayout(name string) error {
	if _, err := store.Ensure(); err != nil {
		return err
	}
	if name == "" {
		var err error
		name, err = store.DefaultName()
		if err != nil {
			return err
		}
	}
	l, err := store.Load(name)
	if err != nil {
		return err
	}
	fmt.Println(layout.Render(l, collect()))
	return nil
}

func listLayouts() error {
	if _, err := store.Ensure(); err != nil {
		return err
	}
	names, err := store.List()
	if err != nil {
		return err
	}
	def, err := store.DefaultName()
	if err != nil {
		return err
	}
	for _, name := range names {
		mark := " "
		if name == def {
			mark = "*"
		}
		fmt.Printf("%s %s\n", mark, name)
	}
	return nil
}

func useLayout(name string) error {
	if _, err := store.Ensure(); err != nil {
		return err
	}
	if _, err := store.Load(name); err != nil {
		return err
	}
	return store.SetDefault(name)
}

func editLayout(name string) error {
	if _, err := store.Ensure(); err != nil {
		return err
	}
	if name == "" {
		var err error
		name, err = store.DefaultName()
		if err != nil {
			return err
		}
	}
	if err := store.ValidName(name); err != nil {
		return err
	}
	l, err := store.Load(name)
	if err != nil {
		if os.IsNotExist(err) {
			l = layout.Layout{Name: name, Canvas: layout.Canvas{Cols: 100, Rows: 40}}
		} else {
			return err
		}
	}
	return editor.Run(l, collect())
}
