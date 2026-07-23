package models

import "time"

type EquipmentGroup struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Description *string     `json:"description"`
	Disabled    bool        `json:"disabled"`
	Archived    bool        `json:"archived"`
	Equipment   []Equipment `json:"equipment,omitempty" gorm:"foreignKey:GroupID"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}
