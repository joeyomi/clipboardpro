package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
)

type UpdateChecker struct {
	app    *ClipboardProApp
	source selfupdate.Source
}

func NewUpdateChecker(app *ClipboardProApp) *UpdateChecker {
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

	release, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug("joeyomi", "clipboardpro"))
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, fmt.Errorf("no releases found")
	}

	currentVersion := uc.normalizeVersion(uc.app.GetVersion())

	hasUpdate, err := uc.safeVersionComparison(release, currentVersion)
	if err != nil {
		log.Printf("Version comparison failed: %v", err)
		return true, release, nil
	}

	return hasUpdate, release, nil
}

func (uc *UpdateChecker) normalizeVersion(version string) string {
	switch version {
	case "dev", "development", "", "0.0.0-dev":
		return "0.0.0-dev"
	default:
		if len(version) > 0 && version[0] == 'v' {
			return version[1:]
		}
		return version
	}
}

func (uc *UpdateChecker) safeVersionComparison(release *selfupdate.Release, currentVersion string) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic during version comparison: %v (current: %s, release: %s)",
				r, currentVersion, release.Version())
		}
	}()

	if currentVersion == "0.0.0-dev" {
		return true, nil
	}

	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		log.Printf("Failed to parse current version '%s': %v", currentVersion, err)
		return true, nil
	}

	releaseVersionStr := strings.TrimPrefix(release.Version(), "v")

	releaseVersion, err := semver.NewVersion(releaseVersionStr)
	if err != nil {
		log.Printf("Failed to parse release version '%s': %v", releaseVersionStr, err)
		return false, fmt.Errorf("invalid release version format: %w", err)
	}

	return releaseVersion.GreaterThan(current), nil
}

func (uc *UpdateChecker) showUpdateDialog(release *selfupdate.Release) {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Update Available!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Version %s is available", release.Version())),
		widget.NewLabel(fmt.Sprintf("Current version: %s", uc.app.GetVersion())),
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
	exe, err := uc.getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := uc.checkWritePermissions(exe); err != nil {
		return fmt.Errorf("insufficient permissions to update. Please run as administrator or move the app to a user-writable location: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: uc.source,
		Validator: &selfupdate.ChecksumValidator{
			UniqueFilename: "checksums.txt",
		},
	})
	if err != nil {
		return err
	}

	return updater.UpdateTo(context.Background(), release, exe)
}

func (uc *UpdateChecker) getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		if filepath.Base(filepath.Dir(exe)) == "MacOS" {
			return exe, nil
		}
	}

	return exe, nil
}

func (uc *UpdateChecker) checkWritePermissions(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
