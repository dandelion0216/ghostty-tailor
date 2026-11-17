//go:build !darwin

package main

// SetKeyRecording does nothing off macOS: there are no key equivalents that
// take shortcuts before the page sees them.
func (a *App) SetKeyRecording(on bool) {}
