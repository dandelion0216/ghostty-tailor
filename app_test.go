package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// newTestApp returns an App editing a temp config, with a fake ghostty binary
// whose `+validate-config` exits with validateExit, and a stubbed reload.
func newTestApp(t *testing.T, config string, validateExit int, reload func() (bool, string, string)) *App {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "home", ".config"))

	bin := filepath.Join(dir, "ghostty")
	script := "#!/bin/sh\necho 'error: invalid value' >&2\nexit " + strconv.Itoa(validateExit) + "\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "config.ghostty")
	if err := os.WriteFile(cfg, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := reloadConfig
	reloadConfig = reload
	t.Cleanup(func() { reloadConfig = orig })
	return &App{bin: bin, cfg: &ConfigFile{Path: cfg}, lang: "en"}
}

func reloadOK() (bool, string, string) { return true, "", "" }

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSetOptionSavesAndReloads(t *testing.T) {
	a := newTestApp(t, "# keep me\nfont-size = 13\n", 0, reloadOK)
	res := a.SetOption("font-size", "15")
	if !res.OK || !res.Reloaded || res.Message != msg("en", "reloaded") {
		t.Errorf("result = %+v", res)
	}
	if got := readFile(t, a.cfg.Path); got != "# keep me\nfont-size = 15\n" {
		t.Errorf("config = %q", got)
	}
	if res.Set["font-size"] != "15" {
		t.Errorf("Set = %v", res.Set)
	}
}

func TestSetOptionRollsBackWhenGhosttyRejects(t *testing.T) {
	const before = "# keep me\nfont-size = 13\n\ntheme = Nord\n"
	reloaded := false
	a := newTestApp(t, before, 1, func() (bool, string, string) { reloaded = true; return true, "", "" })
	res := a.SetOption("font-size", "huge")
	if res.OK {
		t.Fatalf("expected failure, got %+v", res)
	}
	if !strings.HasPrefix(res.Message, msg("en", "rejected")) || !strings.Contains(res.Message, "invalid value") {
		t.Errorf("message = %q", res.Message)
	}
	if got := readFile(t, a.cfg.Path); got != before {
		t.Errorf("config was not restored:\n%q\nwant\n%q", got, before)
	}
	if res.Set["font-size"] != "13" {
		t.Errorf("Set after rollback = %v", res.Set)
	}
	if reloaded {
		t.Errorf("Ghostty should not be reloaded after a rejected change")
	}
}

func TestSetOptionEmptyValueUnsets(t *testing.T) {
	a := newTestApp(t, "font-size = 13\ntheme = Nord\n", 0, reloadOK)
	a.SetOption("font-size", "  ")
	if got := readFile(t, a.cfg.Path); got != "theme = Nord\n" {
		t.Errorf("config = %q", got)
	}
}

func TestResetOption(t *testing.T) {
	a := newTestApp(t, "font-size = 13\ntheme = Nord\n", 0, reloadOK)
	res := a.ResetOption("theme")
	if !res.OK {
		t.Fatalf("result = %+v", res)
	}
	if got := readFile(t, a.cfg.Path); got != "font-size = 13\n" {
		t.Errorf("config = %q", got)
	}
}

func TestSavedButNotReloaded(t *testing.T) {
	a := newTestApp(t, "", 0, func() (bool, string, string) { return false, "notRunning", "" })
	res := a.SetOption("font-size", "15")
	if !res.OK || res.Reloaded || res.Message != msg("en", "notRunning") {
		t.Errorf("result = %+v", res)
	}
}

func TestSetLangChangesMessagesAndPersists(t *testing.T) {
	a := newTestApp(t, "", 0, reloadOK)
	if res := a.SetLang("ja-JP"); !res.OK {
		t.Fatalf("SetLang: %+v", res)
	}
	if a.lang != "ja" || loadSettings().Lang != "ja" {
		t.Errorf("lang = %q, saved = %q", a.lang, loadSettings().Lang)
	}
	if res := a.SetOption("font-size", "15"); res.Message != msg("ja", "reloaded") {
		t.Errorf("message = %q, want Japanese", res.Message)
	}
}
