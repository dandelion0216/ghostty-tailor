package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveInclude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	file := "/cfg/ghostty/config.ghostty"
	tests := []struct{ raw, want string }{
		{"local.ghostty", "/cfg/ghostty/local.ghostty"},
		{"?local.ghostty", "/cfg/ghostty/local.ghostty"}, // optional
		{`"?odd"`, "/cfg/ghostty/?odd"},                  // quoted: the ? is part of the name
		{`"with space"`, "/cfg/ghostty/with space"},
		{"../shared/a", "/cfg/shared/a"},
		{"/etc/ghostty/b", "/etc/ghostty/b"},
		{"~/dotfiles/c", filepath.Join(home, "dotfiles/c")},
		{"?", ""},
	}
	for _, tt := range tests {
		if got := resolveInclude(file, tt.raw); got != tt.want {
			t.Errorf("resolveInclude(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestIncludedValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "ghostty")
	write := func(name, body string) {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.ghostty", "font-size = 16\nfont-family = Noto Sans CJK JP\nconfig-file = nested/c.ghostty\n")
	write("b.ghostty", "font-size = 18\n# comment\nconfig-file = ../ghostty/config.ghostty\n") // cycle back to main: skipped
	write("nested/c.ghostty", "theme = Nord\nconfig-file = ../a.ghostty\n")                    // cycle: skipped

	main := filepath.Join(dir, "config.ghostty")
	lines := strings.Split("theme = Dracula\nconfig-file = ?a.ghostty\nconfig-file = ?missing\nconfig-file = b.ghostty\nfont-size = 12", "\n")
	got := includedValues(main, lines)
	want := map[string][]IncludedValue{
		// a and b load before c, which a includes, as Ghostty does; the last value (b) wins.
		"font-size":   {{"~/.config/ghostty/a.ghostty", "16"}, {"~/.config/ghostty/b.ghostty", "18"}},
		"font-family": {{"~/.config/ghostty/a.ghostty", "Noto Sans CJK JP"}},
		"theme":       {{"~/.config/ghostty/nested/c.ghostty", "Nord"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("includedValues() =\n%v\nwant\n%v", got, want)
	}

	if got := includedValues(main, []string{"theme = Dracula"}); len(got) != 0 {
		t.Errorf("without config-file: %v, want none", got)
	}
}
