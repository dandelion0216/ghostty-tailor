package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// On Linux there is no AppleScript. Ghostty 1.2 and later reload their config
// when they receive SIGUSR2; older versions would be terminated by it, so the
// version is checked first.

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)`)

// supportsSIGUSR2 reports whether `ghostty +version` output is 1.2 or later.
func supportsSIGUSR2(version string) bool {
	m := versionRe.FindStringSubmatch(version)
	if m == nil {
		return false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major > 1 || (major == 1 && minor >= 2)
}

// ghosttyPIDs returns the processes named "ghostty" owned by uid, read from a
// /proc-style directory.
func ghosttyPIDs(procRoot string, uid uint32) []int {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || !e.IsDir() {
			continue
		}
		dir := filepath.Join(procRoot, e.Name())
		comm, err := os.ReadFile(filepath.Join(dir, "comm"))
		if err != nil || strings.TrimSpace(string(comm)) != "ghostty" {
			continue
		}
		if fi, err := os.Stat(dir); err != nil || ownerUID(fi) != uid {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}

func ownerUID(fi os.FileInfo) uint32 {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return st.Uid
	}
	return ^uint32(0)
}

// Replaced in tests.
var (
	procRoot   = "/proc"
	sendSignal = func(pid int) error { return syscall.Kill(pid, syscall.SIGUSR2) }
	ghosttyVer = func() string {
		bin, err := findGhostty()
		if err != nil {
			return ""
		}
		out, _ := runGhostty(bin, "+version")
		return out
	}
)

func reloadBySignal() (bool, string, string) {
	if !supportsSIGUSR2(ghosttyVer()) {
		return false, "linuxNoReload", ""
	}
	pids := ghosttyPIDs(procRoot, uint32(os.Getuid()))
	if len(pids) == 0 {
		return false, "notRunning", ""
	}
	sent := 0
	var lastErr error
	for _, pid := range pids {
		if err := sendSignal(pid); err != nil {
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 {
		return false, "reloadFailed", lastErr.Error()
	}
	return true, "", ""
}
