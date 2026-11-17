//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

static volatile int recordingKeys = 0;
static id keyMonitor = nil;

// While the Keys tab records a shortcut, hand ⌘ key presses straight to the
// focused view. Otherwise AppKit treats them as key equivalents first, and
// some (such as ⌘W and ⌘,) never reach the page.
static void setKeyRecording(int on) {
	recordingKeys = on;
	if (!on || keyMonitor != nil) return;
	dispatch_async(dispatch_get_main_queue(), ^{
		if (keyMonitor != nil) return;
		keyMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown handler:^NSEvent *(NSEvent *event) {
			if (!recordingKeys || !([event modifierFlags] & NSEventModifierFlagCommand)) return event;
			NSResponder *target = [[event window] firstResponder];
			if (target == nil) return event;
			[target keyDown:event];
			return nil;
		}];
	});
}
*/
import "C"

// SetKeyRecording tells macOS whether the Keys tab is recording a shortcut, so
// ⌘ shortcuts reach the page instead of being taken as key equivalents.
func (a *App) SetKeyRecording(on bool) {
	v := 0
	if on {
		v = 1
	}
	C.setKeyRecording(C.int(v))
}
