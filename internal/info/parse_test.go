package info

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOSRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte("NAME=\"Arch Linux\"\nID=arch\nPRETTY_NAME=\"Arch Linux\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	vals := parseOSRelease(path)
	if vals["ID"] != "arch" || vals["PRETTY_NAME"] != "Arch Linux" {
		t.Fatalf("%v", vals)
	}
}

func TestEdidName(t *testing.T) {
	edid := make([]byte, 128)
	off := 54
	edid[off+3] = 0xFC
	copy(edid[off+5:], []byte("LG UltraGear\n"))
	if got := edidName(edid); got != "LG UltraGear" {
		t.Fatalf("got %q", got)
	}
}
