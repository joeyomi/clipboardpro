package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/config"
	"clipboardpro/internal/database"
)

type ItemList struct {
	repository  *database.Repository
	app         AppInterface
	container   *fyne.Container
	list        *widget.List
	items       []*database.ClipboardItem
	searchTerm  string
	statusLabel *widget.Label
}

type AppInterface interface {
	GetRepository() *database.Repository
	CopyItemToClipboard(id int64) error
	GetConfig() *config.Config
	GetWindow() fyne.Window
}

func NewItemList(repository *database.Repository, app AppInterface) *ItemList {
	itemList := &ItemList{
		repository:  repository,
		app:         app,
		items:       make([]*database.ClipboardItem, 0),
		statusLabel: widget.NewLabel("Ready"),
	}

	itemList.createList()
	return itemList
}

// Helper function to safely get the window
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
			return len(il.items)
		},
		func() fyne.CanvasObject {
			return il.createItemTemplate()
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			il.updateItem(id, item)
		},
	)

	il.list.OnSelected = func(id widget.ListItemID) {
		if id < len(il.items) {
			il.copyItem(il.items[id].ID)
		}
		il.list.UnselectAll()
	}
}

func (il *ItemList) createItemTemplate() fyne.CanvasObject {
	// Create a more visually appealing item template
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

	// Action buttons with better icons and tooltips
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

	// Info container for metadata (removed pinnedIcon)
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

	// Main container with better layout
	mainContainer := container.NewBorder(
		nil, nil,
		icon,
		actionContainer,
		textContainer,
	)

	// Add some padding
	return container.NewPadded(
		container.NewVBox(
			mainContainer,
			widget.NewSeparator(),
		),
	)
}

func (il *ItemList) updateItem(id widget.ListItemID, obj fyne.CanvasObject) {
	if id >= len(il.items) {
		return
	}

	item := il.items[id]
	paddedContainer := obj.(*fyne.Container)
	container := paddedContainer.Objects[0].(*fyne.Container)
	mainContainer := container.Objects[0].(*fyne.Container)

	// Get components
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

	// Update content
	icon.SetResource(il.getItemIcon(item.Type))
	title.SetText(il.getItemTitle(item))
	preview.SetText(il.getItemPreview(item))
	timestamp.SetText(il.formatTimeAgo(item.Timestamp))
	size.SetText(il.formatBytes(item.Size))

	// Update pin status with better icons (removed pinnedIcon.Show/Hide)
	if item.Pinned {
		pinButton.SetIcon(theme.ContentRemoveIcon()) // Minus icon for unpinning
		pinButton.SetText("Unpin")
	} else {
		pinButton.SetIcon(theme.ContentAddIcon()) // Plus icon for pinning
		pinButton.SetText("Pin")
	}

	// Set button callbacks
	pinButton.OnTapped = func() {
		il.togglePin(item.ID)
	}

	editButton.OnTapped = func() {
		il.editTitle(item)
	}

	deleteButton.OnTapped = func() {
		il.deleteItem(item.ID)
	}
}

func (il *ItemList) LoadRecentItems() {
	fyne.Do(func() {
		il.statusLabel.SetText("Loading...")
	})

	go func() {
		ctx := context.Background()
		items, err := il.repository.GetRecentItems(ctx, 100)

		fyne.Do(func() {
			if err != nil {
				il.statusLabel.SetText("Error loading items")
				window := il.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("failed to load clipboard history: %w", err), window)
				}
				return
			}

			il.items = items
			il.list.Refresh()

			// Update status
			count := len(items)
			if count == 0 {
				il.statusLabel.SetText("No items yet")
			} else if count == 1 {
				il.statusLabel.SetText("1 item")
			} else {
				il.statusLabel.SetText(fmt.Sprintf("%d items", count))
			}
		})
	}()
}

func (il *ItemList) Search(query string) {
	il.searchTerm = query // Use public field

	if query == "" {
		il.LoadRecentItems()
		return
	}

	fyne.Do(func() {
		il.statusLabel.SetText("Searching...")
	})

	go func() {
		ctx := context.Background()
		items, err := il.repository.SearchItems(ctx, query, 100)

		fyne.Do(func() {
			if err != nil {
				il.statusLabel.SetText("Search failed")
				window := il.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("Search failed: %w", err), window)
				}
				return
			}

			il.items = items
			il.list.Refresh()

			// Update status
			count := len(items)
			if count == 0 {
				il.statusLabel.SetText(fmt.Sprintf("No results for '%s'", query))
			} else if count == 1 {
				il.statusLabel.SetText("1 result")
			} else {
				il.statusLabel.SetText(fmt.Sprintf("%d results", count))
			}
		})
	}()
}

func (il *ItemList) IsSearching() bool {
	return il.searchTerm != ""
}

func (il *ItemList) Refresh() {
	if il.searchTerm == "" { // Use public field
		il.LoadRecentItems()
	} else {
		il.Search(il.searchTerm) // Use public field
	}
}

func (il *ItemList) copyItem(id int64) {
	if err := il.app.CopyItemToClipboard(id); err != nil {
		fyne.Do(func() {
			window := il.getWindow()
			if window != nil {
				dialog.ShowError(fmt.Errorf("failed to copy item: %w", err), window)
			}
		})
		return
	}

	// Show brief success message
	fyne.Do(func() {
		il.statusLabel.SetText("✓ Copied to clipboard")
	})

	// Reset status after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			il.Refresh() // This will update the status
		})
	}()
}

func (il *ItemList) togglePin(id int64) {
	go func() {
		ctx := context.Background()
		if err := il.repository.TogglePin(ctx, id); err != nil {
			fyne.Do(func() {
				window := il.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("failed to pin/unpin item: %w", err), window)
				}
			})
			return
		}

		fyne.Do(func() {
			il.Refresh()
		})
	}()
}

func (il *ItemList) deleteItem(id int64) {
	window := il.getWindow()
	if window == nil {
		return
	}

	fyne.Do(func() {
		dialog.ShowConfirm("Delete Item",
			"Are you sure you want to permanently delete this clipboard item?",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				go func() {
					ctx := context.Background()
					if err := il.repository.DeleteItem(ctx, id); err != nil {
						fyne.Do(func() {
							dialog.ShowError(fmt.Errorf("failed to delete item: %w", err), window)
						})
						return
					}

					fyne.Do(func() {
						il.statusLabel.SetText("Item deleted")
						il.Refresh()
					})
				}()
			}, window)
	})
}

func (il *ItemList) editTitle(item *database.ClipboardItem) {
	window := il.getWindow()
	if window == nil {
		return
	}

	fyne.Do(func() {
		entry := widget.NewEntry()
		entry.SetText(item.Title)
		entry.SetPlaceHolder("Enter a custom title for this item...")

		content := container.NewVBox(
			widget.NewLabel("Give this clipboard item a custom title to make it easier to find:"),
			entry,
		)

		dialog.ShowCustomConfirm("Edit Title", "Save", "Cancel", content, func(confirmed bool) {
			if !confirmed {
				return
			}

			go func() {
				ctx := context.Background()
				if err := il.repository.UpdateTitle(ctx, item.ID, entry.Text); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("failed to update title: %w", err), window)
					})
					return
				}

				fyne.Do(func() {
					il.statusLabel.SetText("Title updated")
					il.Refresh()
				})
			}()
		}, window)
	})
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
