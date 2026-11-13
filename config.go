package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConfigFile edits a Ghostty config file line by line, leaving comments and
// unrelated lines untouched.
type ConfigFile struct {
	Path  string
	lines []string
}

func defaultConfigPath() string {
	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}
	var cands []string
	// On macOS the Application Support file is loaded after the XDG one, so it wins.
	if runtime.GOOS == "darwin" {
		dir := filepath.Join(home, "Library/Application Support/com.mitchellh.ghostty")
		cands = append(cands, filepath.Join(dir, "config.ghostty"), filepath.Join(dir, "config"))
	}
	cands = append(cands, filepath.Join(xdg, "ghostty/config.ghostty"), filepath.Join(xdg, "ghostty/config"))
	for _, c := range cands {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return cands[0]
}

func (c *ConfigFile) Load() error {
	b, err := os.ReadFile(c.Path)
	if os.IsNotExist(err) {
		c.lines = nil
		return nil
	}
	if err != nil {
		return err
	}
	body := strings.TrimRight(string(b), "\n")
	if body == "" {
		c.lines = nil
		return nil
	}
	c.lines = strings.Split(body, "\n")
	return nil
}

func parseLine(line string) (string, string, bool) {
	t := strings.TrimSpace(line)
	if t == "" || strings.HasPrefix(t, "#") {
		return "", "", false
	}
	k, v, ok := strings.Cut(t, "=")
	if !ok {
		return "", "", false
	}
	v = strings.TrimSpace(v)
	if len(v) >= 2 && strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		v = v[1 : len(v)-1]
	}
	return strings.TrimSpace(k), v, true
}

// Values returns the keys set in the file. Repeated keys are joined with newlines.
func (c *ConfigFile) Values() map[string]string {
	vals := map[string]string{}
	for _, l := range c.lines {
		if k, v, ok := parseLine(l); ok {
			if prev, seen := vals[k]; seen {
				vals[k] = prev + "\n" + v
			} else {
				vals[k] = v
			}
		}
	}
	return vals
}

// Set replaces every line for key with a single `key = value` line.
func (c *ConfigFile) Set(key, value string) {
	newLine := key + " = " + value
	var out []string
	placed := false
	for _, l := range c.lines {
		if k, _, ok := parseLine(l); ok && k == key {
			if !placed {
				out = append(out, newLine)
				placed = true
			}
			continue
		}
		out = append(out, l)
	}
	if !placed {
		out = append(out, newLine)
	}
	c.lines = out
}

// SetAll replaces every line for a repeatable key (palette, keybind, ...) with
// one line per value, placed where the first existing line was (or at the end).
// An empty list removes the key.
func (c *ConfigFile) SetAll(key string, values []string) {
	var out []string
	placed := false
	put := func() {
		for _, v := range values {
			out = append(out, key+" = "+v)
		}
		placed = true
	}
	for _, l := range c.lines {
		if k, _, ok := parseLine(l); ok && k == key {
			if !placed {
				put()
			}
			continue
		}
		out = append(out, l)
	}
	if !placed {
		put()
	}
	c.lines = out
}

func (c *ConfigFile) Unset(key string) {
	var out []string
	for _, l := range c.lines {
		if k, _, ok := parseLine(l); ok && k == key {
			continue
		}
		out = append(out, l)
	}
	c.lines = out
}

func (c *ConfigFile) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o755); err != nil {
		return err
	}
	tmp := c.Path + ".tailor-tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(c.lines, "\n")+"\n"), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.Path)
}

// Backup copies the config once, the first time ghostty-tailor touches it.
func (c *ConfigFile) Backup() (string, error) {
	dst := c.Path + ".bak-tailor"
	if _, err := os.Stat(dst); err == nil {
		return dst, nil
	}
	b, err := os.ReadFile(c.Path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return dst, os.WriteFile(dst, b, 0o644)
}
