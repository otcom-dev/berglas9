package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const logoColor = "#98E0EA"

type horizontalRule struct {
	*tview.Box
}

func newHorizontalRule() *horizontalRule {
	return &horizontalRule{Box: tview.NewBox()}
}

func (h *horizontalRule) Draw(screen tcell.Screen) {
	x, y, width, height := h.GetRect()
	if width <= 0 || height <= 0 {
		return
	}
	sw, sh := screen.Size()
	style := tcell.StyleDefault.Foreground(tview.Styles.BorderColor)
	if x > 0 && y >= 0 && y < sh {
		screen.SetContent(x-1, y, tview.BoxDrawingsLightVerticalAndRight, nil, style)
	}
	for i := 0; i < width; i++ {
		if x+i >= 0 && x+i < sw && y >= 0 && y < sh {
			screen.SetContent(x+i, y, tview.BoxDrawingsLightHorizontal, nil, style)
		}
	}
	if x+width >= 0 && x+width < sw && y >= 0 && y < sh {
		screen.SetContent(x+width, y, tview.BoxDrawingsLightVerticalAndLeft, nil, style)
	}
}

func secretUpdatedDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02")
}

func secretUpdatedStamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

type overlayRoot struct {
	*tview.Box
	content tview.Primitive
	overlay *tview.Primitive
}

func newOverlayRoot(content tview.Primitive, overlay *tview.Primitive) *overlayRoot {
	return &overlayRoot{Box: tview.NewBox(), content: content, overlay: overlay}
}

func (o *overlayRoot) SetRect(x, y, width, height int) {
	o.Box.SetRect(x, y, width, height)
	o.content.SetRect(x, y, width, height)
	if o.overlay != nil && *o.overlay != nil {
		(*o.overlay).SetRect(x, y, width, height)
	}
}

func (o *overlayRoot) Draw(screen tcell.Screen) {
	o.content.Draw(screen)
	if o.overlay != nil && *o.overlay != nil {
		(*o.overlay).Draw(screen)
	}
}

func (o *overlayRoot) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if o.overlay != nil && *o.overlay != nil {
			if h := (*o.overlay).InputHandler(); h != nil {
				h(event, setFocus)
				return
			}
		}
		if h := o.content.InputHandler(); h != nil {
			h(event, setFocus)
		}
	}
}

func (o *overlayRoot) Focus(delegate func(p tview.Primitive)) {
	if o.overlay != nil && *o.overlay != nil {
		delegate(*o.overlay)
	} else {
		delegate(o.content)
	}
}

func (o *overlayRoot) HasFocus() bool {
	if o.overlay != nil && *o.overlay != nil {
		return (*o.overlay).HasFocus()
	}
	return o.content.HasFocus()
}

type appUI struct {
	app      *tview.Application
	store    secretStore
	bucket   string
	overlay  *tview.Primitive
	secrets  *tview.Table
	meta     *tview.TextView
	value    *tview.TextView
	status   *tview.TextView
	items    []secretRef
	cache    map[string][]byte
	revealed string
	inModal  bool
}

func runTUI(store secretStore, bucket string) error {
	ui := newAppUI(store, bucket)
	ui.refresh()
	return ui.app.Run()
}

