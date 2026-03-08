package daemon

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
)

type uiState struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	trayMenu   *fyne.Menu
	queueLabel *fyne.MenuItem
}

func (d *Daemon) runUI(ctx context.Context) {
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

	d.ui = &uiState{fyneApp: a, mainWindow: w}
	d.notifyFn = func(title, body string) {
		a.SendNotification(fyne.NewNotification(title, body))
	}

	d.buildTrayMenu(ctx)

	go d.watchQueue(ctx)

	go func() {
		<-ctx.Done()
		fyne.DoAndWait(a.Quit)
	}()

	a.Run()
}

func (d *Daemon) buildTrayMenu(ctx context.Context) {
	d.ui.queueLabel = fyne.NewMenuItem("No pending requests", nil)
	d.ui.queueLabel.Disabled = true

	showItem := fyne.NewMenuItem("Show next request", func() {
		d.presentNextRequest(ctx, func() {})
	})
	settingsItem := fyne.NewMenuItem("Settings…", func() {
		d.showSettings()
	})

	quitItem := fyne.NewMenuItem("Quit hi", func() {
		d.cancel()
	})

	d.ui.trayMenu = fyne.NewMenu("Hi",
		d.ui.queueLabel,
		fyne.NewMenuItemSeparator(),
		showItem,
		settingsItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)

	if desk, ok := d.ui.fyneApp.(desktop.App); ok {
		desk.SetSystemTrayMenu(d.ui.trayMenu)
		desk.SetSystemTrayIcon(theme.NewThemedResource(assets.Tray))
	}
}

func (d *Daemon) watchQueue(ctx context.Context) {
	var dialogOpen atomic.Bool
	dialogOpen.Store(false)

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.queue.Notify():
			n := d.queue.Len()
			fyne.Do(func() { d.updateTrayLabel(n) })
			if n > 0 && !dialogOpen.Load() {
				dialogOpen.Store(true)
				fyne.Do(func() {
					d.presentNextRequest(ctx, func() { dialogOpen.Store(false) })
				})
			}
		}
	}
}

func (d *Daemon) updateTrayLabel(n int) {
	if d.ui == nil || d.ui.queueLabel == nil {
		return
	}
	switch n {
	case 0:
		d.ui.queueLabel.Label = "No pending requests"
	case 1:
		d.ui.queueLabel.Label = "1 request waiting"
	default:
		d.ui.queueLabel.Label = fmt.Sprintf("%d requests waiting", n)
	}
	d.ui.trayMenu.Refresh()
}

func (d *Daemon) presentNextRequest(ctx context.Context, done func()) {
	reqs := d.queue.Peek()
	if len(reqs) == 0 {
		done()
		return
	}
	req := reqs[0]

	switch req.Type {
	case RequestTypeText:
		d.showTextDialog(ctx, req, done)
	case RequestTypeMultiline:
		d.showMultilineDialog(ctx, req, done)
	case RequestTypeChoice:
		d.showChoiceDialog(ctx, req, done)
	case RequestTypeConfirm:
		d.showConfirmDialog(ctx, req, done)
	case RequestTypeNotify:
		d.queue.Respond(req.ID, Response{})
		done()
	}
}

func (d *Daemon) showTextDialog(_ context.Context, req *PendingRequest, done func()) {
	w := d.ui.mainWindow
	entry := widget.NewEntry()
	entry.SetText(req.DefaultVal)
	entry.SetPlaceHolder("Type your answer…")

	promptLabel := widget.NewLabel(req.Prompt)
	promptLabel.Wrapping = fyne.TextWrapWord
	content := container.NewBorder(nil, entry, nil, nil, promptLabel)

	dlg := dialog.NewCustomConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		"Send", "Cancel",
		content,
		func(ok bool) {
			if ok {
				d.queue.Respond(req.ID, Response{TextValue: entry.Text})
			} else {
				d.queue.Respond(req.ID, Response{Cancelled: true})
			}
			w.Resize(fyne.NewSize(1, 1))
			w.Hide()
			done()
		},
		w,
	)
	dlg.Resize(fyne.NewSize(520, 260))
	w.Resize(fyne.NewSize(520, 260))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}

func (d *Daemon) showMultilineDialog(_ context.Context, req *PendingRequest, done func()) {
	w := d.ui.mainWindow
	entry := widget.NewMultiLineEntry()
	entry.SetText(req.DefaultVal)
	entry.SetPlaceHolder("Type your response…")
	entry.Wrapping = fyne.TextWrapWord

	promptLabel := widget.NewLabel(req.Prompt)
	promptLabel.Wrapping = fyne.TextWrapWord
	content := container.NewBorder(promptLabel, nil, nil, nil, entry)

	dlg := dialog.NewCustomConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		"Send", "Cancel",
		content,
		func(ok bool) {
			if ok {
				d.queue.Respond(req.ID, Response{TextValue: entry.Text})
			} else {
				d.queue.Respond(req.ID, Response{Cancelled: true})
			}
			w.Resize(fyne.NewSize(1, 1))
			w.Hide()
			done()
		},
		w,
	)
	dlg.Resize(fyne.NewSize(560, 320))
	w.Resize(fyne.NewSize(560, 320))
	w.CenterOnScreen()
	w.Show()
	w.RequestFocus()
	dlg.Show()
}

func (d *Daemon) showChoiceDialog(_ context.Context, req *PendingRequest, done func()) {
	w := d.ui.mainWindow

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
					d.queue.Respond(req.ID, Response{ChoiceValues: group.Selected})
				} else {
					d.queue.Respond(req.ID, Response{Cancelled: true})
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
				d.queue.Respond(req.ID, Response{ChoiceValues: selected})
			} else {
				d.queue.Respond(req.ID, Response{Cancelled: true})
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

func (d *Daemon) showConfirmDialog(_ context.Context, req *PendingRequest, done func()) {
	w := d.ui.mainWindow
	dlg := dialog.NewConfirm(
		fmt.Sprintf("[%s] %s", req.AgentName, req.Title),
		req.Prompt,
		func(confirmed bool) {
			d.queue.Respond(req.ID, Response{BoolValue: confirmed})
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

func (d *Daemon) showSettings() {
	w := d.ui.mainWindow

	addrEntry := widget.NewEntry()
	addrEntry.SetText(d.mcpAddr)
	addrEntry.SetPlaceHolder(":45678")

	versionLabel := widget.NewLabel("hi v" + daemonVersion)

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
			d.mcpAddr = addrEntry.Text
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
