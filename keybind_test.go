package main

import (
	"reflect"
	"testing"
)

func TestSplitBinding(t *testing.T) {
	tests := []struct{ in, trig, act string }{
		{"super+t=new_tab", "super+t", "new_tab"},
		{"super+==increase_font_size:1", "super+=", "increase_font_size:1"},
		{"super++=increase_font_size:1", "super++", "increase_font_size:1"},
		{"super+ctrl+==equalize_splits", "super+ctrl+=", "equalize_splits"},
		{"super+shift+,=reload_config", "super+shift+,", "reload_config"},
		{"ctrl+a>n=new_window", "ctrl+a>n", "new_window"},
		{"global:super+grave_accent=toggle_quick_terminal", "global:super+grave_accent", "toggle_quick_terminal"},
		{"super+ctrl+f=text:\\x1b=x", "super+ctrl+f", "text:\\x1b=x"},
	}
	for _, tt := range tests {
		trig, act, ok := splitBinding(tt.in)
		if !ok || trig != tt.trig || act != tt.act {
			t.Errorf("splitBinding(%q) = %q, %q, %v; want %q, %q", tt.in, trig, act, ok, tt.trig, tt.act)
		}
	}
	if _, _, ok := splitBinding("clear"); ok {
		t.Errorf("clear should not parse as a binding")
	}
}

func TestNormTrigger(t *testing.T) {
	same := [][2]string{
		{"shift+super+t", "super+shift+t"},
		{"cmd+T", "super+t"},
		{"opt+command+w", "alt+super+w"},
		{"control+a>n", "ctrl+a>n"},
		{"Global:cmd+grave_accent", "global:super+grave_accent"},
		{"super++", "super++"},
	}
	for _, p := range same {
		if a, b := normTrigger(p[0]), normTrigger(p[1]); a != b {
			t.Errorf("normTrigger(%q) = %q, normTrigger(%q) = %q; want equal", p[0], a, p[1], b)
		}
	}
	if normTrigger("super+=") == normTrigger("super++") {
		t.Errorf("= and + keys must differ")
	}
}

var testDefaults = [][2]string{
	{"super+t", "new_tab"},
	{"super+d", "new_split:right"},
	{"super+shift+d", "new_split:down"},
	{"super+9", "last_tab"},
}

func rowsByID(rows []KeyRow) map[string]KeyRow {
	m := map[string]KeyRow{}
	for _, r := range rows {
		m[r.ID] = r
	}
	return m
}

func TestBuildKeyRowsDefaultsOnly(t *testing.T) {
	rows, extra := buildKeyRows(testDefaults, nil)
	if len(rows) != 4 || len(extra) != 0 {
		t.Fatalf("rows = %v, extra = %v", rows, extra)
	}
	for _, r := range rows {
		if !r.Enabled || r.Origin != "default" || r.Trigger != r.DefaultTrigger {
			t.Errorf("row %+v", r)
		}
	}
	if got := keybindLines(rows, extra); len(got) != 0 {
		t.Errorf("unchanged defaults should write nothing, got %v", got)
	}
}

func TestBuildKeyRowsUserChanges(t *testing.T) {
	user := []string{
		"clear-like-unknown",
		"super+9=unbind",                      // disable last_tab
		"super+shift+d=unbind",                // move new_split:down ...
		"ctrl+shift+d=new_split:down",         // ... to ctrl+shift+d
		"super+backslash=toggle_tab_overview", // add
		"super+t=new_tab",                     // repeats a default: no change
	}
	rows, extra := buildKeyRows(testDefaults, user)
	m := rowsByID(rows)
	if !reflect.DeepEqual(extra, []string{"clear-like-unknown"}) {
		t.Errorf("extra = %v", extra)
	}
	if r := m["d0"]; !r.Enabled || r.Origin != "default" {
		t.Errorf("new_tab = %+v", r)
	}
	if r := m["d3"]; r.Enabled || r.Origin != "default" {
		t.Errorf("last_tab should be disabled: %+v", r)
	}
	if r := m["d2"]; !r.Enabled || r.Origin != "changed" || r.Trigger != "ctrl+shift+d" || r.DefaultTrigger != "super+shift+d" {
		t.Errorf("new_split:down should be moved: %+v", r)
	}
	if r := m["a0"]; !r.Enabled || r.Origin != "added" || r.Action != "toggle_tab_overview" {
		t.Errorf("added row = %+v", r)
	}
	if len(rows) != 5 {
		t.Errorf("rows = %d, want 5", len(rows))
	}
}

