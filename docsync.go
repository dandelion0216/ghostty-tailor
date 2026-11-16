package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// The functions below back the i18nsync developer tool (i18nsync.go), which
// keeps i18n/docs.ja.json in sync with new Ghostty versions.

// todoEntry is one option to translate. Previous holds the outdated
// translation, if any, so a translator can update it instead of starting over.
type todoEntry struct {
	En       string `json:"en"`
	Previous string `json:"previous,omitempty"`
	Ja       string `json:"ja"`
}

type docSyncStatus struct {
	UpToDate int
	Stale    []string // translated from an older English doc
	Missing  []string // no translation yet
	Obsolete []string // translated, but Ghostty no longer has the option
}

func syncStatus(docs map[string]string, ja map[string]docTranslation) docSyncStatus {
	var s docSyncStatus
	for _, k := range sortedKeys(docs) {
		t, ok := ja[k]
		switch {
		case !ok:
			s.Missing = append(s.Missing, k)
		case t.Src != docHash(docs[k]):
			s.Stale = append(s.Stale, k)
		default:
			s.UpToDate++
		}
	}
	for _, k := range sortedKeys(ja) {
		if _, ok := docs[k]; !ok {
			s.Obsolete = append(s.Obsolete, k)
		}
	}
	return s
}

func exportTodo(docs map[string]string, ja map[string]docTranslation) map[string]todoEntry {
	s := syncStatus(docs, ja)
	todo := map[string]todoEntry{}
	for _, k := range append(s.Stale, s.Missing...) {
		todo[k] = todoEntry{En: docs[k], Previous: ja[k].Doc}
	}
	return todo
}

// mergeTodo copies translated entries into ja. An entry is skipped when it is
// still untranslated, or when its English no longer matches the installed
// Ghostty (the doc changed again after export).
func mergeTodo(docs map[string]string, ja map[string]docTranslation, todo map[string]todoEntry) (int, []string) {
	merged := 0
	var skipped []string
	for _, k := range sortedKeys(todo) {
		e := todo[k]
		doc, ok := docs[k]
		switch {
		case e.Ja == "":
			skipped = append(skipped, k+": not translated")
		case !ok:
			skipped = append(skipped, k+": Ghostty has no such option")
		case e.En != doc:
			skipped = append(skipped, k+": English doc changed since export")
		default:
			ja[k] = docTranslation{Src: docHash(doc), Doc: e.Ja}
			merged++
		}
	}
	return merged, skipped
}

func pruneObsolete(docs map[string]string, ja map[string]docTranslation) []string {
	removed := syncStatus(docs, ja).Obsolete
	for _, k := range removed {
		delete(ja, k)
	}
	return removed
}

func readTranslations(path string) (map[string]docTranslation, error) {
	ja := map[string]docTranslation{}
	return ja, readJSON(path, &ja)
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// writeJSON keeps the layout of docs.ja.json: sorted keys, one-space indent,
// no HTML escaping and no trailing newline.
func writeJSON(path string, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, bytes.TrimRight(buf.Bytes(), "\n"), 0o644)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
