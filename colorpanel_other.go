//go:build !darwin

package main

// ColorPanelOpen is always false off macOS: the GTK color dialog is modal and
// the UI gets "change" only after it closes.
func (a *App) ColorPanelOpen() bool { return false }
