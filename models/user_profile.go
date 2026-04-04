package models

import "time"

type User_profiles struct {
	ID            uint      `json:"id"`
	User_ID       uint      `json:"user_id"`
	Profile_Image string    `json:"profile_image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
