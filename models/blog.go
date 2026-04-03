package models

import "time"

type Blog struct {
	ID        uint      `json:"id"`
	Ttile     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
