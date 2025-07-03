package database

import (
	"time"

	"github.com/uptrace/bun"
)

type ClipboardItem struct {
	bun.BaseModel `bun:"table:clipboard_items"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Type      string    `bun:"type,notnull" json:"type"`
	Content   string    `bun:"content" json:"content"`
	ImageData []byte    `bun:"image_data" json:"-"`
	Timestamp time.Time `bun:"timestamp,notnull,default:current_timestamp" json:"timestamp"`
	Size      int       `bun:"size,notnull" json:"size"`
	Hash      string    `bun:"hash,unique,notnull" json:"hash"`
	Pinned    bool      `bun:"pinned,default:false" json:"pinned"`
	Title     string    `bun:"title" json:"title"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
