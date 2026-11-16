package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSupportsSIGUSR2(t *testing.T) {
	for v, want := range map[string]bool{
		"Ghostty 1.4.0\n\nVersion\n  - version: 1.4.0": true,
		"Ghostty 1.2.0":       true,
		"Ghostty 1.1.3":       false,
		"Ghostty 2.0.0":       true,
		"Ghostty 1.3.0-dev+1": true,
		"":                    false,
	} {
		if got := supportsSIGUSR2(v); got != want {
			t.Errorf("supportsSIGUSR2(%q) = %v", v, got)
		}
	}
}

// fakeProc builds a /proc-like tree: pid -> comm.
func fakeProc(t *testing.T, procs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for pid, comm := range procs {
		dir := filepath.Join(root, pid)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "comm"), []byte(comm+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "self"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestGhosttyPIDs(t *testing.T) {
	root := fakeProc(t, map[string]string{"100": "ghostty", "200": "zsh", "300": "ghostty", "400": "ghostty-tailor"})
	pids := ghosttyPIDs(root, uint32(os.Getuid()))
	sort.Ints(pids)
	if !reflect.DeepEqual(pids, []int{100, 300}) {
		t.Errorf("pids = %v", pids)
	}
	if pids := ghosttyPIDs(root, uint32(os.Getuid())+1); len(pids) != 0 {
		t.Errorf("other users' processes must be skipped, got %v", pids)
	}
}

func stubSignal(t *testing.T, version string, procs map[string]string, send func(int) error) *[]int {
	t.Helper()
	var sent []int
	oRoot, oSend, oVer := procRoot, sendSignal, ghosttyVer
	t.Cleanup(func() { procRoot, sendSignal, ghosttyVer = oRoot, oSend, oVer })
	procRoot = fakeProc(t, procs)
	ghosttyVer = func() string { return version }
	sendSignal = func(pid int) error { sent = append(sent, pid); return send(pid) }
	return &sent
}

func TestReloadBySignal(t *testing.T) {
	ok := func(int) error { return nil }

	sent := stubSignal(t, "Ghostty 1.4.0", map[string]string{"100": "ghostty", "200": "zsh"}, ok)
	if r, id, _ := reloadBySignal(); !r || id != "" || !reflect.DeepEqual(*sent, []int{100}) {
		t.Errorf("reload = %v %q, sent %v", r, id, *sent)
	}

	sent = stubSignal(t, "Ghostty 1.1.3", map[string]string{"100": "ghostty"}, ok)
	if r, id, _ := reloadBySignal(); r || id != "linuxNoReload" || len(*sent) != 0 {
		t.Errorf("old Ghostty must not be signalled: %v %q, sent %v", r, id, *sent)
	}

	stubSignal(t, "Ghostty 1.4.0", map[string]string{"200": "zsh"}, ok)
	if r, id, _ := reloadBySignal(); r || id != "notRunning" {
		t.Errorf("no Ghostty: %v %q", r, id)
	}

	stubSignal(t, "Ghostty 1.4.0", map[string]string{"100": "ghostty"}, func(int) error { return errors.New("operation not permitted") })
	if r, id, detail := reloadBySignal(); r || id != "reloadFailed" || detail != "operation not permitted" {
		t.Errorf("failed signal: %v %q %q", r, id, detail)
	}
}
