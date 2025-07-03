package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	"clipboardpro/internal/util"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	repo := &Repository{db: db}

	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

func (r *Repository) migrate() error {
	ctx := context.Background()

	// Create tables
	models := []interface{}{
		(*ClipboardItem)(nil),
	}

	for _, model := range models {
		if _, err := r.db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return fmt.Errorf("failed to create table for %T: %w", model, err)
		}
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_clipboard_timestamp ON clipboard_items(timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_clipboard_hash ON clipboard_items(hash)",
		"CREATE INDEX IF NOT EXISTS idx_clipboard_pinned ON clipboard_items(pinned)",
		"CREATE INDEX IF NOT EXISTS idx_clipboard_type ON clipboard_items(type)",
	}

	for _, idx := range indexes {
		if _, err := r.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (r *Repository) SaveClipboardItem(ctx context.Context, item *ClipboardItem) error {
	// Generate hash if not provided
	if item.Hash == "" {
		item.Hash = util.GenerateHash(item.Content, item.ImageData)
	}

	// Check if item with same hash already exists
	exists, err := r.db.NewSelect().
		Model((*ClipboardItem)(nil)).
		Where("hash = ?", item.Hash).
		Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing item: %w", err)
	}

	if exists {
		// Update timestamp to move to top
		_, err = r.db.NewUpdate().
			Model((*ClipboardItem)(nil)).
			Set("timestamp = ?", time.Now()).
			Set("updated_at = ?", time.Now()).
			Where("hash = ?", item.Hash).
			Exec(ctx)
		return err
	}

	// Set timestamps
	now := time.Now()
	item.Timestamp = now
	item.CreatedAt = now
	item.UpdatedAt = now

	// Insert new item
	_, err = r.db.NewInsert().Model(item).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert clipboard item: %w", err)
	}

	return nil
}

func (r *Repository) GetRecentItems(ctx context.Context, limit int) ([]*ClipboardItem, error) {
	var items []*ClipboardItem

	err := r.db.NewSelect().
		Model(&items).
		Order("pinned DESC", "timestamp DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent items: %w", err)
	}

	return items, nil
}

func (r *Repository) SearchItems(ctx context.Context, query string, limit int) ([]*ClipboardItem, error) {
	var items []*ClipboardItem

	err := r.db.NewSelect().
		Model(&items).
		Where("content LIKE ? OR title LIKE ?", "%"+query+"%", "%"+query+"%").
		Order("pinned DESC", "timestamp DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}

	return items, nil
}

func (r *Repository) GetItemByID(ctx context.Context, id int64) (*ClipboardItem, error) {
	var item ClipboardItem
	err := r.db.NewSelect().
		Model(&item).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get item by ID: %w", err)
	}

	return &item, nil
}

func (r *Repository) TogglePin(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model((*ClipboardItem)(nil)).
		Set("pinned = NOT pinned").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to toggle pin: %w", err)
	}

	return nil
}

func (r *Repository) DeleteItem(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*ClipboardItem)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

func (r *Repository) UpdateTitle(ctx context.Context, id int64, title string) error {
	_, err := r.db.NewUpdate().
		Model((*ClipboardItem)(nil)).
		Set("title = ?", title).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update title: %w", err)
	}

	return nil
}

func (r *Repository) CleanupOldItems(ctx context.Context, maxDays int, maxItems int) error {
	cutoffDate := time.Now().AddDate(0, 0, -maxDays)

	// Delete old unpinned items
	_, err := r.db.NewDelete().
		Model((*ClipboardItem)(nil)).
		Where("timestamp < ? AND pinned = FALSE", cutoffDate).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	// Keep only the most recent items (excluding pinned)
	subquery := r.db.NewSelect().
		Model((*ClipboardItem)(nil)).
		Column("id").
		Where("pinned = FALSE").
		Order("timestamp DESC").
		Limit(maxItems)

	_, err = r.db.NewDelete().
		Model((*ClipboardItem)(nil)).
		Where("pinned = FALSE").
		Where("id NOT IN (?)", subquery).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to cleanup excess items: %w", err)
	}

	return nil
}

func (r *Repository) ClearAllItems(ctx context.Context) error {
	_, err := r.db.NewDelete().Model((*ClipboardItem)(nil)).Where("1=1").Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear all items: %w", err)
	}
	return nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}
