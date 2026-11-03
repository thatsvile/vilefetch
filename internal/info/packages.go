package info

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// packageManager is one countable install source.
type packageManager struct {
	ID    string
	Label string
	Count func(ctx context.Context) (int, bool)
}

var packageManagers = []packageManager{
	{ID: "pacman", Label: "pacman", Count: countPacman},
	{ID: "flatpak", Label: "flatpak", Count: countFlatpak},
	{ID: "dpkg", Label: "dpkg", Count: countDpkg},
	{ID: "rpm", Label: "rpm", Count: countRPM},
	{ID: "snap", Label: "snap", Count: countSnap},
	{ID: "nix", Label: "nix", Count: countNix},
	{ID: "brew", Label: "brew", Count: countBrew},
}

func packages(ctx context.Context) map[string]string {
	type result struct {
		id    string
		label string
		n     int
	}
	ch := make(chan result, len(packageManagers))
	var wg sync.WaitGroup
	for _, m := range packageManagers {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, ok := m.Count(ctx)
			if !ok || n <= 0 {
				return
			}
			ch <- result{id: m.ID, label: m.Label, n: n}
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	found := map[string]int{}
	for r := range ch {
		found[r.id] = r.n
	}
	if len(found) == 0 {
		return nil
	}

	out := map[string]string{}
	var parts []string
	total := 0
	for _, m := range packageManagers {
		n, ok := found[m.ID]
		if !ok {
			continue
		}
		out["packages."+m.ID] = strconv.Itoa(n)
		parts = append(parts, fmt.Sprintf("%d (%s)", n, m.Label))
		total += n
	}
	out["packages.total"] = strconv.Itoa(total)
	out["packages.text"] = strings.Join(parts, ", ")
	return out
}

func countPacman(ctx context.Context) (int, bool) {
	const local = "/var/lib/pacman/local"
	entries, err := os.ReadDir(local)
	if err == nil {
		n := 0
		for _, e := range entries {
			if e.IsDir() && e.Name() != "" {
				n++
			}
		}
		return n, true
	}
	if _, err := exec.LookPath("pacman"); err != nil {
		return 0, false
	}
	return countCmdLines(ctx, "pacman", "-Qq")
}

func countFlatpak(ctx context.Context) (int, bool) {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return 0, false
	}
	// Apps only; runtimes inflate the number without matching user expectation.
	return countCmdLines(ctx, "flatpak", "list", "--app", "--columns=application")
}

func countDpkg(ctx context.Context) (int, bool) {
	f, err := os.Open("/var/lib/dpkg/status")
	if err != nil {
		if _, err := exec.LookPath("dpkg-query"); err != nil {
			return 0, false
		}
		return countCmdLines(ctx, "dpkg-query", "-f", "${Package}\n", "-W")
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if sc.Text() == "Status: install ok installed" {
			n++
		}
	}
	return n, true
}

func countRPM(ctx context.Context) (int, bool) {
	if _, err := exec.LookPath("rpm"); err != nil {
		return 0, false
	}
	return countCmdLines(ctx, "rpm", "-qa", "--qf", "%{NAME}\n")
}

func countSnap(ctx context.Context) (int, bool) {
	if _, err := exec.LookPath("snap"); err != nil {
		return 0, false
	}
	n, ok := countCmdLines(ctx, "snap", "list")
	if !ok || n == 0 {
		return n, ok
	}
	// First line is the header.
	if n > 0 {
		n--
	}
	return n, true
}

func countNix(ctx context.Context) (int, bool) {
	if _, err := exec.LookPath("nix-store"); err != nil {
		return 0, false
	}
	// Closures under the current profile; cheap enough for a fetch tool.
	profile := os.Getenv("HOME") + "/.nix-profile"
	if _, err := os.Stat(profile); err != nil {
		profile = "/run/current-system/sw"
		if _, err := os.Stat(profile); err != nil {
			return 0, false
		}
	}
	return countCmdLines(ctx, "nix-store", "-q", "--requisites", profile)
}

func countBrew(ctx context.Context) (int, bool) {
	brew, err := exec.LookPath("brew")
	if err != nil {
		for _, p := range []string{"/home/linuxbrew/.linuxbrew/bin/brew", "/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
			if st, e := os.Stat(p); e == nil && !st.IsDir() {
				brew = p
				err = nil
				break
			}
		}
	}
	if err != nil || brew == "" {
		return 0, false
	}
	formulae, ok1 := countCmdLines(ctx, brew, "list", "--formula")
	casks, ok2 := countCmdLines(ctx, brew, "list", "--cask")
	if !ok1 && !ok2 {
		return 0, false
	}
	n := 0
	if ok1 {
		n += formulae
	}
	if ok2 {
		n += casks
	}
	return n, true
}

func countCmdLines(ctx context.Context, name string, args ...string) (int, bool) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	n := 0
	for len(out) > 0 {
		i := bytes.IndexByte(out, '\n')
		if i < 0 {
			if len(bytes.TrimSpace(out)) > 0 {
				n++
			}
			break
		}
		if len(bytes.TrimSpace(out[:i])) > 0 {
			n++
		}
		out = out[i+1:]
	}
	return n, true
}
