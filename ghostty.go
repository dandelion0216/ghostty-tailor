package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// Option is one Ghostty config key, described from `ghostty +show-config --default --docs`.
type Option struct {
	Key      string   `json:"key"`
	Summary  string   `json:"summary"`
	Doc      string   `json:"doc"`
	Default  string   `json:"default"`
	Kind     string   `json:"kind"` // bool, enum, color, number, text, repeat
	Choices  []string `json:"choices"`
	Category string   `json:"category"`
	Reload   string   `json:"reload"` // live, newwindow, restart
}

type Theme struct {
	Name       string   `json:"name"`
	Background string   `json:"background"`
	Foreground string   `json:"foreground"`
	Palette    []string `json:"palette"`
	Dark       bool     `json:"dark"`
}

func findGhostty() (string, error) {
	if p, err := exec.LookPath("ghostty"); err == nil {
		return p, nil
	}
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		"/Applications/Ghostty.app/Contents/MacOS/ghostty",
		filepath.Join(home, "Applications/Ghostty.app/Contents/MacOS/ghostty"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New("ghostty コマンドが見つかりません")
}

func runGhostty(bin string, args ...string) (string, error) {
	out, err := exec.Command(bin, args...).Output()
	return string(out), err
}

func runGhosttyCombined(bin string, args ...string) (string, error) {
	out, err := exec.Command(bin, args...).CombinedOutput()
	return string(out), err
}

var (
	enumRe   = regexp.MustCompile("(?m)^\\s*\\*\\s+`([^`]+)`")
	hexRe    = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)
	numberRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
)

func parseDocs(out string) []Option {
	var opts []Option
	idx := map[string]int{}
	var doc []string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "#"):
			doc = append(doc, strings.TrimPrefix(strings.TrimPrefix(line, "#"), " "))
		case strings.TrimSpace(line) == "":
			doc = nil
		default:
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			k, v = strings.TrimSpace(k), strings.TrimSpace(v)
			if i, seen := idx[k]; seen {
				opts[i].Kind = "repeat"
				opts[i].Choices = nil
				opts[i].Default += "\n" + v
				continue
			}
			full := strings.TrimSpace(strings.Join(doc, "\n"))
			kind, choices := detectKind(k, v, full)
			idx[k] = len(opts)
			opts = append(opts, Option{
				Key:      k,
				Summary:  firstParagraph(full),
				Doc:      full,
				Default:  v,
				Kind:     kind,
				Choices:  choices,
				Category: categorize(k),
				Reload:   reloadKind(full),
			})
			doc = nil
		}
	}
	return opts
}

func firstParagraph(doc string) string {
	p, _, _ := strings.Cut(doc, "\n\n")
	return strings.Join(strings.Fields(p), " ")
}

// Options whose valid values are written in prose rather than as a bullet list.
var knownChoices = map[string][]string{
	"macos-titlebar-style": {"native", "transparent", "tabs", "hidden"},
}

func detectKind(key, def, doc string) (string, []string) {
	if c, ok := knownChoices[key]; ok {
		return "enum", c
	}
	if def == "true" || def == "false" {
		return "bool", nil
	}
	var choices []string
	seen := map[string]bool{}
	for _, m := range enumRe.FindAllStringSubmatch(doc, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			choices = append(choices, m[1])
		}
	}
	if len(choices) >= 2 && seen[def] {
		return "enum", choices
	}
	if hexRe.MatchString(def) || strings.HasSuffix(key, "-color") || key == "background" || key == "foreground" ||
		strings.HasSuffix(key, "-background") || strings.HasSuffix(key, "-foreground") || key == "cursor-text" {
		return "color", nil
	}
	if numberRe.MatchString(def) {
		return "number", nil
	}
	return "text", nil
}

func reloadKind(doc string) string {
	d := strings.ToLower(doc)
	switch {
	case strings.Contains(d, "cannot be reloaded"), strings.Contains(d, "full restart"), strings.Contains(d, "restart ghostty"):
		return "restart"
	case strings.Contains(d, "new windows"), strings.Contains(d, "new terminals"), strings.Contains(d, "newly created"):
		return "newwindow"
	}
	return "live"
}

var categories = []struct {
	name     string
	prefixes []string
}{
	{"色・テーマ", []string{"search-", "theme", "background", "foreground", "palette", "selection-", "minimum-contrast", "bold-color", "bold-is-bright", "faint-opacity", "unfocused-split", "split-divider-color", "cursor-color", "cursor-text", "cursor-opacity", "alpha-blending"}},
	{"フォント", []string{"font-", "adjust-", "grapheme-", "freetype-"}},
	{"カーソル", []string{"cursor-"}},
	{"ウィンドウ", []string{"window-", "resize-overlay", "focus-follows", "split-", "maximize", "fullscreen", "title", "undo-timeout", "background-blur"}},
	{"マウス・クリップボード", []string{"mouse-", "click-", "clipboard-", "copy-on-select", "link", "right-click", "scroll"}},
	{"キー操作", []string{"keybind", "key-", "quick-terminal", "command-palette"}},
	{"シェル・起動", []string{"command", "initial-command", "shell-integration", "working-directory", "env", "input", "wait-after-command", "abnormal-command", "term", "scrollback", "osc-", "image-storage", "confirm-close", "quit-after", "notify", "desktop-notifications", "bell-", "app-notifications", "progress", "config-"}},
	{"macOS", []string{"macos-"}},
	{"Linux (GTK)", []string{"gtk-", "linux-", "adw-", "x11-", "async-backend", "class", "language"}},
}

