package main

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Before each change ghostty-tailor saves the previous config file to the
// history directory, so any earlier state can be restored.

const historyLimit = 50

// HistoryEntry is one saved state: the config as it was just before the change
// made at Time. Keys are the options that change from it.
type HistoryEntry struct {
	ID   string   `json:"id"`
	Time string   `json:"time"` // RFC 3339
	Keys []string `json:"keys"`
}

type ConfigChange struct {
	Key  string `json:"key"`
	From string `json:"from"`
	To   string `json:"to"`
}

type ImportPreview struct {
	Path    string         `json:"path"`
	Changes []ConfigChange `json:"changes"`
	Error   string         `json:"error"`
}

const historyTimeFormat = "20060102-150405.000"

var historyDirFunc = func() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "ghostty-tailor", "history")
}

// Replaced in tests; the real ones open native dialogs.
var (
	saveDialog = func(a *App, o runtime.SaveDialogOptions) (string, error) { return runtime.SaveFileDialog(a.ctx, o) }
	openDialog = func(a *App, o runtime.OpenDialogOptions) (string, error) { return runtime.OpenFileDialog(a.ctx, o) }
)

var now = time.Now

// saveSnapshot stores lines as a new history entry and returns its path.
func saveSnapshot(lines []string) (string, error) {
	dir := historyDirFunc()
	if dir == "" {
		return "", os.ErrNotExist
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	t := now()
	var p string
	for {
		p = filepath.Join(dir, t.Format(historyTimeFormat)+".ghostty")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			break
		}
		t = t.Add(time.Millisecond)
	}
	if err := os.WriteFile(p, []byte(joinLines(lines)), 0o644); err != nil {
		return "", err
	}
	pruneHistory(dir)
	return p, nil
}

func removeSnapshot(p string) {
	if p != "" {
		_ = os.Remove(p)
	}
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// historyIDs returns snapshot IDs, newest first.
func historyIDs(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var ids []string
	for _, e := range entries {
		if id, ok := strings.CutSuffix(e.Name(), ".ghostty"); ok && !e.IsDir() {
			if _, err := time.ParseInLocation(historyTimeFormat, id, time.Local); err == nil {
				ids = append(ids, id)
			}
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids
}

func pruneHistory(dir string) {
	ids := historyIDs(dir)
	for _, id := range ids[min(historyLimit, len(ids)):] {
		_ = os.Remove(filepath.Join(dir, id+".ghostty"))
	}
}

func readConfigValues(path string) (map[string]string, []string, error) {
	c := &ConfigFile{Path: path}
	if err := c.Load(); err != nil {
		return nil, nil, err
	}
	return c.Values(), c.lines, nil
}

func diffValues(from, to map[string]string) []ConfigChange {
	keys := map[string]bool{}
	for k := range from {
		keys[k] = true
	}
	for k := range to {
		keys[k] = true
	}
	var out []ConfigChange
	for _, k := range slices.Sorted(maps.Keys(keys)) {
		if from[k] != to[k] {
			out = append(out, ConfigChange{Key: k, From: from[k], To: to[k]})
		}
	}
	return out
}

// History lists saved states, newest first, with the options each change touched.
func (a *App) History() []HistoryEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	dir := historyDirFunc()
	if dir == "" {
		return nil
	}
	next, _, err := readConfigValues(a.cfg.Path)
	if err != nil {
		return nil
	}
	var out []HistoryEntry
	for _, id := range historyIDs(dir) {
		vals, _, err := readConfigValues(filepath.Join(dir, id+".ghostty"))
		if err != nil {
			continue
		}
		t, _ := time.ParseInLocation(historyTimeFormat, id, time.Local)
		e := HistoryEntry{ID: id, Time: t.Format(time.RFC3339), Keys: []string{}}
		for _, c := range diffValues(vals, next) {
			e.Keys = append(e.Keys, c.Key)
		}
		out = append(out, e)
		next = vals
	}
	return out
}

// RestoreHistory brings back the config as it was in the given entry. The
// current state is saved to the history first, so the restore can be undone.
func (a *App) RestoreHistory(id string) Result {
	dir := historyDirFunc()
	if dir == "" || strings.ContainsAny(id, `/\`) {
		return Result{Message: msg(a.lang, "historyMissing")}
	}
	lines, err := readExistingLines(filepath.Join(dir, id+".ghostty"))
	if err != nil {
		return Result{Message: msg(a.lang, "historyMissing")}
	}
	return a.apply(func() { a.cfg.lines = lines })
}

// readExistingLines is ConfigFile.Load for a file that must exist: a missing
// file must not be mistaken for an empty config.
func readExistingLines(path string) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	_, lines, err := readConfigValues(path)
	return lines, err
}

// ExportConfig saves a copy of the config file where the user chooses.
func (a *App) ExportConfig() Result {
	p, err := saveDialog(a, runtime.SaveDialogOptions{DefaultFilename: "ghostty-config", Title: msg(a.lang, "exportTitle")})
	if err != nil {
		return Result{Message: err.Error()}
	}
	if p == "" {
		return Result{} // cancelled
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.cfg.Load(); err != nil {
		return Result{Message: err.Error()}
	}
	if err := os.WriteFile(p, []byte(joinLines(a.cfg.lines)), 0o644); err != nil {
		return Result{Message: msg(a.lang, "saveFailed") + err.Error()}
	}
	return Result{OK: true, Message: msg(a.lang, "exported") + p}
}

// PreviewImport lets the user choose a config file and shows what would change.
func (a *App) PreviewImport() ImportPreview {
	p, err := openDialog(a, runtime.OpenDialogOptions{Title: msg(a.lang, "importTitle")})
	if err != nil {
		return ImportPreview{Error: err.Error()}
	}
	if p == "" {
		return ImportPreview{} // cancelled
	}
	return a.previewImport(p)
}

func (a *App) previewImport(p string) ImportPreview {
	a.mu.Lock()
	defer a.mu.Unlock()
	pr := ImportPreview{Path: p, Changes: []ConfigChange{}}
	if out, err := runGhosttyCombined(a.bin, "+validate-config", "--config-file="+p); err != nil {
		pr.Error = msg(a.lang, "importInvalid") + strings.TrimSpace(out)
		return pr
	}
	to, _, err := readConfigValues(p)
	if err != nil {
		pr.Error = err.Error()
		return pr
	}
	if err := a.cfg.Load(); err != nil {
		pr.Error = err.Error()
		return pr
	}
	pr.Changes = diffValues(a.cfg.Values(), to)
	return pr
}

// ImportConfig replaces the config with the file chosen in PreviewImport.
func (a *App) ImportConfig(p string) Result {
	lines, err := readExistingLines(p)
	if err != nil {
		return Result{Message: err.Error()}
	}
	return a.apply(func() { a.cfg.lines = lines })
}
