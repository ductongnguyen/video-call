package models

import (
	"time"
)

type ShortURL struct {
	ID          uint64     `db:"id" json:"id"`
	OriginalURL string     `db:"original_url" json:"original_url"`
	ShortCode   string     `db:"short_code" json:"short_code"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	ExpiredAt   *time.Time `db:"expired_at" json:"expired_at,omitempty"`
	ClickCount  uint       `db:"click_count" json:"click_count"`
	CreatorIP   *string    `db:"creator_ip" json:"creator_ip,omitempty"`
	UserAgent   *string    `db:"user_agent" json:"user_agent,omitempty"`
}
