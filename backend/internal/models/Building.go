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
	// Equipment is a pointer slice field populated only via Preload; a unit's BuildingID is the
	// source of truth for the one-building-per-unit relationship.
	Equipment []Equipment `json:"equipment,omitempty" gorm:"foreignKey:BuildingID"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
