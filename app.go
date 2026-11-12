package main

import (
	"context"
	"strings"
	"sync"
)

type App struct {
	ctx    context.Context
	mu     sync.Mutex
	bin    string
	cfg    *ConfigFile
	backup string
}

type State struct {
	Version    string            `json:"version"`
	ConfigPath string            `json:"configPath"`
	BackupPath string            `json:"backupPath"`
	Options    []Option          `json:"options"`
	Set        map[string]string `json:"set"`
	Error      string            `json:"error"`
}

type Result struct {
	OK       bool              `json:"ok"`
	Message  string            `json:"message"`
	Reloaded bool              `json:"reloaded"`
	Set      map[string]string `json:"set"`
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = &ConfigFile{Path: defaultConfigPath()}
}

func (a *App) Load() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	st := State{ConfigPath: a.cfg.Path, Set: map[string]string{}}
	bin, err := findGhostty()
	if err != nil {
		st.Error = err.Error()
		return st
	}
	a.bin = bin
	if v, err := runGhostty(bin, "+version"); err == nil {
		st.Version = strings.TrimSpace(strings.SplitN(v, "\n", 2)[0])
	}
	docs, err := runGhostty(bin, "+show-config", "--default", "--docs")
	if err != nil {
		st.Error = "設定項目の読み込みに失敗しました: " + err.Error()
		return st
	}
	st.Options = parseDocs(docs)
	if err := a.cfg.Load(); err != nil {
		st.Error = "設定ファイルを読めません: " + err.Error()
		return st
	}
	st.Set = a.cfg.Values()
	if b, err := a.cfg.Backup(); err == nil {
		a.backup = b
	}
	st.BackupPath = a.backup
	return st
}

func (a *App) Themes() []Theme {
	return loadThemes(a.bin)
}

func (a *App) Fonts() []string {
	out, err := runGhostty(a.bin, "+list-fonts")
	if err != nil {
		return nil
	}
	return parseFonts(out)
}

func (a *App) SetOption(key, value string) Result {
	return a.apply(func() {
		if strings.TrimSpace(value) == "" {
			a.cfg.Unset(key)
		} else {
			a.cfg.Set(key, value)
		}
	})
}

func (a *App) ResetOption(key string) Result {
	return a.apply(func() { a.cfg.Unset(key) })
}

func (a *App) apply(edit func()) Result {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.cfg.Load(); err != nil {
		return Result{Message: err.Error()}
	}
	before := append([]string(nil), a.cfg.lines...)
	edit()
	if err := a.cfg.Save(); err != nil {
		return Result{Message: "保存に失敗しました: " + err.Error()}
	}
	if out, err := runGhosttyCombined(a.bin, "+validate-config", "--config-file="+a.cfg.Path); err != nil {
		a.cfg.lines = before
		_ = a.cfg.Save()
		return Result{Message: "Ghostty がこの値を受け付けなかったため元に戻しました: " + strings.TrimSpace(out), Set: a.cfg.Values()}
	}
	reloaded, msg := reloadGhostty()
	if reloaded {
		msg = "保存して Ghostty に反映しました"
	}
	return Result{OK: true, Message: msg, Reloaded: reloaded, Set: a.cfg.Values()}
}
