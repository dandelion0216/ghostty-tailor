package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Settings are ghostty-tailor's own preferences, kept apart from the Ghostty config.
type Settings struct {
	Lang string `json:"lang"`
}

func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "ghostty-tailor", "settings.json")
}

func loadSettings() Settings {
	var s Settings
	if b, err := os.ReadFile(settingsPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func saveSettings(s Settings) error {
	p := settingsPath()
	if p == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(p, append(b, '\n'), 0o644)
}

func normalizeLang(v string) string {
	if strings.HasPrefix(strings.ToLower(v), "ja") {
		return "ja"
	}
	return "en"
}

var appleLangRe = regexp.MustCompile(`[A-Za-z][A-Za-z_-]*`)

// detectLang returns the OS UI language. Apps started from Finder have no LANG,
// so macOS reads the preferred languages list instead.
func detectLang() string {
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("defaults", "read", "-g", "AppleLanguages").Output(); err == nil {
			if m := appleLangRe.FindString(string(out)); m != "" {
				return normalizeLang(m)
			}
		}
	}
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return normalizeLang(v)
		}
	}
	return "en"
}

var messages = map[string]map[string]string{
	"en": {
		"notFound":       "The ghostty command was not found",
		"loadOptions":    "Failed to read the config options: ",
		"readConfig":     "Cannot read the config file: ",
		"saveFailed":     "Failed to save: ",
		"rejected":       "Ghostty rejected this value, so the change was reverted: ",
		"reloaded":       "Saved and applied to Ghostty",
		"linuxNoReload":  "Live reload is not supported on Linux yet. Saved only — press Ctrl+Shift+, in Ghostty to reload",
		"reloadFailed":   "Failed to ask Ghostty to reload: ",
		"notRunning":     "Ghostty is not running. Saved only",
		"noTerminal":     "Ghostty has no open window. Saved only",
		"settingsFailed": "Failed to save the language setting: ",
	},
	"ja": {
		"notFound":       "ghostty コマンドが見つかりません",
		"loadOptions":    "設定項目の読み込みに失敗しました: ",
		"readConfig":     "設定ファイルを読めません: ",
		"saveFailed":     "保存に失敗しました: ",
		"rejected":       "Ghostty がこの値を受け付けなかったため元に戻しました: ",
		"reloaded":       "保存して Ghostty に反映しました",
		"linuxNoReload":  "Linux での即時反映は未対応です。保存のみ行いました（Ghostty で Ctrl+Shift+, を押すと反映されます）",
		"reloadFailed":   "Ghostty への再読み込み指示に失敗しました: ",
		"notRunning":     "Ghostty が起動していないため、保存のみ行いました",
		"noTerminal":     "Ghostty のウィンドウがないため、保存のみ行いました",
		"settingsFailed": "言語設定の保存に失敗しました: ",
	},
}

func msg(lang, id string) string {
	return messages[normalizeLang(lang)][id]
}

// Japanese translations of the option docs. Each entry records a hash of the
// English text it was translated from, so a doc that changed in a newer Ghostty
// falls back to English instead of showing an outdated translation.
//
//go:embed i18n/docs.ja.json
var docsJaJSON []byte

type docTranslation struct {
	Src string `json:"src"`
	Doc string `json:"doc"`
}

func docHash(doc string) string {
	sum := sha256.Sum256([]byte(doc))
	return hex.EncodeToString(sum[:8])
}

func applyDocTranslations(opts []Option) {
	var ja map[string]docTranslation
	if err := json.Unmarshal(docsJaJSON, &ja); err != nil {
		return
	}
	for i, o := range opts {
		if t, ok := ja[o.Key]; ok && t.Src == docHash(o.Doc) {
			opts[i].DocJa = t.Doc
			opts[i].SummaryJa = firstParagraph(t.Doc)
		}
	}
}
