package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const sampleConfig = `# My Ghostty config
theme = Catppuccin Mocha

# fonts
font-family = "JetBrains Mono"
font-size = 14
keybind = super+c=copy_to_clipboard
keybind = super+v=paste_from_clipboard
`

func writeConfig(t *testing.T, body string) *ConfigFile {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.ghostty")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &ConfigFile{Path: p}
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	return c
}

func saveAndRead(t *testing.T, c *ConfigFile) string {
	t.Helper()
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(c.Path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestConfigValues(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	want := map[string]string{
		"theme":       "Catppuccin Mocha",
		"font-family": "JetBrains Mono", // surrounding quotes are removed
		"font-size":   "14",
		"keybind":     "super+c=copy_to_clipboard\nsuper+v=paste_from_clipboard",
	}
	if got := c.Values(); !reflect.DeepEqual(got, want) {
		t.Errorf("Values() = %#v, want %#v", got, want)
	}
}

func TestConfigSetReplacesOnlyTargetLine(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.Set("font-size", "16")
	want := `# My Ghostty config
theme = Catppuccin Mocha

# fonts
font-family = "JetBrains Mono"
font-size = 16
keybind = super+c=copy_to_clipboard
keybind = super+v=paste_from_clipboard
`
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after Set:\n%s\nwant:\n%s", got, want)
	}
}

func TestConfigSetAppendsNewKey(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.Set("cursor-style", "bar")
	want := sampleConfig + "cursor-style = bar\n"
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after Set:\n%s\nwant:\n%s", got, want)
	}
}

func TestConfigSetCollapsesRepeatedKeyAtFirstPosition(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.Set("keybind", "super+x=cut")
	want := `# My Ghostty config
theme = Catppuccin Mocha

# fonts
font-family = "JetBrains Mono"
font-size = 14
keybind = super+x=cut
`
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after Set:\n%s\nwant:\n%s", got, want)
	}
}

func TestConfigSetIgnoresCommentedOutKey(t *testing.T) {
	c := writeConfig(t, "# font-size = 20\nfont-size = 12\n")
	c.Set("font-size", "18")
	want := "# font-size = 20\nfont-size = 18\n"
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after Set = %q, want %q", got, want)
	}
}

func TestConfigUnsetRemovesEveryLineForKey(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.Unset("keybind")
	c.Unset("theme")
	want := `# My Ghostty config

# fonts
font-family = "JetBrains Mono"
font-size = 14
`
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after Unset:\n%s\nwant:\n%s", got, want)
	}
}

func TestConfigMissingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "config.ghostty")
	c := &ConfigFile{Path: p}
	if err := c.Load(); err != nil {
		t.Fatalf("Load() on a missing file: %v", err)
	}
	if v := c.Values(); len(v) != 0 {
		t.Errorf("Values() = %v, want empty", v)
	}
	c.Set("font-size", "15")
	if got := saveAndRead(t, c); got != "font-size = 15\n" {
		t.Errorf("new file = %q", got)
	}
}

func TestConfigEmptyFile(t *testing.T) {
	c := writeConfig(t, "")
	c.Set("font-size", "15")
	if got := saveAndRead(t, c); got != "font-size = 15\n" {
		t.Errorf("file after Set on an empty file = %q, want %q", got, "font-size = 15\n")
	}
}

func TestConfigSaveLeavesNoTempFile(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.Set("font-size", "16")
	saveAndRead(t, c)
	if _, err := os.Stat(c.Path + ".tailor-tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file left behind: %v", err)
	}
}

func TestConfigBackupOnlyOnce(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	dst, err := c.Backup()
	if err != nil {
		t.Fatal(err)
	}
	if dst != c.Path+".bak-tailor" {
		t.Errorf("backup path = %q", dst)
	}
	c.Set("font-size", "30")
	saveAndRead(t, c)
	if _, err := c.Backup(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != sampleConfig {
		t.Errorf("backup was overwritten:\n%s", b)
	}
}

func TestConfigBackupWithoutConfig(t *testing.T) {
	c := &ConfigFile{Path: filepath.Join(t.TempDir(), "config.ghostty")}
	dst, err := c.Backup()
	if err != nil || dst != "" {
		t.Errorf("Backup() = %q, %v; want no backup and no error", dst, err)
	}
}

func TestConfigSetAllReplacesRepeatedKeyInPlace(t *testing.T) {
	c := writeConfig(t, sampleConfig)
	c.SetAll("keybind", []string{"super+x=cut", "super+y=redo", "super+z=undo"})
	want := `# My Ghostty config
theme = Catppuccin Mocha

# fonts
font-family = "JetBrains Mono"
font-size = 14
keybind = super+x=cut
keybind = super+y=redo
keybind = super+z=undo
`
	if got := saveAndRead(t, c); got != want {
		t.Errorf("file after SetAll:\n%s\nwant:\n%s", got, want)
	}
}

func TestConfigSetAllAppendsAndRemoves(t *testing.T) {
	c := writeConfig(t, "theme = Nord\n")
	c.SetAll("palette", []string{"1=#ff0000", "9=#ff6b8b"})
	if got := saveAndRead(t, c); got != "theme = Nord\npalette = 1=#ff0000\npalette = 9=#ff6b8b\n" {
		t.Errorf("file after SetAll = %q", got)
	}
	c.SetAll("palette", nil)
	if got := saveAndRead(t, c); got != "theme = Nord\n" {
		t.Errorf("file after SetAll(nil) = %q", got)
	}
}
