package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/creativeprojects/go-selfupdate"
)

type UpdateChecker struct {
	app    *ClipboardProApp
	source selfupdate.Source
}

func NewUpdateChecker(app *ClipboardProApp) *UpdateChecker {
	// Use GitHub releases as source
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		log.Printf("Failed to create update source: %v", err)
		return nil
	}

	return &UpdateChecker{
		app:    app,
		source: source,
	}
}

func (uc *UpdateChecker) CheckForUpdates(ctx context.Context, showNoUpdateDialog bool) {
	if uc == nil {
		return
	}

	go func() {
		hasUpdate, release, err := uc.checkForUpdatesAsync(ctx)
		if err != nil {
			log.Printf("Update check failed: %v", err)
			if showNoUpdateDialog {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to check for updates: %v", err), uc.app.window)
				})
			}
			return
		}

		if hasUpdate {
			fyne.Do(func() {
				uc.showUpdateDialog(release)
			})
		} else if showNoUpdateDialog {
			fyne.Do(func() {
				dialog.ShowInformation("No Updates", "You're running the latest version!", uc.app.window)
			})
		}
	}()
}

func (uc *UpdateChecker) checkForUpdatesAsync(ctx context.Context) (bool, *selfupdate.Release, error) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: uc.source,
		Validator: &selfupdate.ChecksumValidator{
			UniqueFilename: "checksums.txt",
		},
	})
	if err != nil {
		return false, nil, err
	}

	release, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug("joeyomi", "clipboardpro")) // TBD: Change this
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, fmt.Errorf("no releases found")
	}

	return release.GreaterThan(Version), release, nil
}

func (uc *UpdateChecker) showUpdateDialog(release *selfupdate.Release) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Update Available!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Version %s is available", release.Version())),
		widget.NewLabel(fmt.Sprintf("Current version: %s", Version)),
		widget.NewLabel(""),
		widget.NewLabel("Would you like to download and install the update?"),
		widget.NewLabel(""),
		widget.NewLabelWithStyle("Note: The application will restart after the update.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
	)

	dialog.ShowCustomConfirm("Update Available", "Update Now", "Later", content,
		func(update bool) {
			if update {
				uc.performUpdate(release)
			}
		}, uc.app.window)
}

func (uc *UpdateChecker) performUpdate(release *selfupdate.Release) {
	progress := dialog.NewProgressInfinite("Updating", "Downloading update...", uc.app.window)
	progress.Show()

	go func() {
		defer progress.Hide()

		if err := uc.doUpdate(release); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("update failed: %v", err), uc.app.window)
			})
			return
		}

		fyne.Do(func() {
			dialog.ShowInformation("Update Complete",
				"ClipBoard Pro has been updated successfully!\n\nPlease restart the application to use the new version.",
				uc.app.window)
		})
	}()
}

func (uc *UpdateChecker) doUpdate(release *selfupdate.Release) error {
	// Get the correct executable path
	exe, err := uc.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Check permissions
	if err := uc.checkWritePermissions(exe); err != nil {
		return fmt.Errorf("insufficient permissions to update. Please run as administrator or move the app to a user-writable location: %w", err)
	}

	// Create updater
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: uc.source,
		Validator: &selfupdate.ChecksumValidator{
			UniqueFilename: "checksums.txt",
		},
	})
	if err != nil {
		return err
	}

	// Perform update
	return updater.UpdateTo(context.Background(), release, exe)
}

func (uc *UpdateChecker) getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	// On macOS, if we're inside an app bundle, update the actual binary
	if runtime.GOOS == "darwin" {
		if filepath.Base(filepath.Dir(exe)) == "MacOS" {
			// We're inside Contents/MacOS/ - this is correct for app bundles
			return exe, nil
		}
	}

	return exe, nil
}

func (uc *UpdateChecker) checkWritePermissions(path string) error {
	// Try to open the file for writing
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