func TestBindingOverDefaultKeyDisablesThatDefault(t *testing.T) {
	rows, _ := buildKeyRows(testDefaults, []string{"super+d=new_split:down"})
	m := rowsByID(rows)
	if m["d1"].Enabled {
		t.Errorf("new_split:right on super+d should be replaced: %+v", m["d1"])
	}
	if r := m["a0"]; r.Trigger != "super+d" || r.Action != "new_split:down" || !r.Enabled {
		t.Errorf("added = %+v", r)
	}
}

func TestKeybindLinesRoundTrip(t *testing.T) {
	user := []string{"super+9=unbind", "super+shift+d=unbind", "ctrl+shift+d=new_split:down", "super+backslash=toggle_tab_overview"}
	rows, extra := buildKeyRows(testDefaults, user)
	lines := keybindLines(rows, extra)
	want := []string{"super+shift+d=unbind", "super+9=unbind", "ctrl+shift+d=new_split:down", "super+backslash=toggle_tab_overview"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("lines = %v, want %v", lines, want)
	}
	again, _ := buildKeyRows(testDefaults, lines)
	if !reflect.DeepEqual(again, rows) {
		t.Errorf("round trip changed rows:\n%+v\n%+v", again, rows)
	}
}

func TestKeybindLinesFromEdits(t *testing.T) {
	rows, _ := buildKeyRows(testDefaults, nil)
	rows[0].Trigger = "super+n" // move new_tab
	rows[3].Enabled = false     // disable last_tab
	rows = append(rows, KeyRow{ID: "a0", Action: "toggle_tab_overview", Trigger: "super+backslash", Enabled: true, Origin: "added"})
	rows = append(rows, KeyRow{ID: "a1", Action: "quit", Trigger: "", Enabled: true, Origin: "added"}) // no key yet: skipped
	got := keybindLines(rows, nil)
	want := []string{"super+t=unbind", "super+9=unbind", "super+n=new_tab", "super+backslash=toggle_tab_overview"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lines = %v, want %v", got, want)
	}
	// Moving a default back to its own key writes nothing for it.
	rows, _ = buildKeyRows(testDefaults, []string{"super+t=unbind", "super+n=new_tab"})
	rows[0].Trigger = "super+t"
	if got := keybindLines(rows, nil); len(got) != 0 {
		t.Errorf("restored default should write nothing, got %v", got)
	}
}

func TestSetKeybinds(t *testing.T) {
	a := newTestApp(t, "theme = Nord\nkeybind = super+9=unbind\n", 0, reloadOK)
	a.defaultKeys = testDefaults
	st := a.Keybinds()
	m := rowsByID(st.Rows)
	if m["d3"].Enabled {
		t.Fatalf("last_tab should start disabled: %+v", m["d3"])
	}
	rows := st.Rows
	for i := range rows {
		if rows[i].ID == "d3" {
			rows[i].Enabled = true
		}
		if rows[i].ID == "d1" {
			rows[i].Trigger = "ctrl+d"
		}
	}
	if res := a.SetKeybinds(rows, st.Extra); !res.OK {
		t.Fatalf("SetKeybinds: %+v", res)
	}
	if got := readFile(t, a.cfg.Path); got != "theme = Nord\nkeybind = super+d=unbind\nkeybind = ctrl+d=new_split:right\n" {
		t.Errorf("config = %q", got)
	}
}

// Ghostty binds some defaults twice (super+digit_1 and super+1); the UI moves
// one of them and disables the other.
func TestTwinDefaultsRoundTrip(t *testing.T) {
	defaults := [][2]string{{"super+digit_1", "goto_tab:1"}, {"super+1", "goto_tab:1"}}
	rows, _ := buildKeyRows(defaults, nil)
	rows[0].Trigger = "ctrl+1"
	rows[1].Enabled = false
	lines := keybindLines(rows, nil)
	want := []string{"super+digit_1=unbind", "super+1=unbind", "ctrl+1=goto_tab:1"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("lines = %v, want %v", lines, want)
	}
	again, _ := buildKeyRows(defaults, lines)
	rows[0].Origin = "changed" // the moved binding comes back marked as changed
	if !reflect.DeepEqual(again, rows) {
		t.Errorf("round trip changed rows:\n%+v\n%+v", again, rows)
	}
}
