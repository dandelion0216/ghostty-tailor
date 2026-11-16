//go:build i18nsync

// Command i18nsync keeps i18n/docs.ja.json in sync with the option docs of the
// installed Ghostty. It is a developer tool, built only with the i18nsync tag:
//
//	go run -tags i18nsync . status          # count up-to-date, stale, missing and obsolete entries
//	go run -tags i18nsync . export todo.json # write stale and missing entries to translate
//	go run -tags i18nsync . merge todo.json  # merge translated entries back
//	go run -tags i18nsync . prune           # drop entries for options Ghostty no longer has
package main

import (
	"fmt"
	"os"
)

const docsJaPath = "i18n/docs.ja.json"

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	if err := run(os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "i18nsync:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run -tags i18nsync . status | export <file> | merge <file> | prune")
	os.Exit(2)
}

func run(cmd string, args []string) error {
	docs, err := currentDocs()
	if err != nil {
		return err
	}
	ja, err := readTranslations(docsJaPath)
	if err != nil {
		return err
	}
	switch cmd {
	case "status":
		s := syncStatus(docs, ja)
		fmt.Printf("up to date: %d\nstale: %d\nmissing: %d\nobsolete: %d\n", s.UpToDate, len(s.Stale), len(s.Missing), len(s.Obsolete))
		for _, k := range s.Stale {
			fmt.Println("  stale:", k)
		}
		for _, k := range s.Missing {
			fmt.Println("  missing:", k)
		}
		for _, k := range s.Obsolete {
			fmt.Println("  obsolete:", k)
		}
	case "export":
		if len(args) != 1 {
			usage()
		}
		todo := exportTodo(docs, ja)
		if err := writeJSON(args[0], todo); err != nil {
			return err
		}
		fmt.Printf("wrote %d entries to %s\n", len(todo), args[0])
	case "merge":
		if len(args) != 1 {
			usage()
		}
		var todo map[string]todoEntry
		if err := readJSON(args[0], &todo); err != nil {
			return err
		}
		merged, skipped := mergeTodo(docs, ja, todo)
		for _, msg := range skipped {
			fmt.Fprintln(os.Stderr, "skipped:", msg)
		}
		if err := writeJSON(docsJaPath, ja); err != nil {
			return err
		}
		fmt.Printf("merged %d entries into %s\n", merged, docsJaPath)
	case "prune":
		removed := pruneObsolete(docs, ja)
		if err := writeJSON(docsJaPath, ja); err != nil {
			return err
		}
		fmt.Printf("removed %d obsolete entries\n", len(removed))
	default:
		usage()
	}
	return nil
}

// currentDocs returns key -> English doc from the installed Ghostty.
func currentDocs() (map[string]string, error) {
	bin, err := findGhostty()
	if err != nil {
		return nil, err
	}
	out, err := runGhostty(bin, "+show-config", "--default", "--docs")
	if err != nil {
		return nil, err
	}
	docs := map[string]string{}
	for _, o := range parseDocs(out) {
		if o.Doc != "" {
			docs[o.Key] = o.Doc
		}
	}
	return docs, nil
}
