package components

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/database"
)

type ItemListController struct {
	repository  *database.Repository
	app         AppInterface
	statusLabel *widget.Label // Reference to the UI status label
	items       []*database.ClipboardItem
	searchTerm  string
	listRefresh func() // Callback to refresh the UI list
	getWindow   func() fyne.Window // Callback to get the main window
}

func NewItemListController(repository *database.Repository, app AppInterface, statusLabel *widget.Label, listRefresh func(), getWindow func() fyne.Window) *ItemListController {
	return &ItemListController{
		repository:  repository,
		app:         app,
		statusLabel: statusLabel,
		listRefresh: listRefresh,
		getWindow:   getWindow,
	}
}

func (ilc *ItemListController) GetItems() []*database.ClipboardItem {
	return ilc.items
}

// LoadRecentItems loads the most recent clipboard items from the database.
func (ilc *ItemListController) LoadRecentItems() {
	fyne.Do(func() {
		ilc.statusLabel.SetText("Loading...")
	})

	go func() {
		ctx := context.Background()
		items, err := ilc.repository.GetRecentItems(ctx, 100)

		fyne.Do(func() {
			if err != nil {
				ilc.statusLabel.SetText("Error loading items")
				window := ilc.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("failed to load clipboard history: %w", err), window)
				}
				return
			}

			ilc.items = items
			ilc.listRefresh()

			// Update status
			count := len(items)
			if count == 0 {
				ilc.statusLabel.SetText("No items yet")
			} else if count == 1 {
				ilc.statusLabel.SetText("1 item")
			} else {
				ilc.statusLabel.SetText(fmt.Sprintf("%d items", count))
			}
		})
	}()
}

// Search searches for clipboard items based on a query.
func (ilc *ItemListController) Search(query string) {
	ilc.searchTerm = query

	if query == "" {
		ilc.LoadRecentItems()
		return
	}

	fyne.Do(func() {
		ilc.statusLabel.SetText("Searching...")
	})

	go func() {
		ctx := context.Background()
		items, err := ilc.repository.SearchItems(ctx, query, 100)

		fyne.Do(func() {
			if err != nil {
				ilc.statusLabel.SetText("Search failed")
				window := ilc.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("Search failed: %w", err), window)
				}
				return
			}

			ilc.items = items
			ilc.listRefresh()

			// Update status
			count := len(items)
			if count == 0 {
				ilc.statusLabel.SetText(fmt.Sprintf("No results for '%s'", query))
			} else if count == 1 {
				ilc.statusLabel.SetText("1 result")
			} else {
				ilc.statusLabel.SetText(fmt.Sprintf("%d results", count))
			}
		})
	}()
}

// IsSearching returns true if a search query is active.
func (ilc *ItemListController) IsSearching() bool {
	return ilc.searchTerm != ""
}

// Refresh refreshes the item list based on the current search term.
func (ilc *ItemListController) Refresh() {
	if ilc.searchTerm == "" {
		ilc.LoadRecentItems()
	} else {
		ilc.Search(ilc.searchTerm)
	}
}

// CopyItem copies the item with the given ID to the clipboard.
func (ilc *ItemListController) CopyItem(id int64) {
	if err := ilc.app.CopyItemToClipboard(id); err != nil {
		fyne.Do(func() {
			window := ilc.getWindow()
			if window != nil {
				dialog.ShowError(fmt.Errorf("failed to copy item: %w", err), window)
			}
		})
		return
	}

	fyne.Do(func() {
		ilc.statusLabel.SetText("✓ Copied to clipboard")
	})

	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			ilc.Refresh()
		})
	}()
}

// TogglePin toggles the pinned status of an item.
func (ilc *ItemListController) TogglePin(id int64) {
	go func() {
		ctx := context.Background()
		if err := ilc.repository.TogglePin(ctx, id); err != nil {
			fyne.Do(func() {
				window := ilc.getWindow()
				if window != nil {
					dialog.ShowError(fmt.Errorf("failed to pin/unpin item: %w", err), window)
				}
			})
			return
		}

		fyne.Do(func() {
			ilc.Refresh()
		})
	}()
}

// DeleteItem deletes an item after user confirmation.
func (ilc *ItemListController) DeleteItem(id int64) {
	window := ilc.getWindow()
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
					if err := ilc.repository.DeleteItem(ctx, id); err != nil {
						fyne.Do(func() {
							dialog.ShowError(fmt.Errorf("failed to delete item: %w", err), window)
						})
						return
					}

					fyne.Do(func() {
						ilc.statusLabel.SetText("Item deleted")
						ilc.Refresh()
					})
				}()
			}, window)
	})
}

// EditTitle allows editing the title of an item.
func (ilc *ItemListController) EditTitle(item *database.ClipboardItem) {
	window := ilc.getWindow()
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
				if err := ilc.repository.UpdateTitle(ctx, item.ID, entry.Text); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("failed to update title: %w", err), window)
					})
					return
				}

				fyne.Do(func() {
					ilc.statusLabel.SetText("Title updated")
					ilc.Refresh()
				})
			}()
		}, window)
	})
}
