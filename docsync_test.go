package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func syncFixture() (map[string]string, map[string]docTranslation) {
	docs := map[string]string{
		"font-size":   "Font size in points.",
		"theme":       "Theme to use. Changed in a newer Ghostty.",
		"window-save": "Whether to save window state.",
	}
	ja := map[string]docTranslation{
		"font-size": {Src: docHash("Font size in points."), Doc: "フォントサイズ（ポイント）です。"},
		"theme":     {Src: docHash("Theme to use."), Doc: "使用するテーマです。"},
		"old-key":   {Src: "0000000000000000", Doc: "なくなった項目です。"},
	}
	return docs, ja
}

func TestSyncStatus(t *testing.T) {
	docs, ja := syncFixture()
	s := syncStatus(docs, ja)
	if s.UpToDate != 1 || !reflect.DeepEqual(s.Stale, []string{"theme"}) ||
		!reflect.DeepEqual(s.Missing, []string{"window-save"}) || !reflect.DeepEqual(s.Obsolete, []string{"old-key"}) {
		t.Errorf("status = %+v", s)
	}
}

func TestExportTodo(t *testing.T) {
	docs, ja := syncFixture()
	want := map[string]todoEntry{
		"theme":       {En: docs["theme"], Previous: "使用するテーマです。"},
		"window-save": {En: docs["window-save"]},
	}
	if got := exportTodo(docs, ja); !reflect.DeepEqual(got, want) {
		t.Errorf("todo = %+v", got)
	}
}

func TestMergeTodo(t *testing.T) {
	docs, ja := syncFixture()
	todo := exportTodo(docs, ja)
	todo["theme"] = todoEntry{En: docs["theme"], Ja: "使用するテーマです（新しい説明）。"}
	todo["gone"] = todoEntry{En: "x", Ja: "x"}
	todo["font-size"] = todoEntry{En: "Font size.", Ja: "古い英文からの訳"}
	merged, skipped := mergeTodo(docs, ja, todo)
	if merged != 1 {
		t.Errorf("merged = %d", merged)
	}
	wantSkipped := []string{
		"font-size: English doc changed since export",
		"gone: Ghostty has no such option",
		"window-save: not translated",
	}
	if !reflect.DeepEqual(skipped, wantSkipped) {
		t.Errorf("skipped = %v", skipped)
	}
	if got := ja["theme"]; got.Src != docHash(docs["theme"]) || got.Doc != "使用するテーマです（新しい説明）。" {
		t.Errorf("theme = %+v", got)
	}
	if got := ja["font-size"].Doc; got != "フォントサイズ（ポイント）です。" {
		t.Errorf("font-size should be unchanged, got %q", got)
	}
	if s := syncStatus(docs, ja); !reflect.DeepEqual(s.Stale, []string(nil)) {
		t.Errorf("theme should be up to date: %+v", s)
	}
}

func TestPruneObsolete(t *testing.T) {
	docs, ja := syncFixture()
	if removed := pruneObsolete(docs, ja); !reflect.DeepEqual(removed, []string{"old-key"}) {
		t.Errorf("removed = %v", removed)
	}
	if _, ok := ja["old-key"]; ok {
		t.Errorf("old-key still present")
	}
}

// The bundled file must survive a read/write round trip unchanged, so merging
// only shows the entries that really changed in a diff.
func TestWriteJSONKeepsLayout(t *testing.T) {
	orig, err := os.ReadFile("i18n/docs.ja.json")
	if err != nil {
		t.Fatal(err)
	}
	ja, err := readTranslations("i18n/docs.ja.json")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "docs.ja.json")
	if err := writeJSON(out, ja); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(out); string(got) != string(orig) {
		t.Errorf("round trip changed docs.ja.json")
	}
}
