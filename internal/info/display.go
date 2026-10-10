package info

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func displays() map[string]string {
	matches, err := filepath.Glob("/sys/class/drm/card*-*")
	if err != nil || len(matches) == 0 {
		return nil
	}
	sort.Strings(matches)
	out := map[string]string{}
	n := 0
	for _, dir := range matches {
		base := filepath.Base(dir)
		if !strings.Contains(base, "-") {
			continue
		}
		status := readTrim(filepath.Join(dir, "status"))
		if status != "" && !strings.HasPrefix(status, "connected") {
			continue
		}
		if status == "" {
			continue
		}
		name := connectorName(base)
		if edid, err := os.ReadFile(filepath.Join(dir, "edid")); err == nil {
			if parsed := edidName(edid); parsed != "" {
				name = parsed
			}
		}
		res := firstMode(readTrim(filepath.Join(dir, "modes")))
		text := strings.TrimSpace(name + " " + res)
		prefix := "display." + strconv.Itoa(n)
		out[prefix+".name"] = name
		out[prefix+".resolution"] = res
		out[prefix+".text"] = text
		n++
	}
	if n == 0 {
		return nil
	}
	out["display.count"] = strconv.Itoa(n)
	return out
}

func connectorName(base string) string {
	if i := strings.IndexByte(base, '-'); i >= 0 && i+1 < len(base) {
		return base[i+1:]
	}
	return base
}

func firstMode(modes string) string {
	line, _, _ := strings.Cut(modes, "\n")
	return strings.TrimSpace(line)
}

func edidName(edid []byte) string {
	if len(edid) < 128 {
		return ""
	}
	for i := 0; i < 4; i++ {
		off := 54 + i*18
		if edid[off] != 0 || edid[off+1] != 0 || edid[off+3] != 0xFC {
			continue
		}
		raw := edid[off+5 : off+18]
		raw = bytes.Trim(raw, "\x00 ")
		if i := bytes.IndexByte(raw, '\n'); i >= 0 {
			raw = raw[:i]
		}
		return strings.TrimSpace(string(raw))
	}
	return ""
}
