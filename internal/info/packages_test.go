package info

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCountCmdLines(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, ok := countCmdLines(ctx, "sh", "-c", "printf 'a\\nb\\n\\nc\\n'")
	if !ok {
		t.Fatal("expected ok")
	}
	if n != 3 {
		t.Fatalf("got %d", n)
	}
}

func TestCountPacmanLocal(t *testing.T) {
	if _, err := os.Stat("/var/lib/pacman/local"); err != nil {
		t.Skip("no pacman local db")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, ok := countPacman(ctx)
	if !ok {
		t.Fatal("expected ok")
	}
	if n < 1 {
		t.Fatalf("expected packages, got %d", n)
	}
}

func TestCountDpkgStatusLines(t *testing.T) {
	body := "Package: a\nStatus: install ok installed\n\nPackage: b\nStatus: deinstall ok config-files\n\nPackage: c\nStatus: install ok installed\n"
	n := 0
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		if sc.Text() == "Status: install ok installed" {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("got %d", n)
	}
}

func TestPackagesCollectShape(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := packages(ctx)
	if out == nil {
		t.Skip("no package managers detected")
	}
	if out["packages.text"] == "" || out["packages.total"] == "" {
		t.Fatalf("%v", out)
	}
	hasOne := false
	for _, m := range packageManagers {
		if out["packages."+m.ID] != "" {
			hasOne = true
			break
		}
	}
	if !hasOne {
		t.Fatalf("no per-manager keys: %v", out)
	}
}
