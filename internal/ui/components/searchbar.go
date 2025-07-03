package components

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SearchBar struct {
	itemList    *ItemList
	entry       *widget.Entry
	clearButton *widget.Button
	container   *fyne.Container
	searchTimer *time.Timer
}

func NewSearchBar(itemList *ItemList) *SearchBar {
	sb := &SearchBar{
		itemList:    itemList,
		searchTimer: nil,
	}

	sb.createSearchBar()
	return sb
}

func (sb *SearchBar) Create() fyne.CanvasObject {
	if sb.container == nil {
		searchIcon := widget.NewIcon(theme.SearchIcon())

		sb.container = container.NewBorder(
			nil, nil,
			searchIcon,
			sb.clearButton,
			sb.entry,
		)
	}
	return sb.container
}

func (sb *SearchBar) createSearchBar() {
	sb.entry = widget.NewEntry()
	sb.entry.SetPlaceHolder("Search clipboard history...")

	sb.clearButton = widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		sb.Clear()
	})
	sb.clearButton.Hide() // Initially hidden

	sb.entry.OnChanged = func(text string) {
		// Show/hide clear button
		if text == "" {
			sb.clearButton.Hide()
		} else {
			sb.clearButton.Show()
		}

		// Debounce search to avoid too many queries
		if sb.searchTimer != nil {
			sb.searchTimer.Stop()
		}

		sb.searchTimer = time.AfterFunc(300*time.Millisecond, func() {
			sb.itemList.Search(text)
		})
	}

	sb.entry.OnSubmitted = func(text string) {
		sb.itemList.Search(text)
	}
}

func (sb *SearchBar) Focus() {
	if sb.itemList != nil && sb.itemList.app != nil && sb.itemList.app.GetWindow() != nil {
		sb.itemList.app.GetWindow().Canvas().Focus(sb.entry)
	}
}

func (sb *SearchBar) Clear() {
	sb.entry.SetText("")
	sb.clearButton.Hide()
}

func (sb *SearchBar) GetText() string {
	return sb.entry.Text
}
