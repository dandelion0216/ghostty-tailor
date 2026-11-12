package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestParseDocs(t *testing.T) {
	opts := parseDocs(readFixture(t, "show-config-docs.txt"))
	byKey := map[string]Option{}
	var keys []string
	for _, o := range opts {
		byKey[o.Key] = o
		keys = append(keys, o.Key)
	}

	wantKeys := []string{"language", "font-family", "font-size", "cursor-style", "background", "cursor-style-blink",
		"window-padding-x", "macos-titlebar-style", "keybind", "theme", "undo-timeout", "zzz-future-option"}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Fatalf("keys = %v, want %v (in file order, repeated keys once)", keys, wantKeys)
	}

	tests := []struct {
		key, kind, def, category, reload string
		choices                          []string
	}{
		{"language", "text", "", "linux", "restart", nil},
		{"font-family", "text", "", "font", "live", nil},
		{"font-size", "number", "13", "font", "live", nil},
		{"cursor-style", "enum", "block", "cursor", "live", []string{"block", "bar", "underline", "block_hollow"}},
		{"background", "color", "#282c34", "colors", "live", nil},
		{"cursor-style-blink", "bool", "true", "cursor", "live", nil},
		{"window-padding-x", "number", "2", "window", "newwindow", nil},
		{"macos-titlebar-style", "enum", "transparent", "macos", "live", []string{"native", "transparent", "tabs", "hidden"}},
		{"keybind", "repeat", "super+c=copy_to_clipboard\nsuper+v=paste_from_clipboard", "keys", "live", nil},
		{"theme", "text", "", "colors", "live", nil},
		{"undo-timeout", "text", "5s", "window", "live", nil},
		{"zzz-future-option", "text", "auto", "other", "live", nil},
	}
	for _, tt := range tests {
		o := byKey[tt.key]
		if o.Kind != tt.kind || o.Default != tt.def || o.Category != tt.category || o.Reload != tt.reload || !reflect.DeepEqual(o.Choices, tt.choices) {
			t.Errorf("%s = {kind %q, default %q, category %q, reload %q, choices %v}, want {%q, %q, %q, %q, %v}",
				tt.key, o.Kind, o.Default, o.Category, o.Reload, o.Choices, tt.kind, tt.def, tt.category, tt.reload, tt.choices)
		}
	}

	lang := byKey["language"]
	if want := "Set Ghostty's graphical user interface language to a language other than the system default language. For example:"; lang.Summary != want {
		t.Errorf("language summary = %q, want %q", lang.Summary, want)
	}
	if want := "Set Ghostty's graphical user interface language to a language other than the\nsystem default language. For example:\n\n    language = de\n\nWarning: This setting cannot be reloaded at runtime. To change the language\nyou must fully restart Ghostty."; lang.Doc != want {
		t.Errorf("language doc = %q, want %q", lang.Doc, want)
	}
}

func TestDetectKindNeedsDefaultInChoices(t *testing.T) {
	doc := "Values:\n  * `one`\n  * `two`"
	if kind, _ := detectKind("x", "three", doc); kind == "enum" {
		t.Errorf("default not in the bullet list should not be an enum")
	}
	if kind, choices := detectKind("x", "two", doc); kind != "enum" || !reflect.DeepEqual(choices, []string{"one", "two"}) {
		t.Errorf("detectKind = %q %v", kind, choices)
	}
}

func TestReloadKind(t *testing.T) {
	tests := map[string]string{
		"This cannot be reloaded at runtime.":         "restart",
		"Requires a full restart of Ghostty.":         "restart",
		"Only affects new windows.":                   "newwindow",
		"Applies to newly created terminals.":         "newwindow",
		"Changes apply to all terminals immediately.": "live",
	}
	for doc, want := range tests {
		if got := reloadKind(doc); got != want {
			t.Errorf("reloadKind(%q) = %q, want %q", doc, got, want)
		}
	}
}

func TestParseFonts(t *testing.T) {
	got := parseFonts(readFixture(t, "list-fonts.txt"))
	want := []string{"Andale Mono", "BIZ UDGothic", "JetBrains Mono"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseFonts = %v, want %v", got, want)
	}
}

func TestLoadThemes(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)

	bin := filepath.Join(root, "prefix", "bin", "ghostty")
	bundled := filepath.Join(root, "prefix", "share", "ghostty", "themes")
	user := filepath.Join(home, ".config", "ghostty", "themes")
	for _, d := range []string{filepath.Dir(bin), bundled, user} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, body string) {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(bin, "")
	write(filepath.Join(bundled, "night"), "background = 101010\nforeground = #eeeeee\npalette = 1=#ff0000\npalette = 2=00ff00\n")
	write(filepath.Join(bundled, "Day"), "background = #fafafa\nforeground = #111111\n")
	write(filepath.Join(bundled, "Mine"), "background = #000000\n")
	write(filepath.Join(user, "Mine"), "background = #ffffff\n") // the user's copy wins

	themes := loadThemes(bin)
	var names []string
	byName := map[string]Theme{}
	for _, th := range themes {
		names = append(names, th.Name)
		byName[th.Name] = th
	}
	if want := []string{"Day", "Mine", "night"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("theme names = %v, want %v (case-insensitive order, no duplicates)", names, want)
	}
	night := byName["night"]
	if night.Background != "#101010" || night.Palette[1] != "#ff0000" || night.Palette[2] != "#00ff00" || !night.Dark {
		t.Errorf("night = %+v", night)
	}
	if byName["Day"].Dark {
		t.Errorf("Day should be light")
	}
	if byName["Mine"].Background != "#ffffff" {
		t.Errorf("user theme should override the bundled one, got %+v", byName["Mine"])
	}
}

func TestLuminance(t *testing.T) {
	if l := luminance("#000000"); l != 0 {
		t.Errorf("black = %v", l)
	}
	if l := luminance("#FFFFFF"); l < 0.999 {
		t.Errorf("white = %v", l)
	}
	if l := luminance("not-a-color"); l != 0 {
		t.Errorf("invalid = %v", l)
	}
}
