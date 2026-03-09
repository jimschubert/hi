package ui

import (
	"context"
	"fmt"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jimschubert/hi/assets"
	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon/store"
)

type Hi struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	trayMenu   *fyne.Menu
	queueLabel *fyne.MenuItem
	queue      *store.Queue
	dialogOpen atomic.Bool
	cancel     context.CancelFunc
	ctx        context.Context
}

func NewHi(ctx context.Context, queue *store.Queue) *Hi {
	// structure of this function modeled after https://github.com/fyne-io/flatpak_demo/blob/main/main.go
	a := app.NewWithID("dev.jimschubert.hi")
	a.Settings().SetTheme(&hiTheme{})
	a.SetIcon(assets.Icon)

	// Hidden main window — required for systray on some platforms.
	w := a.NewWindow("Hi — Human Intelligence")
	w.SetCloseIntercept(func() {
		// hide instead of exit
		w.Hide()
	})
	w.Resize(fyne.NewSize(1, 1))
	w.Hide()

	ctx, cancel := context.WithCancel(ctx)

	hi := &Hi{fyneApp: a, mainWindow: w, queue: queue, cancel: cancel, ctx: ctx}
	return hi
}

func (h *Hi) Notify(title, body string) {
	h.fyneApp.SendNotification(fyne.NewNotification(title, body))
}

func (h *Hi) Run() {
	h.buildTrayMenu(h.ctx)

	go h.watchQueue(h.ctx)

	go func() {
		<-h.ctx.Done()
		fyne.DoAndWait(h.fyneApp.Quit)
	}()

	h.fyneApp.Run()
}

func (h *Hi) buildTrayMenu(ctx context.Context) {
	h.queueLabel = fyne.NewMenuItem("No pending requests", nil)
	h.queueLabel.Disabled = true

	showItem := fyne.NewMenuItem("Show next request", func() {
		h.dialogOpen.Store(true)
		h.presentNextRequest(ctx, func() {
			h.dialogOpen.Store(false)
		})
	})

	settingsItem := fyne.NewMenuItem("Settings…", func() {
		h.showSettings()
	})

	quitItem := fyne.NewMenuItem("Quit hi", func() {
		h.cancel()
	})

	h.trayMenu = fyne.NewMenu("Hi",
		h.queueLabel,
		fyne.NewMenuItemSeparator(),
		showItem,
		settingsItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)

	if desk, ok := h.fyneApp.(desktop.App); ok {
		desk.SetSystemTrayMenu(h.trayMenu)
		desk.SetSystemTrayIcon(theme.NewThemedResource(assets.Tray))
	}
}

func (h *Hi) watchQueue(ctx context.Context) {
	h.dialogOpen.Store(false)

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.queue.Notify():
			n := h.queue.Len()
			fyne.Do(func() { h.updateTrayLabel(n) })

			// This will automatically open the first dialog…
			if n > 0 && !h.dialogOpen.Load() {
				h.dialogOpen.Store(true)
				fyne.Do(func() {
					h.presentNextRequest(ctx, func() { h.dialogOpen.Store(false) })
				})
			}
		}
	}
}

func (h *Hi) updateTrayLabel(n int) {
	if h.queueLabel == nil {
		return
	}
	switch n {
	case 0:
		h.queueLabel.Label = "No pending requests"
	case 1:
		h.queueLabel.Label = "1 request waiting"
	default:
		h.queueLabel.Label = fmt.Sprintf("%d requests waiting", n)
	}
	h.trayMenu.Refresh()
}

func (h *Hi) presentNextRequest(ctx context.Context, done func()) {
	reqs := h.queue.Peek()
	if len(reqs) == 0 {
		done()
		return
	}
	req := reqs[0]

	switch req.Type {
	case store.RequestTypeText:
		h.showTextDialog(ctx, req, done)
	case store.RequestTypeMultiline:
		h.showMultilineDialog(ctx, req, done)
	case store.RequestTypeChoice:
		h.showChoiceDialog(ctx, req, done)
	case store.RequestTypeConfirm:
		h.showConfirmDialog(ctx, req, done)
	case store.RequestTypeNotify:
		h.queue.Respond(req.ID, store.Response{})
		done()
	}
}

func (h *Hi) showTextDialog(_ context.Context, req *store.PendingRequest, done func()) {
	w := h.mainWindow
	entry := widget.NewEntry()
	h.showTextEntry(entry, w, req, 260, 520, done)
}

func (h *Hi) showMultilineDialog(_ context.Context, req *store.PendingRequest, done func()) {
	w := h.mainWindow
	entry := widget.NewMultiLineEntry()
	entry.SetMinRowsVisible(4)
	h.showTextEntry(entry, w, req, 320, 560, done)
}

