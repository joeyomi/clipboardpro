package clipboard

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.design/x/clipboard"

	"clipboardpro/internal/config"
	"clipboardpro/internal/database"
	"clipboardpro/internal/util"
)

type Monitor struct {
	repository *database.Repository
	config     *config.Config
	lastHash   string
	eventChan  chan MonitorEvent
	isRunning  bool
}

func NewMonitor(repository *database.Repository, config *config.Config) *Monitor {
	return &Monitor{
		repository: repository,
		config:     config,
		eventChan:  make(chan MonitorEvent, 100),
	}
}

func (m *Monitor) Start(ctx context.Context) error {
	if m.isRunning {
		return fmt.Errorf("monitor is already running")
	}

	// Initialize clipboard
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	m.isRunning = true
	log.Println("Clipboard monitor started")

	// Start monitoring in a separate goroutine
	go m.monitorLoop(ctx)

	return nil
}

func (m *Monitor) Stop() {
	if !m.isRunning {
		return
	}

	m.isRunning = false
	log.Println("Clipboard monitor stopped")
}

func (m *Monitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(m.config.MonitorInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkClipboard(ctx)
		}
	}
}

func (m *Monitor) checkClipboard(ctx context.Context) {
	// Try to read text first
	textData := clipboard.Read(clipboard.FmtText)
	if len(textData) > 0 {
		m.processClipboardData(ctx, &ClipboardData{
			Type:      "text",
			Content:   string(textData),
			Size:      len(textData),
			Timestamp: time.Now(),
		})
		return
	}

	// Try to read image
	imageData := clipboard.Read(clipboard.FmtImage)
	if len(imageData) > 0 {
		m.processClipboardData(ctx, &ClipboardData{
			Type:      "image",
			ImageData: imageData,
			Size:      len(imageData),
			Timestamp: time.Now(),
		})
		return
	}
}

func (m *Monitor) processClipboardData(ctx context.Context, data *ClipboardData) {
	// Check size limit
	if data.Size > m.config.MaxItemSize {
		log.Printf("Clipboard item too large: %d bytes (max: %d)", data.Size, m.config.MaxItemSize)
		return
	}

	// Generate hash
	hash := util.GenerateHash(data.Content, data.ImageData)

	// Skip if same as last item
	if hash == m.lastHash {
		return
	}

	m.lastHash = hash

	// Create database item
	item := &database.ClipboardItem{
		Type:      data.Type,
		Content:   data.Content,
		ImageData: data.ImageData,
		Size:      data.Size,
		Hash:      hash,
		Timestamp: data.Timestamp,
	}

	// Save to database
	if err := m.repository.SaveClipboardItem(ctx, item); err != nil {
		log.Printf("Failed to save clipboard item: %v", err)
		m.eventChan <- MonitorEvent{
			Type:  "error",
			Error: err,
		}
		return
	}

	// Notify listeners
	m.eventChan <- MonitorEvent{
		Type: "new_item",
		Data: data,
	}

	log.Printf("Saved clipboard item: %s (%d bytes)", data.Type, data.Size)
}

func (m *Monitor) CopyItemToClipboard(ctx context.Context, id int64) error {
	item, err := m.repository.GetItemByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	switch item.Type {
	case "text":
		clipboard.Write(clipboard.FmtText, []byte(item.Content))
	case "image":
		if len(item.ImageData) > 0 {
			clipboard.Write(clipboard.FmtImage, item.ImageData)
		}
	default:
		return fmt.Errorf("unsupported clipboard type: %s", item.Type)
	}

	// Update the hash to current so we don't re-capture this item
	m.lastHash = item.Hash

	log.Printf("Copied item to clipboard: %s", item.Type)
	return nil
}

func (m *Monitor) EventChannel() <-chan MonitorEvent {
	return m.eventChan
}
