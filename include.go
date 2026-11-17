package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Files pulled in with `config-file` are read by Ghostty after the file that
// names them, so a value set there wins over the main file. ghostty-tailor
// only edits the main file; it reads the included ones to tell the user which
// options they set.

// IncludedValue is one line of an included file.
type IncludedValue struct {
	File  string `json:"file"` // shown to the user, with ~ for the home directory
	Value string `json:"value"`
}

// includedValues reads the files included from the main config (given as its
// lines), and the files those include, in the order Ghostty loads them. It
// returns the values they set, by key. Missing files and cycles are skipped.
func includedValues(mainPath string, lines []string) map[string][]IncludedValue {
	out := map[string][]IncludedValue{}
	seen := map[string]bool{filepath.Clean(mainPath): true}
	queue := includes(mainPath, lines)
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		if seen[path] {
			continue
		}
		seen[path] = true
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fileLines := strings.Split(string(b), "\n")
		for _, l := range fileLines {
			if k, v, ok := parseLine(l); ok && k != "config-file" {
				out[k] = append(out[k], IncludedValue{File: displayPath(path), Value: v})
			}
		}
		// Files included from this one load after the ones already queued.
		queue = append(queue, includes(path, fileLines)...)
	}
	return out
}

// includes returns the paths named by `config-file` lines in a file.
func includes(file string, lines []string) []string {
	var paths []string
	for _, l := range lines {
		if k, v, ok := parseRawLine(l); ok && k == "config-file" {
			if p := resolveInclude(file, v); p != "" {
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// resolveInclude turns a `config-file` value into a path. A leading ? only
// marks the file as optional, unless the value is quoted. Relative paths are
// relative to the file containing the line.
func resolveInclude(file, raw string) string {
	if isQuoted(raw) {
		raw = raw[1 : len(raw)-1]
	} else {
		raw = strings.TrimPrefix(raw, "?")
	}
	if raw == "" {
		return ""
	}
	if rest, ok := strings.CutPrefix(raw, "~/"); ok {
		home, _ := os.UserHomeDir()
		raw = filepath.Join(home, rest)
	}
	if !filepath.IsAbs(raw) {
		raw = filepath.Join(filepath.Dir(file), raw)
	}
	return filepath.Clean(raw)
}

func displayPath(p string) string {
	home, _ := os.UserHomeDir()
	if rest, ok := strings.CutPrefix(p, home+string(filepath.Separator)); ok && home != "" {
		return "~/" + rest
	}
	return p
}