func categorize(key string) string {
	for _, c := range categories {
		for _, p := range c.prefixes {
			if strings.HasPrefix(key, p) {
				return c.name
			}
		}
	}
	return "その他"
}

func themeDirs(bin string) []string {
	home, _ := os.UserHomeDir()
	dirs := []string{filepath.Join(home, ".config/ghostty/themes")}
	if runtime.GOOS == "darwin" {
		dirs = append(dirs, filepath.Join(home, "Library/Application Support/com.mitchellh.ghostty/themes"))
	}
	// <prefix>/bin/ghostty -> <prefix>/share/ghostty/themes, and the macOS app bundle layout.
	real, err := filepath.EvalSymlinks(bin)
	if err == nil {
		dirs = append(dirs,
			filepath.Join(filepath.Dir(real), "../share/ghostty/themes"),
			filepath.Join(filepath.Dir(real), "../Resources/ghostty/themes"))
	}
	return append(dirs, "/usr/share/ghostty/themes", "/usr/local/share/ghostty/themes")
}

func loadThemes(bin string) []Theme {
	seen := map[string]bool{}
	var themes []Theme
	for _, dir := range themeDirs(bin) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || seen[e.Name()] {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			seen[e.Name()] = true
			t := Theme{Name: e.Name(), Palette: make([]string, 16)}
			for _, line := range strings.Split(string(b), "\n") {
				k, v, ok := strings.Cut(line, "=")
				if !ok {
					continue
				}
				k, v = strings.TrimSpace(k), strings.TrimSpace(v)
				switch k {
				case "background":
					t.Background = normHex(v)
				case "foreground":
					t.Foreground = normHex(v)
				case "palette":
					n, c, ok := strings.Cut(v, "=")
					if !ok {
						continue
					}
					var i int
					for _, ch := range n {
						i = i*10 + int(ch-'0')
					}
					if i >= 0 && i < 16 {
						t.Palette[i] = normHex(c)
					}
				}
			}
			t.Dark = luminance(t.Background) < 0.5
			themes = append(themes, t)
		}
	}
	sort.Slice(themes, func(i, j int) bool { return strings.ToLower(themes[i].Name) < strings.ToLower(themes[j].Name) })
	return themes
}

func normHex(v string) string {
	if hexRe.MatchString(v) && !strings.HasPrefix(v, "#") {
		return "#" + v
	}
	return v
}

func luminance(hex string) float64 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}
	var rgb [3]float64
	for i := 0; i < 3; i++ {
		var n int
		for _, ch := range hex[i*2 : i*2+2] {
			n *= 16
			switch {
			case ch >= '0' && ch <= '9':
				n += int(ch - '0')
			case ch >= 'a' && ch <= 'f':
				n += int(ch-'a') + 10
			case ch >= 'A' && ch <= 'F':
				n += int(ch-'A') + 10
			}
		}
		rgb[i] = float64(n) / 255
	}
	return 0.299*rgb[0] + 0.587*rgb[1] + 0.114*rgb[2]
}

func parseFonts(out string) []string {
	var fonts []string
	for _, line := range strings.Split(out, "\n") {
		if line != "" && !strings.HasPrefix(line, " ") {
			fonts = append(fonts, strings.TrimSpace(line))
		}
	}
	return fonts
}

// reloadGhostty asks the running Ghostty to reload its config.
func reloadGhostty() (bool, string) {
	if runtime.GOOS != "darwin" {
		return false, "Linux での即時反映は試作品では未対応です（Ghostty で Ctrl+Shift+, を押してください）"
	}
	script := `if application "Ghostty" is running then
	tell application "Ghostty"
		if (count of terminals) > 0 then
			perform action "reload_config" on terminal 1
			return "ok"
		end if
	end tell
	return "noterminal"
end if
return "notrunning"`
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	res := strings.TrimSpace(string(out))
	switch {
	case err != nil:
		return false, "Ghostty への再読み込み指示に失敗しました: " + res
	case res == "notrunning":
		return false, "Ghostty が起動していないため、保存のみ行いました"
	case res == "noterminal":
		return false, "Ghostty のウィンドウがないため、保存のみ行いました"
	}
	return true, ""
}
