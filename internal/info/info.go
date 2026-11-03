// Package info collects the fields layouts can print.
package info

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Collect gathers every v1 field. Individual probes fail soft.
func Collect(ctx context.Context) map[string]string {
	out := map[string]string{}
	merge(out, identity())
	merge(out, operatingSystem())
	merge(out, kernel())
	merge(out, uptime(ctx))
	merge(out, cpuInfo(ctx))
	merge(out, memory(ctx))
	merge(out, rootDisk(ctx))
	merge(out, gpuName())
	merge(out, displays())
	merge(out, battery())
	merge(out, session())
	merge(out, packages(ctx))
	return out
}

func merge(dst, src map[string]string) {
	for k, v := range src {
		if v != "" {
			dst[k] = v
		}
	}
}

func identity() map[string]string {
	name := os.Getenv("USER")
	if u, err := user.Current(); err == nil && u.Username != "" {
		name = u.Username
	}
	hostName, _ := os.Hostname()
	title := name
	if hostName != "" {
		title = name + "@" + hostName
	}
	return map[string]string{
		"user.name":  name,
		"host.name":  hostName,
		"title":      title,
		"host.model": readTrim("/sys/class/dmi/id/product_name"),
	}
}

func operatingSystem() map[string]string {
	vals := parseOSRelease("/etc/os-release")
	pretty := vals["PRETTY_NAME"]
	if pretty == "" {
		pretty = strings.TrimSpace(vals["NAME"] + " " + vals["VERSION"])
	}
	return map[string]string{
		"os.id":      vals["ID"],
		"os.name":    vals["NAME"],
		"os.version": vals["VERSION_ID"],
		"os.pretty":  pretty,
	}
}

func parseOSRelease(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[k] = strings.Trim(v, `"'`)
	}
	return out
}

func kernel() map[string]string {
	rel := readTrim("/proc/sys/kernel/osrelease")
	return map[string]string{
		"kernel.name":    "Linux",
		"kernel.release": rel,
	}
}

func uptime(ctx context.Context) map[string]string {
	inf, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil
	}
	return map[string]string{"uptime": formatUptime(inf.Uptime)}
}

func formatUptime(sec uint64) string {
	d := time.Duration(sec) * time.Second
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	mins := d / time.Minute
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

func cpuInfo(ctx context.Context) map[string]string {
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil || len(infos) == 0 {
		return nil
	}
	cores, _ := cpu.CountsWithContext(ctx, true)
	model := strings.TrimSpace(infos[0].ModelName)
	mhz := infos[0].Mhz
	text := model
	if cores > 0 {
		text = fmt.Sprintf("%s (%d)", model, cores)
	}
	if mhz > 0 {
		text = fmt.Sprintf("%s @ %.2f GHz", text, mhz/1000)
	}
	return map[string]string{
		"cpu.model": model,
		"cpu.cores": strconv.Itoa(cores),
		"cpu.text":  text,
	}
}

func memory(ctx context.Context) map[string]string {
	out := map[string]string{}
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		out["memory.used"] = formatBytes(vm.Used)
		out["memory.total"] = formatBytes(vm.Total)
		out["memory.percent"] = strconv.Itoa(int(vm.UsedPercent + 0.5))
		out["memory.text"] = out["memory.used"] + " / " + out["memory.total"]
	}
	if sw, err := mem.SwapMemoryWithContext(ctx); err == nil && sw.Total > 0 {
		out["swap.used"] = formatBytes(sw.Used)
		out["swap.total"] = formatBytes(sw.Total)
		out["swap.percent"] = strconv.Itoa(int(sw.UsedPercent + 0.5))
		out["swap.text"] = out["swap.used"] + " / " + out["swap.total"]
	}
	return out
}

func rootDisk(ctx context.Context) map[string]string {
	usage, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil
	}
	pct := strconv.Itoa(int(usage.UsedPercent + 0.5))
	return map[string]string{
		"disk.used":    formatBytes(usage.Used),
		"disk.total":   formatBytes(usage.Total),
		"disk.percent": pct,
		"disk.text":    formatBytes(usage.Used) + " / " + formatBytes(usage.Total),
	}
}

func gpuName() map[string]string {
	out, err := exec.Command("lspci", "-mm").Output()
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		fields := quotedFields(line)
		if len(fields) < 3 {
			continue
		}
		class := fields[0]
		if strings.Contains(class, "VGA") || strings.Contains(class, "3D") || strings.Contains(class, "Display") {
			names = append(names, strings.TrimSpace(fields[1]+" "+fields[2]))
		}
	}
	if len(names) == 0 {
		return nil
	}
	if len(names) > 2 {
		names = names[:2]
	}
	return map[string]string{"gpu.name": strings.Join(names, ", ")}
}

func quotedFields(s string) []string {
	var f []string
	for {
		i := strings.IndexByte(s, '"')
		if i < 0 {
			break
		}
		s = s[i+1:]
		j := strings.IndexByte(s, '"')
		if j < 0 {
			break
		}
		f = append(f, s[:j])
		s = s[j+1:]
	}
	return f
}

func battery() map[string]string {
	matches, _ := filepath.Glob("/sys/class/power_supply/BAT*")
	for _, dir := range matches {
		cap := readTrim(filepath.Join(dir, "capacity"))
		if cap == "" {
			continue
		}
		status := readTrim(filepath.Join(dir, "status"))
		text := cap + "%"
		if status != "" {
			text = text + " " + status
		}
		return map[string]string{
			"battery.percent": cap,
			"battery.status":  status,
			"battery.text":    text,
		}
	}
	return nil
}

func session() map[string]string {
	shell := filepath.Base(os.Getenv("SHELL"))
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = os.Getenv("XDG_SESSION_DESKTOP")
	}
	return map[string]string{
		"shell.name":    shell,
		"terminal.name": terminalName(),
		"desktop.name":  desktop,
	}
}

func terminalName() string {
	if v := os.Getenv("TERM_PROGRAM"); v != "" {
		return v
	}
	if v := os.Getenv("LC_TERMINAL"); v != "" {
		return v
	}
	pid := os.Getpid()
	shells := map[string]bool{"bash": true, "zsh": true, "fish": true, "sh": true, "dash": true, "nu": true, "elvish": true, "vilefetch": true}
	for i := 0; i < 8; i++ {
		comm := readTrim(fmt.Sprintf("/proc/%d/comm", pid))
		ppid := parentPID(pid)
		if comm != "" && !shells[comm] && pid != os.Getpid() {
			return comm
		}
		if ppid <= 1 || ppid == pid {
			if comm != "" && !shells[comm] {
				return comm
			}
			break
		}
		pid = ppid
	}
	if t := os.Getenv("TERM"); t != "" {
		return t
	}
	return ""
}

func parentPID(pid int) int {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0
	}
	// pid (comm) state ppid ...
	s := string(data)
	i := strings.LastIndex(s, ")")
	if i < 0 || i+2 >= len(s) {
		return 0
	}
	rest := strings.Fields(s[i+2:])
	if len(rest) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(rest[1])
	return n
}

func readTrim(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func formatBytes(n uint64) string {
	const gi = 1024 * 1024 * 1024
	const mi = 1024 * 1024
	if n >= gi {
		return fmt.Sprintf("%.1f GiB", float64(n)/float64(gi))
	}
	if n >= mi {
		return fmt.Sprintf("%d MiB", n/mi)
	}
	return fmt.Sprintf("%d B", n)
}
