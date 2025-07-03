package clipboard

import (
	"time"
)

type ClipboardData struct {
	Type      string
	Content   string
	ImageData []byte
	Size      int

	Timestamp time.Time
}

type MonitorEvent struct {
	Type  string
	Data  *ClipboardData
	Error error
}