// showText displays immediately
// multiline does not
// single choice does
// multi-choice does not
// confirm does
// notification does not

func (h *Hi) showTextEntry(entry *widget.Entry, w fyne.Window, req *store.PendingRequest, height float32, width float32, done func()) {
	entry.SetText(req.DefaultVal)
	entry.SetPlaceHolder("Type your response…")
	entry.Wrapping = fyne.TextWrapWord

	promptLabel := widget.NewLabel(req.Prompt)
	promptLabel.Wrapping = fyne.TextWrapWord
	content := container.NewVBox(promptLabel, entry)

	dlg := dialog.NewCustomConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		"Send", "Cancel",
		content,
		func(ok bool) {
			if ok {
				h.queue.Respond(req.ID, store.Response{TextValue: entry.Text})
			} else {
				h.queue.Respond(req.ID, store.Response{Cancelled: true})
			}
			w.Resize(fyne.NewSize(1, 1))
			w.Hide()
			done()
		},
		w,
	)
	dlg.Resize(fyne.NewSize(width, height))
	w.Resize(fyne.NewSize(width, height))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}

func (h *Hi) showChoiceDialog(_ context.Context, req *store.PendingRequest, done func()) {
	w := h.mainWindow

	promptLabel := widget.NewLabel(req.Prompt)
	promptLabel.Wrapping = fyne.TextWrapWord

	if req.MultiSelect {
		group := widget.NewCheckGroup(req.Choices, nil)
		content := container.NewBorder(promptLabel, nil, nil, nil, group)
		dlg := dialog.NewCustomConfirm(
			fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
			"Select", "Cancel",
			content,
			func(ok bool) {
				if ok {
					h.queue.Respond(req.ID, store.Response{ChoiceValues: group.Selected})
				} else {
					h.queue.Respond(req.ID, store.Response{Cancelled: true})
				}
				w.Resize(fyne.NewSize(1, 1))
				w.Hide()
				done()
			},
			w,
		)
		dlg.Resize(fyne.NewSize(480, 320))
		w.Resize(fyne.NewSize(480, 320))
		w.CenterOnScreen()
		w.Show()
		w.RequestFocus()
		dlg.Show()
		return
	}

	radio := widget.NewRadioGroup(req.Choices, nil)
	content := container.NewBorder(promptLabel, nil, nil, nil, radio)
	dlg := dialog.NewCustomConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		"Select", "Cancel",
		content,
		func(ok bool) {
			if ok {
				selected := []string{}
				if radio.Selected != "" {
					selected = []string{radio.Selected}
				}
				h.queue.Respond(req.ID, store.Response{ChoiceValues: selected})
			} else {
				h.queue.Respond(req.ID, store.Response{Cancelled: true})
			}
			w.Resize(fyne.NewSize(1, 1))
			w.Hide()
			done()
		},
		w,
	)
	dlg.Resize(fyne.NewSize(480, 320))
	w.Resize(fyne.NewSize(480, 320))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}

func (h *Hi) showConfirmDialog(_ context.Context, req *store.PendingRequest, done func()) {
	w := h.mainWindow
	dlg := dialog.NewConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		req.Prompt,
		func(confirmed bool) {
			h.queue.Respond(req.ID, store.Response{BoolValue: confirmed})
			w.Resize(fyne.NewSize(1, 1))
			w.Hide()
			done()
		},
		w,
	)
	w.Resize(fyne.NewSize(400, 150))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}

func (h *Hi) showSettings() {
	w := h.mainWindow
	mcpAddr := config.GetMcpAddress(h.ctx)
	addrEntry := widget.NewEntry()
	addrEntry.SetText(mcpAddr)
	addrEntry.SetPlaceHolder(":45678")

	versionLabel := widget.NewLabel("hi v" + config.GetDaemonVersion(h.ctx))

	helpText := widget.NewLabel("Changes to the MCP address require restarting the daemon.")
	helpText.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewSeparator(),
		widget.NewLabel("MCP listen address:"),
		addrEntry,
		helpText,
		widget.NewSeparator(),
		versionLabel,
	)

	dlg := dialog.NewCustomConfirm("Hi Settings", "Save", "Close", content, func(save bool) {
		if save && addrEntry.Text != "" {
			// Address changes take effect on next daemon restart.
			// TODO: persist to config file, then load alongside config in startup
			h.ctx = config.StoreMcpAddress(h.ctx, addrEntry.Text)
		}
		w.Resize(fyne.NewSize(1, 1))
		w.Hide()
	}, w)

	dlg.Resize(fyne.NewSize(450, 280))
	w.Resize(fyne.NewSize(450, 280))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}
