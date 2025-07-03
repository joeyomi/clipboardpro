package components

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/config"
)

type SettingsController struct {
	config *config.Config
	parent fyne.Window
	onSave func(*config.Config)
}

func NewSettingsController(cfg *config.Config, parent fyne.Window, onSave func(*config.Config)) *SettingsController {
	return &SettingsController{
		config: cfg,
		parent: parent,
		onSave: onSave,
	}
}

func (sc *SettingsController) SaveSettings(maxItemsEntry, maxDaysEntry *widget.Entry, darkModeCheck, checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck *widget.Check) {
	// Validate inputs
	maxItems, err := strconv.Atoi(maxItemsEntry.Text)
	if err != nil {
		dialog.ShowError(err, sc.parent)
		return
	}

	maxDays, err := strconv.Atoi(maxDaysEntry.Text)
	if err != nil {
		dialog.ShowError(err, sc.parent)
		return
	}

	// Create new config
	newConfig := &config.Config{}
	*newConfig = *sc.config

	newConfig.MaxHistoryItems = maxItems
	newConfig.MaxHistoryDays = maxDays
	newConfig.DarkMode = darkModeCheck.Checked
	newConfig.CheckUpdatesOnStartup = checkUpdatesOnStartupCheck.Checked
	newConfig.AutoDownloadUpdates = autoDownloadUpdatesCheck.Checked

	sc.onSave(newConfig)
}

func (sc *SettingsController) ResetSettings() {
	dialog.ShowConfirm("Reset Settings",
		"Are you sure you want to reset all settings to their default values?",
		func(confirmed bool) {
			if confirmed {
				defaultConfig := config.Default()
				sc.onSave(defaultConfig)
			}
		}, sc.parent)
}
