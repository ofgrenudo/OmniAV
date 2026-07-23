package models

import (
	"time"
)

type Building struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Archived    bool      `json:"archived"`
	Description *string   `json:"description"`
	Address     *string   `json:"address"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
