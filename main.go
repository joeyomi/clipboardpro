package main

import (
	"log"
	"os"

	"clipboardpro/internal/app"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	clipboardApp, err := app.NewClipboardProApp()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
		os.Exit(1)
	}

	clipboardApp.ShowAndRun()
}
