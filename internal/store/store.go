// Package store reads and writes layouts under the XDG config directory.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"vilefetch/internal/info"
	"vilefetch/internal/layout"
)

var nameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// Config selects the default layout.
type Config struct {
	Default string `json:"default"`
}

// Dir is $XDG_CONFIG_HOME/vilefetch or ~/.config/vilefetch.
func Dir() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "vilefetch"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "vilefetch"), nil
}

func ValidName(name string) error {
	if !nameRE.MatchString(name) {
		return fmt.Errorf("invalid layout name %q", name)
	}
	return nil
}

// Ensure creates the config directory and seeds classic on first run.
func Ensure() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(dir, "layouts"), 0o755); err != nil {
		return "", err
	}
	names, err := List()
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		fields := info.Collect(context.Background())
		classic := layout.Classic(fields)
		if err := Save(classic); err != nil {
			return "", err
		}
		if err := SetDefault("classic"); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// List returns saved layout names.
func List() ([]string, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, "layouts"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		names = append(names, e.Name()[:len(e.Name())-5])
	}
	sort.Strings(names)
	return names, nil
}

// Load reads one layout.
func Load(name string) (layout.Layout, error) {
	if err := ValidName(name); err != nil {
		return layout.Layout{}, err
	}
	dir, err := Dir()
	if err != nil {
		return layout.Layout{}, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "layouts", name+".json"))
	if err != nil {
		return layout.Layout{}, err
	}
	var l layout.Layout
	if err := json.Unmarshal(b, &l); err != nil {
		return layout.Layout{}, err
	}
	if l.Name == "" {
		l.Name = name
	}
	return l, nil
}

// Save writes a layout under its name.
func Save(l layout.Layout) error {
	if err := ValidName(l.Name); err != nil {
		return err
	}
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "layouts"), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(filepath.Join(dir, "layouts", l.Name+".json"), b, 0o644)
}

// ConfigLoad reads config.json.
func ConfigLoad() (Config, error) {
	dir, err := Dir()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{Default: "classic"}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.Default == "" {
		c.Default = "classic"
	}
	return c, nil
}

// SetDefault points config.json at name.
func SetDefault(name string) error {
	if err := ValidName(name); err != nil {
		return err
	}
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(Config{Default: name}, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(filepath.Join(dir, "config.json"), b, 0o644)
}

// DefaultName is the layout printed by a bare invocation.
func DefaultName() (string, error) {
	c, err := ConfigLoad()
	if err != nil {
		return "", err
	}
	return c.Default, nil
}
