package main

import (
	"sort"
	"strconv"
	"strings"
)

// KeyRow is one shortcut as shown in the Keys tab: a default binding (possibly
// disabled or moved to another key) or one the user added.
type KeyRow struct {
	ID             string `json:"id"`
	Action         string `json:"action"`
	Trigger        string `json:"trigger"`
	DefaultTrigger string `json:"defaultTrigger"` // empty for added rows
	Enabled        bool   `json:"enabled"`
	Origin         string `json:"origin"` // default, changed, added
}

type KeybindState struct {
	Rows    []KeyRow `json:"rows"`
	Actions []string `json:"actions"`
	// Extra keeps keybind lines the Keys tab does not model (e.g. `clear`),
	// so saving never drops them.
	Extra []string `json:"extra"`
}

// splitBinding splits "trigger=action". The key itself may be "=" or "+",
// as in "super+==increase_font_size:1" or "super++=increase_font_size:1".
func splitBinding(s string) (string, string, bool) {
	for i := 1; i < len(s); i++ {
		if s[i] != '=' {
			continue
		}
		if s[i-1] == '+' && !(i >= 2 && s[i-2] == '+') {
			continue // "+=" here means the key is "="
		}
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), true
	}
	return "", "", false
}

var modAlias = map[string]string{
	"cmd": "super", "command": "super", "super": "super",
	"opt": "alt", "option": "alt", "alt": "alt",
	"control": "ctrl", "ctrl": "ctrl",
	"shift": "shift",
}

var modOrder = map[string]int{"ctrl": 0, "alt": 1, "shift": 2, "super": 3}

// normTrigger makes triggers comparable: modifier aliases and order are
// unified and keys are lower-cased. Prefixes such as "global:" and key
// sequences ("ctrl+a>n") are kept.
func normTrigger(t string) string {
	t = strings.TrimSpace(t)
	var prefix string
	for {
		i := strings.Index(t, ":")
		if i <= 0 || strings.ContainsAny(t[:i], "+>") {
			break
		}
		prefix += strings.ToLower(t[:i+1])
		t = t[i+1:]
	}
	chords := strings.Split(t, ">")
	for ci, chord := range chords {
		var mods []string
		key := ""
		parts := strings.Split(chord, "+")
		// "super++" splits into ["super", "", ""]: the key is "+".
		if strings.HasSuffix(chord, "++") {
			parts = append(strings.Split(strings.TrimSuffix(chord, "++"), "+"), "+")
		}
		for i, p := range parts {
			lp := strings.ToLower(strings.TrimSpace(p))
			if m, ok := modAlias[lp]; ok && i < len(parts)-1 {
				mods = append(mods, m)
				continue
			}
			key = lp
			if p == "+" {
				key = "+"
			}
		}
		sort.Slice(mods, func(a, b int) bool { return modOrder[mods[a]] < modOrder[mods[b]] })
		chords[ci] = strings.Join(append(mods, key), "+")
	}
	return prefix + strings.Join(chords, ">")
}

func parseKeybindLines(text string) [][2]string {
	var out [][2]string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "keybind ="))
		if t, a, ok := splitBinding(line); ok {
			out = append(out, [2]string{t, a})
		}
	}
	return out
}

// buildKeyRows applies the user's keybind lines on top of Ghostty's defaults.
func buildKeyRows(defaults [][2]string, user []string) ([]KeyRow, []string) {
	rows := make([]KeyRow, 0, len(defaults)+len(user))
	for i, d := range defaults {
		rows = append(rows, KeyRow{ID: "d" + strconv.Itoa(i), Action: d[1], Trigger: d[0], DefaultTrigger: d[0], Enabled: true, Origin: "default"})
	}
	var extra []string
	unbound := map[int]bool{} // default rows explicitly unbound by the user
	added := 0
	for _, line := range user {
		t, a, ok := splitBinding(line)
		if !ok {
			extra = append(extra, line)
			continue
		}
		nt := normTrigger(t)
		if sameAsDefault(rows, nt, a) {
			continue // a line that repeats a default changes nothing
		}
		// Whatever was on this key before is replaced.
		kept := rows[:0]
		for i, r := range rows {
			if r.Enabled && normTrigger(r.Trigger) == nt {
				if r.Origin == "added" {
					continue
				}
				r.Enabled = false
				if a == "unbind" && r.Trigger == r.DefaultTrigger {
					unbound[i] = true
				}
			}
			kept = append(kept, r)
		}
		rows = kept
		if a == "unbind" {
			continue
		}
		// A default action that was unbound and is now bound to another key
		// is shown as that default with a changed key.
		merged := false
		for i := range rows {
			r := &rows[i]
			if r.Origin == "default" && !r.Enabled && unbound[i] && r.Action == a {
				r.Trigger, r.Enabled, r.Origin = t, true, "changed"
				merged = true
				break
			}
		}
		if !merged {
			rows = append(rows, KeyRow{ID: "a" + strconv.Itoa(added), Action: a, Trigger: t, Enabled: true, Origin: "added"})
			added++
		}
	}
	return rows, extra
}

func sameAsDefault(rows []KeyRow, nt, action string) bool {
	for _, r := range rows {
		if r.Origin == "default" && r.Enabled && r.Action == action && normTrigger(r.Trigger) == nt {
			return true
		}
	}
	return false
}

// keybindLines turns rows back into config lines: lines we do not model first,
// then unbinds of default keys, then the bindings that differ from defaults.
func keybindLines(rows []KeyRow, extra []string) []string {
	lines := append([]string{}, extra...)
	var binds []string
	for _, r := range rows {
		switch {
		case r.Origin == "added":
			if r.Enabled && r.Trigger != "" {
				binds = append(binds, r.Trigger+"="+r.Action)
			}
		case !r.Enabled:
			lines = append(lines, r.DefaultTrigger+"=unbind")
		case normTrigger(r.Trigger) != normTrigger(r.DefaultTrigger):
			lines = append(lines, r.DefaultTrigger+"=unbind")
			binds = append(binds, r.Trigger+"="+r.Action)
		}
	}
	return append(lines, binds...)
}

// Keybinds returns the shortcuts for the Keys tab.
func (a *App) Keybinds() KeybindState {
	a.mu.Lock()
	defer a.mu.Unlock()
	st := KeybindState{}
	if a.defaultKeys == nil {
		out, err := runGhostty(a.bin, "+list-keybinds", "--default")
		if err != nil {
			return st
		}
		a.defaultKeys = parseKeybindLines(out)
		if out, err := runGhostty(a.bin, "+list-actions"); err == nil {
			for _, l := range strings.Split(out, "\n") {
				if l = strings.TrimSpace(l); l != "" && l != "unbind" && l != "ignore" {
					a.actions = append(a.actions, l)
				}
			}
		}
	}
	if err := a.cfg.Load(); err != nil {
		return st
	}
	var user []string
	if v, ok := a.cfg.Values()["keybind"]; ok {
		user = strings.Split(v, "\n")
	}
	st.Rows, st.Extra = buildKeyRows(a.defaultKeys, user)
	st.Actions = a.actions
	return st
}

// SetKeybinds writes the rows edited in the Keys tab.
func (a *App) SetKeybinds(rows []KeyRow, extra []string) Result {
	return a.apply(func() { a.cfg.SetAll("keybind", keybindLines(rows, extra)) })
}
