package components

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"clipboardpro/internal/config"
)

type SettingsDialog struct {
	controller *SettingsController
	config     *config.Config
	parent     fyne.Window
}

func NewSettingsDialog(cfg *config.Config, parent fyne.Window, onSave func(*config.Config)) *SettingsDialog {
	sd := &SettingsDialog{
		config: cfg,
		parent: parent,
	}
	sd.controller = NewSettingsController(cfg, parent, onSave)
	return sd
}

func (sd *SettingsDialog) Show() {
	content := sd.createContent()

	dialog.ShowCustom("Settings", "Close", content, sd.parent)
}

func (sd *SettingsDialog) createContent() fyne.CanvasObject {
	maxItemsEntry := sd.createNumericEntry(strconv.Itoa(sd.config.MaxHistoryItems))
	maxDaysEntry := sd.createNumericEntry(strconv.Itoa(sd.config.MaxHistoryDays))

	darkModeCheck := sd.createCheckbox("Use dark theme", sd.config.DarkMode)

	// Update settings
	checkUpdatesOnStartupCheck := sd.createCheckbox("Check for updates on startup", sd.config.CheckUpdatesOnStartup)
	autoDownloadUpdatesCheck := sd.createCheckbox("Automatically download updates", sd.config.AutoDownloadUpdates)

	tabs := container.NewAppTabs(
		sd.createStorageTab(maxItemsEntry, maxDaysEntry),
		sd.createAppearanceTab(darkModeCheck),
		sd.createUpdatesTab(checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck),
	)

	saveButton := sd.createSaveButton(maxItemsEntry, maxDaysEntry, darkModeCheck, checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck)
	resetButton := sd.createResetButton()

	buttonContainer := container.NewHBox(
		resetButton,
		layout.NewSpacer(), // Spacer
		saveButton,
	)

	mainContent := container.NewVBox(
		tabs,
		widget.NewSeparator(),
		buttonContainer,
	)

	// Set a reasonable size
	mainContent.Resize(fyne.NewSize(500, 400))

	return mainContent
}

func (sd *SettingsDialog) createNumericEntry(initialValue string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(initialValue)
	entry.Validator = func(text string) error {
		if _, err := strconv.Atoi(text); err != nil {
			return fmt.Errorf("must be a number")
		}
		return nil
	}
	return entry
}

func (sd *SettingsDialog) createCheckbox(text string, checked bool) *widget.Check {
	check := widget.NewCheck(text, nil)
	check.SetChecked(checked)
	return check
}

func (sd *SettingsDialog) createStorageTab(maxItemsEntry, maxDaysEntry *widget.Entry) *container.TabItem {
	storageForm := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Maximum items to keep", maxItemsEntry),
			widget.NewFormItem("Delete items older than (days)", maxDaysEntry),
		},
	}
	return container.NewTabItem("Storage", container.NewVBox(
		widget.NewLabelWithStyle("Clipboard History Storage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		storageForm,
	))
}

func (sd *SettingsDialog) createAppearanceTab(darkModeCheck *widget.Check) *container.TabItem {
	appearanceForm := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("", darkModeCheck),
		},
	}
	return container.NewTabItem("Appearance", container.NewVBox(
		widget.NewLabelWithStyle("User Interface", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		appearanceForm,
		widget.NewLabel("Note: Dark theme functionality is not yet implemented."),
	))
}

func (sd *SettingsDialog) createUpdatesTab(checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck *widget.Check) *container.TabItem {
	updatesForm := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("", checkUpdatesOnStartupCheck),
			widget.NewFormItem("", autoDownloadUpdatesCheck),
		},
	}

	infoText := widget.NewLabel("ClipBoard Pro can automatically check for and install updates to keep you secure and up-to-date with the latest features.")
	infoText.Wrapping = fyne.TextWrapWord

	return container.NewTabItem("Updates", container.NewVBox(
		widget.NewLabelWithStyle("Automatic Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		infoText,
		widget.NewLabel(""),
		updatesForm,
	))
}

func (sd *SettingsDialog) createSaveButton(maxItemsEntry, maxDaysEntry *widget.Entry, darkModeCheck, checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck *widget.Check) *widget.Button {
	saveButton := widget.NewButton("Save Settings", func() {
		sd.controller.SaveSettings(maxItemsEntry, maxDaysEntry, darkModeCheck, checkUpdatesOnStartupCheck, autoDownloadUpdatesCheck)
	})
	saveButton.Importance = widget.HighImportance
	return saveButton
}

func (sd *SettingsDialog) createResetButton() *widget.Button {
	resetButton := widget.NewButton("Reset to Defaults", func() {
		sd.controller.ResetSettings()
	})
	resetButton.Importance = widget.LowImportance
	return resetButton
}
