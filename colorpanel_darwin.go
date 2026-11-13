//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

static int colorPanelVisible(void) {
	__block int visible = 0;
	void (^check)(void) = ^{
		visible = [NSColorPanel sharedColorPanelExists] && [[NSColorPanel sharedColorPanel] isVisible];
	};
	if ([NSThread isMainThread]) {
		check();
	} else {
		dispatch_sync(dispatch_get_main_queue(), check);
	}
	return visible;
}
*/
import "C"

// ColorPanelOpen reports whether the macOS color panel is on screen. WebKit
// sends "change" for every pick, so the UI asks this to save a color only after
// the user closes the panel.
func (a *App) ColorPanelOpen() bool {
	return C.colorPanelVisible() != 0
}
