package domain

import (
	"testing"
	"time"
)

// TestProcess_CommandWindowWhenGridLossTimeUnknown covers the black-start
// command window when the dispatcher issues a real command without recording
// when the main grid was actually lost. With no reference instant there is
// nothing to measure the window against, so the command must not be reported
// as overdue. Commands that do carry a grid-loss time are still judged.
func TestProcess_CommandWindowWhenGridLossTimeUnknown(t *testing.T) {
	now := fixedTime()

	unknown := NewProcess("bsp-unknown", KindReal, now, time.Time{}, DefaultSettings())
	if unknown.CommandOverdueRecorded {
		t.Error("a real command with no recorded grid-loss time must not be flagged overdue")
	}
	if hasEvent(unknown, EventBlackStartOverdue) {
		t.Error("no black_start_overdue event expected when the grid-loss time is unknown")
	}
	for _, e := range unknown.Events {
		if e.Type == EventBlackStartCommanded {
			continue
		}
		t.Errorf("unexpected event %q on a freshly commanded process", e.Type)
	}

	inWindow := NewProcess("bsp-in-window", KindReal, now, now.Add(-time.Minute), DefaultSettings())
	if inWindow.CommandOverdueRecorded {
		t.Error("a command issued 1 minute after grid loss is inside the 5 minute window")
	}

	late := NewProcess("bsp-late", KindReal, now, now.Add(-6*time.Minute), DefaultSettings())
	if !late.CommandOverdueRecorded {
		t.Error("a command issued 6 minutes after grid loss is outside the 5 minute window")
	}
	if !hasEvent(late, EventBlackStartOverdue) {
		t.Error("expected a black_start_overdue event for a command outside the window")
	}
}
