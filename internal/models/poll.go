package models

import "time"

type Poll struct {
	ID        string    `json:"id" mapstructure:"id"` // UUID
	Question  string    `json:"question" mapstructure:"question"`
	Options   []string  `json:"options" mapstructure:"options"`
	CreatedBy string    `json:"created_by" mapstructure:"created_by"`
	UpdatedBy string    `json:"updated_by" mapstructure:"updated_by"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	ExpiresAt time.Time `json:"expires_at" mapstructure:"expires_at"`
	IsClosed  bool      `json:"is_closed" mapstructure:"is_closed"`
}
