package models

import "time"

type Otp_verification struct {
	ID        int32     `json:"id"`
	OtpKey    string    `json:"otp_key"`
	UserID    int32     `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `json:"is_used"`
}
