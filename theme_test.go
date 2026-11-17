package main

import "testing"

func TestDarkGTKTheme(t *testing.T) {
	tests := []struct {
		name, portal, scheme, theme, want string
	}{
		{"portal says dark", "(<uint32 1>,)\n", "'default'\n", "'Yaru'\n", "Yaru:dark"},
		{"older portal output", "(<<uint32 1>>,)\n", "", "'Adwaita'\n", "Adwaita:dark"},
		{"portal says light", "(<uint32 2>,)\n", "'prefer-dark'\n", "'Adwaita'\n", ""}, // the portal wins
		{"no preference", "(<uint32 0>,)\n", "", "'Adwaita'\n", ""},
		{"no portal, GNOME dark", "", "'prefer-dark'\n", "'Adwaita'\n", "Adwaita:dark"},
		{"no portal, GNOME light", "", "'default'\n", "'Adwaita'\n", ""},
		{"theme unknown", "(<uint32 1>,)\n", "", "", "Adwaita:dark"},
		{"already a dark theme", "(<uint32 1>,)\n", "", "'Yaru-dark'\n", ""},
	}
	for _, tt := range tests {
		if got := darkGTKTheme(tt.portal, tt.scheme, tt.theme); got != tt.want {
			t.Errorf("%s: darkGTKTheme() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
