package main

import (
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func copyChordLabel() string {
	if runtime.GOOS == "darwin" {
		return "⌘C"
	}
	return "Ctrl+C"
}

func keyHintLine() string {
	return " ↑/↓ select   m mask/unmask   " + copyChordLabel() + " copy   n new   e edit   r refresh   h help   q quit "
}

func helpBody() string {
	chord := padLabel(copyChordLabel(), 10)
	text := "berglas9\n\n" +
		"↑/↓       Select a secret\n" +
		"m         Mask or unmask the selected value\n" +
		chord + "Copy the selected secret (main screen)\n" +
		"n         New secret: name, then value\n" +
		"e         Edit the selected secret\n" +
		"r         Refresh\n" +
		"h         Help\n" +
		"q         Quit\n"
	if runtime.GOOS == "darwin" {
		text += "Ctrl+C    Quit\n"
	}
	text += "\nCreate and edit\n" +
		"Ctrl+S    Save\n" +
		"Esc       Back (create value) or cancel\n" +
		chord + "Cancel. Does not copy.\n"
	return text
}

func padLabel(label string, width int) string {
	n := len([]rune(label))
	if n >= width {
		return label
	}
	return label + strings.Repeat(" ", width-n)
}

// isCopyEvent reports the platform copy shortcut.
// macOS uses Command+C. Other systems use Ctrl+C.
func isCopyEvent(event *tcell.EventKey) bool {
	return isCopyEventOS(runtime.GOOS, event)
}

func isCopyEventOS(goos string, event *tcell.EventKey) bool {
	if event == nil {
		return false
	}
	if goos == "darwin" {
		return isCommandC(event)
	}
	return event.Key() == tcell.KeyCtrlC
}

// isCommandC reports Command+C as terminals deliver it through the kitty
// keyboard protocol: the Super modifier, which tcell calls ModMeta.
func isCommandC(event *tcell.EventKey) bool {
	if event.Modifiers()&tcell.ModMeta == 0 {
		return false
	}
	r := event.Rune()
	return event.Key() == tcell.KeyRune && (r == 'c' || r == 'C')
}

func isDismiss(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyCtrlC || isCopyEvent(event)
}
