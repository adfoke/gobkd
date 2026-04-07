package model

import "time"

type User struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
