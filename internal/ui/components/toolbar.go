package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Toolbar struct {
	toolbar        *widget.Toolbar
	itemList       *ItemList
	onShowSettings func()
	onClearAll     func()
	onShowAbout    func()
	onCheckUpdates func()
}

func NewToolbar(itemList *ItemList, onShowSettings, onClearAll, onShowAbout, onCheckUpdates func()) *Toolbar {
	tb := &Toolbar{
		itemList:       itemList,
		onShowSettings: onShowSettings,
		onClearAll:     onClearAll,
		onShowAbout:    onShowAbout,
		onCheckUpdates: onCheckUpdates,
	}

	tb.createToolbar()
	return tb
}

func (tb *Toolbar) Create() fyne.CanvasObject {
	return tb.toolbar
}

func (tb *Toolbar) createToolbar() {
	tb.toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			tb.itemList.Refresh()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DownloadIcon(), tb.onCheckUpdates),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), tb.onShowSettings),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentClearIcon(), tb.onClearAll),
		widget.NewToolbarAction(theme.InfoIcon(), tb.onShowAbout),
	)
}
