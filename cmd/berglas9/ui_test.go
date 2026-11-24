package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestMainLayout(t *testing.T) {
	updated := time.Date(2026, 11, 24, 15, 4, 5, 0, time.UTC)
	ui := newAppUI(&memStore{}, "my-bucket")
	screen := tcell.NewSimulationScreen("")
	ui.app.SetScreen(screen)
	screen.SetSize(100, 24)

	ui.items = []secretRef{{
		Name:       "api-key",
		Generation: 1771234567890123,
		UpdatedAt:  updated,
	}}
	ui.renderList()
	ui.app.ForceDraw()

	got := screenText(screen)
	if strings.Contains(got, "secrets stay masked") {
		t.Fatalf("header still has the tagline:\n%s", got)
	}
	if !strings.Contains(got, "berglas9") || !strings.Contains(got, "my-bucket") {
		t.Fatalf("header missing title or bucket:\n%s", got)
	}
	if !strings.Contains(got, "m mask/unmask") {
		t.Fatalf("toolbar missing mask/unmask:\n%s", got)
	}
	if strings.Contains(got, "v reveal") {
		t.Fatalf("toolbar still says reveal:\n%s", got)
	}
	listLine := lineContaining(got, "api-key")
	if strings.Contains(listLine, "1771234567890123") || strings.Contains(strings.ToLower(listLine), "generation") {
		t.Fatalf("list still shows generation: %q", listLine)
	}
	if !strings.Contains(listLine, secretUpdatedDate(updated)) || strings.Contains(listLine, "15:04:05") {
		t.Fatalf("list should show the date without the time: %q", listLine)
	}
	if !strings.Contains(got, secretUpdatedStamp(updated)) {
		t.Fatalf("detail missing updated time:\n%s", got)
	}
	if strings.Contains(got, "Value") {
		t.Fatalf("value still has its own frame:\n%s", got)
	}
	if !strings.Contains(got, "├") || !strings.Contains(got, "┤") {
		t.Fatalf("detail is missing the divider:\n%s", got)
	}
	if !strings.Contains(got, maskedSecret) {
		t.Fatalf("selected value is not masked:\n%s", got)
	}
}

func lineContaining(text, needle string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

func screenText(screen tcell.SimulationScreen) string {
	width, height := screen.Size()
	var b strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mainc, _, _, _ := screen.GetContent(x, y)
			if mainc == 0 {
				b.WriteRune(' ')
			} else {
				b.WriteRune(mainc)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
