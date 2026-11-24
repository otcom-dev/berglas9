package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCopyEventUsesCommandCOnDarwin(t *testing.T) {
	cmdC := tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModMeta)
	cmdShiftC := tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModMeta|tcell.ModShift)
	ctrlC := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)
	plainC := tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone)
	altC := tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt)

	if !isCopyEventOS("darwin", cmdC) || !isCopyEventOS("darwin", cmdShiftC) {
		t.Fatal("darwin Command+C was not copy")
	}
	if isCopyEventOS("darwin", ctrlC) || isCopyEventOS("darwin", plainC) || isCopyEventOS("darwin", altC) {
		t.Fatal("darwin treated a non-Command chord as copy")
	}
	if !isCopyEventOS("linux", ctrlC) {
		t.Fatal("linux Ctrl+C was not copy")
	}
	if isCopyEventOS("linux", cmdC) || isCopyEventOS("linux", plainC) {
		t.Fatal("linux treated a non-Ctrl chord as copy")
	}
}

func TestCopyChordLabel(t *testing.T) {
	if got := copyChordLabel(); got != "Ctrl+C" && got != "⌘C" {
		t.Fatalf("label = %q", got)
	}
	if !isCopyEvent(tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)) && copyChordLabel() != "⌘C" {
		t.Fatal("label and event disagree")
	}
}