func newAppUI(store secretStore, bucket string) *appUI {
	ui := &appUI{
		app:    tview.NewApplication(),
		store:  store,
		bucket: bucket,
		cache:  map[string][]byte{},
	}
	var overlay tview.Primitive
	ui.overlay = &overlay

	logo := tview.NewTextView().SetDynamicColors(true).
		SetText(fmt.Sprintf("[%s::b]berglas9[-]", logoColor))
	location := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight).
		SetText(fmt.Sprintf("[%s]bucket[-] %s", logoColor, bucket))
	header := tview.NewFlex().
		AddItem(logo, 0, 1, false).
		AddItem(location, 0, 1, false)

	ui.secrets = tview.NewTable().SetSelectable(true, false).SetEvaluateAllRows(true)
	ui.secrets.SetBorder(true).SetTitle(" Secrets ")
	ui.secrets.SetSelectedStyle(tcell.StyleDefault.
		Foreground(tview.Styles.PrimitiveBackgroundColor).
		Background(tview.Styles.PrimaryTextColor))
	ui.secrets.SetSelectionChangedFunc(func(row, column int) {
		ui.revealed = ""
		ui.renderDetail()
	})

	ui.meta = tview.NewTextView()
	ui.value = tview.NewTextView().SetDynamicColors(false).SetScrollable(true).SetWrap(true)
	ui.value.SetBorderPadding(0, 0, 1, 1)
	detail := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.meta, 3, 0, false).
		AddItem(newHorizontalRule(), 1, 0, false).
		AddItem(ui.value, 0, 1, false)
	detail.SetBorder(true).SetTitle(" Secret ")

	body := tview.NewFlex().
		AddItem(ui.secrets, 0, 2, true).
		AddItem(detail, 0, 3, false)

	keys := tview.NewTextView().SetDynamicColors(true).
		SetText(fmt.Sprintf("[%s]%s[-]", logoColor, keyHintLine()))
	ui.status = tview.NewTextView().SetDynamicColors(true)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(keys, 1, 0, false).
		AddItem(ui.status, 1, 0, false)

	root := newOverlayRoot(content, ui.overlay)
	ui.app.SetRoot(root, true).SetFocus(ui.secrets)
	ui.app.SetInputCapture(ui.capture)
	ui.renderDetail()
	return ui
}

func (ui *appUI) capture(event *tcell.EventKey) *tcell.EventKey {
	if ui.inModal {
		return event
	}
	if isCopyEvent(event) {
		ui.copySelected()
		return nil
	}
	switch event.Rune() {
	case 'q':
		ui.app.Stop()
		return nil
	case 'm', 'M':
		ui.toggleReveal()
		return nil
	case 'n':
		ui.openCreate()
		return nil
	case 'e':
		ui.openEdit()
		return nil
	case 'r':
		ui.refresh()
		return nil
	case 'h', '?':
		ui.openHelp()
		return nil
	}
	return event
}

func (ui *appUI) setStatus(msg string) {
	ui.status.SetText(" " + msg)
}

func (ui *appUI) selected() (secretRef, bool) {
	if ui.secrets.GetRowCount() == 0 || len(ui.items) == 0 {
		return secretRef{}, false
	}
	idx, _ := ui.secrets.GetSelection()
	if idx < 0 || idx >= len(ui.items) {
		return secretRef{}, false
	}
	return ui.items[idx], true
}

func cacheKey(ref secretRef) string {
	return fmt.Sprintf("%s#%d", ref.Name, ref.Generation)
}

func (ui *appUI) renderList() {
	keep := ""
	if ref, ok := ui.selected(); ok {
		keep = ref.Name
	}
	ui.secrets.Clear()
	next := 0
	for i, item := range ui.items {
		ui.secrets.SetCell(i, 0, tview.NewTableCell(item.Name).SetExpansion(0))
		ui.secrets.SetCell(i, 1, tview.NewTableCell(secretUpdatedDate(item.UpdatedAt)).
			SetAlign(tview.AlignRight).
			SetExpansion(1))
		if item.Name == keep {
			next = i
		}
	}
	if len(ui.items) > 0 {
		ui.secrets.Select(next, 0)
	}
	ui.renderDetail()
}

func (ui *appUI) renderDetail() {
	ref, ok := ui.selected()
	if !ok {
		ui.meta.SetText(" No secrets in this bucket.\n Press n to create one.")
		ui.value.SetText("")
		return
	}
	ui.meta.SetText(fmt.Sprintf(" Name:        %s\n Generation:  %d\n Updated:     %s", ref.Name, ref.Generation, secretUpdatedStamp(ref.UpdatedAt)))
	revealed := ui.revealed == ref.Name
	var pt []byte
	if revealed {
		pt = ui.cache[cacheKey(ref)]
	}
	ui.value.SetText(displayValue(pt, revealed))
}

func (ui *appUI) refresh() {
	ui.setStatus("Loading…")
	go func() {
		items, err := ui.store.List(context.Background())
		ui.app.QueueUpdateDraw(func() {
			if err != nil {
				ui.setStatus(err.Error())
				return
			}
			ui.items = items
			ui.revealed = ""
			ui.renderList()
			ui.setStatus(fmt.Sprintf("%d secrets", len(items)))
		})
	}()
}

