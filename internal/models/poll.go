package models

import "time"

type Poll struct {
	ID        string           `json:"id" mapstructure:"id"` // UUID
	Question  string           `json:"question" mapstructure:"question"`
	Options   []string         `json:"options,omitempty" mapstructure:"options"`
	Votes     map[string]int64 `json:"votes,omitempty" mapstructure:"votes"`
	CreatedBy string           `json:"created_by,omitempty" mapstructure:"created_by"`
	UpdatedBy string           `json:"updated_by,omitempty" mapstructure:"updated_by"`
	CreatedAt time.Time        `json:"created_at,omitempty" mapstructure:"created_at"`
	ExpiresAt time.Time        `json:"expires_at,omitempty" mapstructure:"expires_at"`
	IsClosed  bool             `json:"is_closed,omitempty" mapstructure:"is_closed"`
}
