package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// withClock makes snapshot names predictable: each call is one second later.
func withClock(t *testing.T) {
	t.Helper()
	cur := time.Date(2026, 11, 16, 10, 0, 0, 0, time.Local)
	orig := now
	now = func() time.Time { cur = cur.Add(time.Second); return cur }
	t.Cleanup(func() { now = orig })
}

func historyDir(t *testing.T) string {
	t.Helper()
	return historyDirFunc()
}

func TestApplySavesHistory(t *testing.T) {
	a := newTestApp(t, "font-size = 13\n", 0, reloadOK)
	withClock(t)
	a.SetOption("font-size", "14")
	a.SetOption("theme", "Nord")
	a.SetOption("theme", "Nord") // no change: no new entry

	h := a.History()
	if len(h) != 2 {
		t.Fatalf("history = %+v", h)
	}
	if !reflect.DeepEqual(h[0].Keys, []string{"theme"}) || !reflect.DeepEqual(h[1].Keys, []string{"font-size"}) {
		t.Errorf("keys = %v, %v", h[0].Keys, h[1].Keys)
	}
	if got := readFile(t, filepath.Join(historyDir(t), h[1].ID+".ghostty")); got != "font-size = 13\n" {
		t.Errorf("oldest snapshot = %q", got)
	}
}

func TestRejectedChangeLeavesNoHistory(t *testing.T) {
	a := newTestApp(t, "font-size = 13\n", 1, reloadOK)
	withClock(t)
	if res := a.SetOption("font-size", "x"); res.OK {
		t.Fatalf("should be rejected: %+v", res)
	}
	if h := a.History(); len(h) != 0 {
		t.Errorf("history = %+v", h)
	}
}

func TestRestoreHistoryCanBeUndone(t *testing.T) {
	a := newTestApp(t, "# mine\nfont-size = 13\n", 0, reloadOK)
	withClock(t)
	a.SetOption("font-size", "20")
	first := a.History()[0].ID

	if res := a.RestoreHistory(first); !res.OK {
		t.Fatalf("restore: %+v", res)
	}
	if got := readFile(t, a.cfg.Path); got != "# mine\nfont-size = 13\n" {
		t.Errorf("config = %q", got)
	}
	h := a.History()
	if len(h) != 2 {
		t.Fatalf("restore should add an entry: %+v", h)
	}
	if res := a.RestoreHistory(h[0].ID); !res.OK || readFile(t, a.cfg.Path) != "# mine\nfont-size = 20\n" {
		t.Errorf("undo restore: %+v, config %q", res, readFile(t, a.cfg.Path))
	}
	for _, id := range []string{"nope", "../x", ""} {
		if res := a.RestoreHistory(id); res.OK {
			t.Errorf("RestoreHistory(%q) should fail", id)
		}
	}
}

func TestHistoryKeepsNewest(t *testing.T) {
	a := newTestApp(t, "", 0, reloadOK)
	withClock(t)
	for i := 0; i < historyLimit+5; i++ {
		a.SetOption("font-size", strings.Repeat("1", i+1))
	}
	ids := historyIDs(historyDir(t))
	if len(ids) != historyLimit {
		t.Fatalf("kept %d entries", len(ids))
	}
	newest := readFile(t, filepath.Join(historyDir(t), ids[0]+".ghostty"))
	if newest != "font-size = "+strings.Repeat("1", historyLimit+4)+"\n" {
		t.Errorf("newest = %q", newest)
	}
}

func TestExportConfig(t *testing.T) {
	a := newTestApp(t, "# mine\ntheme = Nord\n", 0, reloadOK)
	out := filepath.Join(t.TempDir(), "exported")
	orig := saveDialog
	t.Cleanup(func() { saveDialog = orig })

	saveDialog = func(*App, runtime.SaveDialogOptions) (string, error) { return "", nil }
	if res := a.ExportConfig(); res.OK || res.Message != "" {
		t.Errorf("cancel: %+v", res)
	}
	saveDialog = func(*App, runtime.SaveDialogOptions) (string, error) { return out, nil }
	if res := a.ExportConfig(); !res.OK {
		t.Fatalf("export: %+v", res)
	}
	if got := readFile(t, out); got != "# mine\ntheme = Nord\n" {
		t.Errorf("exported = %q", got)
	}
}

func TestImportConfig(t *testing.T) {
	a := newTestApp(t, "theme = Nord\nfont-size = 13\n", 0, reloadOK)
	withClock(t)
	in := filepath.Join(t.TempDir(), "other")
	if err := os.WriteFile(in, []byte("# from my other Mac\ntheme = Dracula\nfont-family = Menlo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pr := a.previewImport(in)
	want := []ConfigChange{
		{Key: "font-family", To: "Menlo"},
		{Key: "font-size", From: "13"},
		{Key: "theme", From: "Nord", To: "Dracula"},
	}
	if pr.Error != "" || !reflect.DeepEqual(pr.Changes, want) {
		t.Errorf("preview = %+v", pr)
	}
	if res := a.ImportConfig(in); !res.OK {
		t.Fatalf("import: %+v", res)
	}
	if got := readFile(t, a.cfg.Path); got != "# from my other Mac\ntheme = Dracula\nfont-family = Menlo\n" {
		t.Errorf("config = %q", got)
	}
	if h := a.History(); len(h) != 1 {
		t.Errorf("import should be undoable: %+v", h)
	}
	if res := a.ImportConfig(filepath.Join(t.TempDir(), "missing")); res.OK {
		t.Errorf("a missing file must not empty the config")
	}
}

func TestPreviewImportRejectsInvalidFile(t *testing.T) {
	a := newTestApp(t, "theme = Nord\n", 1, reloadOK)
	in := filepath.Join(t.TempDir(), "bad")
	if err := os.WriteFile(in, []byte("font-size = x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if pr := a.previewImport(in); !strings.HasPrefix(pr.Error, msg("en", "importInvalid")) {
		t.Errorf("preview = %+v", pr)
	}
}