func (ui *appUI) toggleReveal() {
	ref, ok := ui.selected()
	if !ok {
		ui.setStatus("No secret selected")
		return
	}
	if ui.revealed == ref.Name {
		ui.revealed = ""
		ui.renderDetail()
		ui.setStatus("Masked " + ref.Name)
		return
	}
	if _, ok := ui.cache[cacheKey(ref)]; ok {
		ui.revealed = ref.Name
		ui.renderDetail()
		return
	}
	ui.setStatus("Loading " + ref.Name + "…")
	go func() {
		pt, err := ui.store.Access(context.Background(), ref.Name, ref.Generation)
		ui.app.QueueUpdateDraw(func() {
			if err != nil {
				ui.setStatus(err.Error())
				return
			}
			ui.cache[cacheKey(ref)] = append([]byte(nil), pt...)
			ui.revealed = ref.Name
			ui.renderDetail()
			ui.setStatus("")
		})
	}()
}

func (ui *appUI) copySelected() {
	ref, ok := ui.selected()
	if !ok {
		ui.setStatus("No secret selected")
		return
	}
	if pt, ok := ui.cache[cacheKey(ref)]; ok {
		ui.writeClipboard(ref.Name, pt)
		return
	}
	ui.setStatus("Loading " + ref.Name + "…")
	go func() {
		pt, err := ui.store.Access(context.Background(), ref.Name, ref.Generation)
		ui.app.QueueUpdateDraw(func() {
			if err != nil {
				ui.setStatus(err.Error())
				return
			}
			ui.cache[cacheKey(ref)] = append([]byte(nil), pt...)
			ui.writeClipboard(ref.Name, pt)
		})
	}()
}

func (ui *appUI) writeClipboard(name string, pt []byte) {
	secret := string(pt)
	go func() {
		err := clipboard.WriteAll(secret)
		ui.app.QueueUpdateDraw(func() {
			if err != nil {
				ui.setStatus("Clipboard unavailable")
				return
			}
			ui.setStatus(copiedStatus(name))
		})
	}()
}

func (ui *appUI) showModal(p tview.Primitive, width, height int) {
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(p, width, 1, true).
			AddItem(nil, 0, 1, false), height, 1, true).
		AddItem(nil, 0, 1, false)
	ui.inModal = true
	*ui.overlay = modal
}

func (ui *appUI) closeModal() {
	ui.inModal = false
	*ui.overlay = nil
	ui.app.SetFocus(ui.secrets)
}

func (ui *appUI) openHelp() {
	text := tview.NewTextView().SetText(helpBody()).SetDynamicColors(false)
	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(text, 0, 1, true).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(" Esc / "+copyChordLabel()+" close "), 1, 0, false)
	box.SetBorder(true).SetTitle(" Help ")
	close := func() {
		ui.closeModal()
	}
	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || isDismiss(event) {
			close()
			return nil
		}
		if event.Rune() == 'q' || event.Rune() == 'h' {
			close()
			return nil
		}
		return event
	})
	ui.showModal(box, 62, 19)
	ui.app.SetFocus(box)
}

