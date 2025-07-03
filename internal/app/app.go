package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/clipboard"
	"clipboardpro/internal/config"
	"clipboardpro/internal/database"
	"clipboardpro/internal/ui/components"
)

// Build-time variables (set by GoReleaser)
var (
	Version   = "0.0.0-dev" // Will be replaced by -ldflags
	BuildDate = "unknown"   // Will be replaced by -ldflags
	GitCommit = "unknown"   // Will be replaced by -ldflags
)

const (
	AppName = "ClipBoard Pro"
	AppID   = "com.clipboardpro.app"

	// Update configuration
	UpdateCheckInterval = 24 * time.Hour // Check daily
)

type ClipboardProApp struct {
	fyneApp    fyne.App
	window     fyne.Window
	config     *config.Config
	repository *database.Repository
	monitor    *clipboard.Monitor

	// UI Components
	itemList  *components.ItemList
	searchBar *components.SearchBar
	toolbar   *components.Toolbar
	statusBar *widget.Label

	// Update functionality
	updateChecker *UpdateChecker

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewClipboardProApp() (*ClipboardProApp, error) {
	fyneApp := app.NewWithID(AppID)

	// Set app metadata using build-time version
	app.SetMetadata(fyne.AppMetadata{
		ID:      AppID,
		Name:    AppName,
		Version: Version,
		Build:   1,
	})

	ctx, cancel := context.WithCancel(context.Background())

	clipboardApp := &ClipboardProApp{
		fyneApp:    fyneApp,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	if err := clipboardApp.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	return clipboardApp, nil
}

func (a *ClipboardProApp) initialize() error {
	if err := a.initConfig(); err != nil {
		return err
	}
	if err := a.initDatabase(); err != nil {
		return err
	}
	a.initServices()
	a.initUIComponents()

	// Initialize update checker
	a.updateChecker = NewUpdateChecker(a)

	// Create main window
	a.createMainWindow()

	return nil
}

func (a *ClipboardProApp) initConfig() error {
	configDir, err := a.getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	a.config, err = config.Load(configPath)
	if err != nil {
		log.Printf("Creating default configuration: %v", err)
		a.config = config.Default()
		if err := a.config.Save(configPath); err != nil {
			log.Printf("Failed to save default config: %v", err)
		}
	}
	return nil
}

func (a *ClipboardProApp) initDatabase() error {
	configDir, err := a.getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	a.repository, err = database.NewRepository(filepath.Join(configDir, "clipboard.db"))
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

func (a *ClipboardProApp) initServices() {
	a.monitor = clipboard.NewMonitor(a.repository, a.config)
}

func (a *ClipboardProApp) initUIComponents() {
	a.itemList = components.NewItemList(a.repository, a)
	a.searchBar = components.NewSearchBar(a.itemList)
	a.toolbar = components.NewToolbar(a.itemList, a.showSettings, a.clearAll, a.showAbout, a.checkForUpdates)
	a.statusBar = widget.NewLabel("Starting ClipBoard Pro...")
}

func (a *ClipboardProApp) createMainWindow() {
	a.window = a.fyneApp.NewWindow(AppName)
	a.window.SetMaster()
	a.window.Resize(fyne.NewSize(900, 700))
	a.window.CenterOnScreen()

	// Create main content
	content := a.createMainContent()
	a.window.SetContent(content)

	// Set window callbacks
	a.window.SetCloseIntercept(func() {
		a.cleanup()
		a.fyneApp.Quit()
	})

	// Show welcome message for first-time users
	a.showWelcomeIfFirstRun()
}

func (a *ClipboardProApp) createMainContent() fyne.CanvasObject {
	// Create a welcome panel for when there are no items
	welcomeContent := container.NewVBox(
		widget.NewIcon(theme.InfoIcon()),
		widget.NewLabelWithStyle("Welcome to ClipBoard Pro!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Your clipboard history will appear here automatically."),
		widget.NewLabel("Copy some text or images to get started!"),
	)
	welcomeContent.Hide() // Initially hidden

	// Create main content area
	contentArea := container.NewMax(
		a.itemList.Create(),
		welcomeContent,
	)

	// Create main layout with better spacing
	mainContainer := container.NewBorder(
		container.NewVBox(
			a.toolbar.Create(),
			widget.NewSeparator(),
			a.searchBar.Create(),
			widget.NewSeparator(),
		),
		container.NewBorder(
			widget.NewSeparator(),
			nil, nil, nil,
			container.NewPadded(a.statusBar),
		),
		nil, nil,
		contentArea,
	)

	return mainContainer
}

func (a *ClipboardProApp) showWelcomeIfFirstRun() {
	// Check if this is the first run
	configDir, _ := a.getConfigDir()
	firstRunFile := filepath.Join(configDir, ".first_run")

	if _, err := os.Stat(firstRunFile); os.IsNotExist(err) {
		// Create the first run marker
		os.WriteFile(firstRunFile, []byte(""), 0644)

		// Show welcome dialog
		welcomeText := `Welcome to ClipBoard Pro!

This application will automatically save everything you copy to your clipboard, making it easy to find and reuse later.

Key features:
- Automatic clipboard monitoring
- Search through your history
- Pin important items
- Organize with custom titles
- Automatic updates

ClipBoard Pro runs in the background and can be accessed from the system tray.`

		content := container.NewVBox(
			widget.NewLabelWithStyle("🎉 Welcome!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(welcomeText),
		)

		dialog.ShowCustom("Welcome to ClipBoard Pro", "Get Started", content, a.window)
	}
}

func (a *ClipboardProApp) ShowAndRun() {
	defer a.repository.Close()

	// Start background services
	go func() {
		if err := a.monitor.Start(a.ctx); err != nil {
			log.Printf("Failed to start clipboard monitor: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to start clipboard monitoring.\n\nClipBoard Pro needs access to your clipboard to work properly.\n\nError: %v", err), a.window)
			})
		} else {
			fyne.Do(func() {
				a.statusBar.SetText("✓ Clipboard monitoring active")
			})
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				if a.itemList != nil {
					fyne.Do(func() {
						// Only refresh if showing recent items (not search results)
						if !a.itemList.IsSearching() {
							a.itemList.LoadRecentItems()
						}
					})
				}
			}
		}
	}()

	go a.startCleanupRoutine()

	// Check for updates on startup (after a delay)
	if a.config.CheckUpdatesOnStartup {
		go func() {
			time.Sleep(5 * time.Second) // Wait for app to fully load
			if a.updateChecker != nil {
				a.updateChecker.CheckForUpdates(a.ctx, false) // Don't show "no updates" dialog
			}
		}()
	}

	// Load initial data
	a.itemList.LoadRecentItems()

	log.Printf("%s %s started", AppName, Version)

	// Update status
	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			a.statusBar.SetText("Ready • Monitoring clipboard")
		})
	}()

	// Show window and run app
	a.window.Show()
	a.fyneApp.Run()

	// Cleanup
	a.cleanup()
}

