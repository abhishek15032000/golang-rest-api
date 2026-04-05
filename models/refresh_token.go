package models

import "time"

type RefreshToken struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	Expiry    time.Time `json:"expiry"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
