package components

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/config"
	"clipboardpro/internal/database"
)

type ItemList struct {
	controller  *ItemListController
	app         AppInterface
	container   *fyne.Container
	list        *widget.List
	statusLabel *widget.Label
}

type AppInterface interface {
	GetRepository() *database.Repository
	CopyItemToClipboard(id int64) error
	GetConfig() *config.Config
	GetWindow() fyne.Window
}

func NewItemList(repository *database.Repository, app AppInterface) *ItemList {
	statusLabel := widget.NewLabel("Ready")
	itemList := &ItemList{
		app:         app,
		statusLabel: statusLabel,
	}

	itemList.controller = NewItemListController(
		repository,
		app,
		statusLabel,
		itemList.listRefresh,
		itemList.getWindow,
	)

	itemList.createList()
	return itemList
}

func (il *ItemList) getWindow() fyne.Window {
	if il.app != nil {
		return il.app.GetWindow()
	}
	return nil
}

func (il *ItemList) Create() fyne.CanvasObject {
	if il.container == nil {
		// Create header with count and status
		header := container.NewBorder(
			nil, nil,
			widget.NewLabelWithStyle("Clipboard History", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			il.statusLabel,
		)

		// Create list container with header
		il.container = container.NewBorder(
			header,
			nil, nil, nil,
			il.list,
		)
	}
	return il.container
}

func (il *ItemList) createList() {
	il.list = widget.NewList(
		func() int {
			return len(il.controller.GetItems())
		},
		func() fyne.CanvasObject {
			return il.createItemTemplate()
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			il.updateItem(id, item)
		},
	)

	il.list.OnSelected = func(id widget.ListItemID) {
		if id < len(il.controller.GetItems()) {
			il.controller.CopyItem(il.controller.GetItems()[id].ID)
		}
		il.list.UnselectAll()
	}
}

func (il *ItemList) listRefresh() {
	il.list.Refresh()
}

func (il *ItemList) createItemTemplate() fyne.CanvasObject {
	icon := widget.NewIcon(theme.DocumentIcon())
	icon.Resize(fyne.NewSize(32, 32))

	title := widget.NewLabel("")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Truncation = fyne.TextTruncateEllipsis

	preview := widget.NewLabel("")
	preview.Wrapping = fyne.TextWrapWord

	timestamp := widget.NewLabel("")
	timestamp.TextStyle = fyne.TextStyle{Italic: true}

	size := widget.NewLabel("")
	size.TextStyle = fyne.TextStyle{Monospace: true}

	pinButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), nil)
	pinButton.Importance = widget.LowImportance

	editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editButton.Importance = widget.LowImportance

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteButton.Importance = widget.LowImportance

	actionContainer := container.NewHBox(
		pinButton,
		editButton,
		deleteButton,
	)

	infoContainer := container.NewHBox(
		timestamp,
		widget.NewSeparator(),
		size,
		layout.NewSpacer(),
	)

	textContainer := container.NewVBox(
		title,
		preview,
		infoContainer,
	)

	mainContainer := container.NewBorder(
		nil, nil,
		icon,
		actionContainer,
		textContainer,
	)

	return container.NewPadded(
		container.NewVBox(
			mainContainer,
			widget.NewSeparator(),
		),
	)
}

func (il *ItemList) updateItem(id widget.ListItemID, obj fyne.CanvasObject) {
	items := il.controller.GetItems()
	if id >= len(items) {
		return
	}

	item := items[id]
	paddedContainer := obj.(*fyne.Container)
	container := paddedContainer.Objects[0].(*fyne.Container)
	mainContainer := container.Objects[0].(*fyne.Container)

	icon := mainContainer.Objects[1].(*widget.Icon)
	textContainer := mainContainer.Objects[0].(*fyne.Container)
	actionContainer := mainContainer.Objects[2].(*fyne.Container)

	title := textContainer.Objects[0].(*widget.Label)
	preview := textContainer.Objects[1].(*widget.Label)
	infoContainer := textContainer.Objects[2].(*fyne.Container)

	timestamp := infoContainer.Objects[0].(*widget.Label)
	size := infoContainer.Objects[2].(*widget.Label)

	pinButton := actionContainer.Objects[0].(*widget.Button)
	editButton := actionContainer.Objects[1].(*widget.Button)
	deleteButton := actionContainer.Objects[2].(*widget.Button)

	icon.SetResource(il.getItemIcon(item.Type))
	title.SetText(il.getItemTitle(item))
	preview.SetText(il.getItemPreview(item))
	timestamp.SetText(il.formatTimeAgo(item.Timestamp))
	size.SetText(il.formatBytes(item.Size))

	if item.Pinned {
		pinButton.SetIcon(theme.ContentRemoveIcon()) // Minus icon for unpinning
		pinButton.SetText("Unpin")
	} else {
		pinButton.SetIcon(theme.ContentAddIcon()) // Plus icon for pinning
		pinButton.SetText("Pin")
	}

	pinButton.OnTapped = func() {
		il.controller.TogglePin(item.ID)
	}

	editButton.OnTapped = func() {
		il.controller.EditTitle(item)
	}

	deleteButton.OnTapped = func() {
		il.controller.DeleteItem(item.ID)
	}
}

func (il *ItemList) LoadRecentItems() {
	il.controller.LoadRecentItems()
}

func (il *ItemList) Search(query string) {
	il.controller.Search(query)
}

func (il *ItemList) IsSearching() bool {
	return il.controller.IsSearching()
}

func (il *ItemList) Refresh() {
	il.controller.Refresh()
}



func (il *ItemList) getItemIcon(itemType string) fyne.Resource {
	switch itemType {
	case "text":
		return theme.DocumentIcon()
	case "image":
		return theme.FileImageIcon()
	default:
		return theme.FileIcon()
	}
}

func (il *ItemList) getItemTitle(item *database.ClipboardItem) string {
	if item.Title != "" {
		return item.Title
	}

	switch item.Type {
	case "text":
		content := strings.ReplaceAll(item.Content, "\n", " ")
		content = strings.TrimSpace(content)
		if len(content) > 60 {
			return content[:60] + "..."
		}
		if content == "" {
			return "Empty text"
		}
		return content
	case "image":
		return "Image from clipboard"
	default:
		return "Clipboard item"
	}
}

func (il *ItemList) getItemPreview(item *database.ClipboardItem) string {
	switch item.Type {
	case "text":
		content := strings.ReplaceAll(item.Content, "\n", " ")
		content = strings.TrimSpace(content)
		if len(content) > 120 {
			content = content[:120] + "..."
		}
		if content == "" {
			return "This item contains no visible text"
		}
		return content
	case "image":
		return fmt.Sprintf("PNG image • %s", il.formatBytes(item.Size))
	default:
		return fmt.Sprintf("Clipboard data • %s", il.formatBytes(item.Size))
	}
}

func (il *ItemList) formatTimeAgo(timestamp time.Time) string {
	now := time.Now()
	diff := now.Sub(timestamp)

	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	return timestamp.Format("Jan 2, 2006")
}

func (il *ItemList) formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}