func (a *ClipboardProApp) cleanup() {
	log.Println("Shutting down ClipBoard Pro...")
	fyne.Do(func() {
		a.statusBar.SetText("Shutting down...")
	})

	a.cancelFunc()
	a.monitor.Stop()
	if a.repository != nil {
		a.repository.Close()
	}
	log.Println("ClipBoard Pro shutdown complete")
}

func (a *ClipboardProApp) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.repository.CleanupOldItems(a.ctx, a.config.MaxHistoryDays, a.config.MaxHistoryItems); err != nil {
				log.Printf("Cleanup failed: %v", err)
			}
		}
	}
}

func (a *ClipboardProApp) checkForUpdates() {
	if a.updateChecker != nil {
		a.updateChecker.CheckForUpdates(a.ctx, true) // Show "no updates" dialog
	}
}

func (a *ClipboardProApp) showSettings() {
	if a.window == nil {
		log.Printf("Warning: Window is nil, cannot show settings")
		return
	}

	fyne.Do(func() {
		settingsDialog := components.NewSettingsDialog(a.config, a.window)
		settingsDialog.Show(func(newConfig *config.Config) {
			a.config = newConfig
			configDir, _ := a.getConfigDir()
			if err := a.config.Save(filepath.Join(configDir, "config.json")); err != nil {
				log.Printf("Failed to save config: %v", err)
			}
			a.itemList.Refresh()
			fyne.Do(func() {
				a.statusBar.SetText("Settings saved")
			})
		})
	})
}

func (a *ClipboardProApp) clearAll() {
	if a.window == nil {
		log.Printf("Warning: Window is nil, cannot clear all")
		return
	}

	fyne.Do(func() {
		dialog.ShowConfirm("Clear All", "Are you sure you want to clear all clipboard history? This action cannot be undone.",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				go func() {
					ctx := context.Background()
					if err := a.repository.ClearAllItems(ctx); err != nil {
						fyne.Do(func() {
							dialog.ShowError(fmt.Errorf("failed to clear all items: %w", err), a.window)
						})
						return
					}
					fyne.Do(func() {
						a.itemList.Refresh()
						a.statusBar.SetText("All clipboard history cleared")
					})
				}()
			}, a.window)
	})
}

func (a *ClipboardProApp) showAbout() {
	if a.window == nil {
		log.Printf("Warning: Window is nil, cannot show about")
		return
	}

	fyne.Do(func() {
		content := container.NewVBox(
			widget.NewLabelWithStyle("ClipBoard Pro", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle(fmt.Sprintf("Version %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle(fmt.Sprintf("Built: %s", BuildDate), fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle(fmt.Sprintf("Commit: %s", GitCommit), fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabel(""),
			widget.NewLabel("Advanced clipboard manager for desktop"),
			widget.NewLabel(""),
			widget.NewLabel("Copyright © 2025 ClipBoard Pro Team"),
			widget.NewLabel("All rights reserved."),
		)

		dialog.ShowCustom("About ClipBoard Pro", "Close", content, a.window)
	})
}

func (a *ClipboardProApp) getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".clipboardpro")
	return configDir, os.MkdirAll(configDir, 0755)
}

func (a *ClipboardProApp) GetRepository() *database.Repository {
	return a.repository
}

func (a *ClipboardProApp) GetConfig() *config.Config {
	return a.config
}

func (a *ClipboardProApp) GetWindow() fyne.Window {
	return a.window
}

func (a *ClipboardProApp) CopyItemToClipboard(id int64) error {
	return a.monitor.CopyItemToClipboard(a.ctx, id)
}