func (ui *appUI) openCreate() {
	var name, value string
	step := 0
	box := tview.NewFlex().SetDirection(tview.FlexRow)
	box.SetBorder(true).SetTitle(" New secret ")

	var show func()
	show = func() {
		box.Clear()
		if step == 0 {
			input := tview.NewInputField().SetLabel("Name  ").SetText(name).SetFieldWidth(48)
			input.SetChangedFunc(func(text string) { name = text })
			hint := tview.NewTextView().SetText(" Enter next    Esc / " + copyChordLabel() + " cancel ")
			advance := func() {
				if strings.TrimSpace(name) == "" {
					hint.SetText(" Name is required.  Enter next    Esc / " + copyChordLabel() + " cancel ")
					ui.setStatus("Name is required")
					return
				}
				step = 1
				show()
			}
			input.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEnter {
					advance()
				}
			})
			input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape || isDismiss(event) {
					ui.closeModal()
					return nil
				}
				return event
			})
			box.AddItem(input, 1, 0, true)
			box.AddItem(hint, 1, 0, false)
			ui.showModal(box, 68, 7)
			ui.app.SetFocus(input)
			return
		}

		label := tview.NewTextView().SetDynamicColors(false).SetText("Name  " + trimName(name))
		area := tview.NewTextArea().SetText(value, true)
		hint := tview.NewTextView().SetText(" Ctrl+S create    Esc back    " + copyChordLabel() + " cancel ")
		submit := func() {
			value = area.GetText()
			object, plaintext, err := buildCreate(name, value)
			if err != nil {
				hint.SetText(" Value is required.  Ctrl+S create    Esc back    " + copyChordLabel() + " cancel ")
				ui.setStatus("Value is required")
				return
			}
			ui.closeModal()
			ui.setStatus("Creating " + object + "…")
			go func() {
				err := ui.store.Create(context.Background(), object, plaintext)
				ui.app.QueueUpdateDraw(func() {
					if err != nil {
						ui.setStatus(err.Error())
						return
					}
					ui.setStatus("Created " + object)
					ui.refresh()
				})
			}()
		}
		capture := func(event *tcell.EventKey) *tcell.EventKey {
			if isDismiss(event) {
				ui.closeModal()
				return nil
			}
			switch event.Key() {
			case tcell.KeyEscape:
				value = area.GetText()
				step = 0
				show()
				return nil
			case tcell.KeyCtrlS:
				submit()
				return nil
			}
			return event
		}
		area.SetInputCapture(capture)
		box.AddItem(label, 1, 0, false)
		box.AddItem(area, 0, 1, true)
		box.AddItem(hint, 1, 0, false)
		ui.showModal(box, 68, 14)
		ui.app.SetFocus(area)
	}
	show()
}

func (ui *appUI) openEdit() {
	ref, ok := ui.selected()
	if !ok {
		ui.setStatus("No secret selected")
		return
	}
	ui.setStatus("Loading " + ref.Name + "…")
	go func() {
		pt, err := ui.store.Access(context.Background(), ref.Name, ref.Generation)
		ui.app.QueueUpdateDraw(func() {
			if err != nil {
				ui.setStatus(err.Error())
				return
			}
			ui.cache[cacheKey(ref)] = append([]byte(nil), pt...)
			ui.showEdit(ref, string(pt))
		})
	}()
}

func (ui *appUI) showEdit(ref secretRef, current string) {
	area := tview.NewTextArea().SetText(current, false)
	label := tview.NewTextView().SetDynamicColors(false).SetText("Name  " + ref.Name)
	hint := tview.NewTextView().SetText(" Ctrl+S save    Esc / " + copyChordLabel() + " cancel ")
	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(label, 1, 0, false).
		AddItem(area, 0, 1, true).
		AddItem(hint, 1, 0, false)
	box.SetBorder(true).SetTitle(" Edit secret ")
	save := func() {
		value := area.GetText()
		if strings.TrimSpace(value) == "" {
			hint.SetText(" Value is required.  Ctrl+S save    Esc / " + copyChordLabel() + " cancel ")
			ui.setStatus("Value is required")
			return
		}
		plaintext := []byte(value)
		ui.closeModal()
		ui.setStatus("Updating " + ref.Name + "…")
		go func() {
			err := ui.store.Update(context.Background(), ref, plaintext)
			ui.app.QueueUpdateDraw(func() {
				if err != nil {
					ui.setStatus(err.Error())
					return
				}
				delete(ui.cache, cacheKey(ref))
				ui.revealed = ""
				ui.setStatus("Updated " + ref.Name)
				ui.refresh()
			})
		}()
	}
	capture := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || isDismiss(event) {
			ui.closeModal()
			ui.setStatus("")
			return nil
		}
		switch event.Key() {
		case tcell.KeyCtrlS:
			save()
			return nil
		}
		return event
	}
	area.SetInputCapture(capture)
	ui.showModal(box, 68, 14)
	ui.app.SetFocus(area)
}

func trimName(name string) string {
	return strings.TrimSpace(name)
}
