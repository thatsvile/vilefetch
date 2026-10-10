package store

import (
	"testing"

	"vilefetch/internal/layout"
)

func TestSaveListAndDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	l := layout.Layout{Name: "demo", Canvas: layout.Canvas{Cols: 40, Rows: 20}}
	if err := Save(l); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault("demo"); err != nil {
		t.Fatal(err)
	}
	names, err := List()
	if err != nil || len(names) != 1 || names[0] != "demo" {
		t.Fatalf("%v %v", names, err)
	}
	got, err := DefaultName()
	if err != nil || got != "demo" {
		t.Fatalf("%s %v", got, err)
	}
	loaded, err := Load("demo")
	if err != nil || loaded.Name != "demo" {
		t.Fatalf("%+v %v", loaded, err)
	}
}
