package main

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GTK 3 does not read the desktop's dark style setting (GNOME's color-scheme
// or the XDG desktop portal), so on Linux WebKitGTK reports
// prefers-color-scheme: light even when the desktop is dark. Before GTK starts,
// GTK_THEME is set to the dark variant of the current theme. A GTK_THEME the
// user set is left alone. The style is read once at launch.
func applyDarkPreference() {
	if runtime.GOOS != "linux" || os.Getenv("GTK_THEME") != "" {
		return
	}
	portal, _ := exec.Command("gdbus", "call", "--session", "--timeout", "2",
		"--dest", "org.freedesktop.portal.Desktop", "--object-path", "/org/freedesktop/portal/desktop",
		"--method", "org.freedesktop.portal.Settings.ReadOne", "org.freedesktop.appearance", "color-scheme").Output()
	scheme, _ := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	theme, _ := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme").Output()
	if v := darkGTKTheme(string(portal), string(scheme), string(theme)); v != "" {
		os.Setenv("GTK_THEME", v)
	}
}

// The portal answers e.g. "(<uint32 1>,)": 0 no preference, 1 dark, 2 light.
var portalSchemeRe = regexp.MustCompile(`uint32 (\d)`)

// darkGTKTheme returns the GTK_THEME to use, or "" to leave GTK as it is.
// portal is the gdbus output, scheme and theme the gsettings output (quoted).
// The portal wins; GNOME's setting is used when the portal is not available.
func darkGTKTheme(portal, scheme, theme string) string {
	var dark bool
	if m := portalSchemeRe.FindStringSubmatch(portal); m != nil {
		dark = m[1] == "1"
	} else {
		dark = strings.Contains(scheme, "prefer-dark")
	}
	if !dark {
		return ""
	}
	theme = strings.Trim(strings.TrimSpace(theme), "'")
	if theme == "" {
		theme = "Adwaita"
	}
	if strings.Contains(strings.ToLower(theme), "dark") {
		return "" // already a dark theme
	}
	return theme + ":dark"
}
