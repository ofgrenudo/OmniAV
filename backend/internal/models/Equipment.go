package models

import (
	"time"
)

type Equipment struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Disabled    bool    `json:"disabled"`
	Archived    bool    `json:"archived"`
	// GroupID is required: the frontend only ever presents EquipmentGroup (e.g. "Cow Cart"),
	// never an individual Equipment unit (e.g. "Cow Cart A"), so every unit must belong to one.
	GroupID uint `json:"groupId" gorm:"not null"`
	// Group is a pointer so it's omitted from JSON entirely when not Preloaded, instead of
	// serializing as a misleading zero-value record.
	Group *EquipmentGroup `json:"group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
	// BuildingID is the unit's permanent home. Every unit is stocked in exactly one building; a
	// move is an admin edit to this column, not a side effect of a request. Non-null enforced at
	// the DB level so orphaned "in central storage" units can't exist by accident.
	BuildingID uint      `json:"buildingId" gorm:"not null"`
	Building   *Building `json:"building,omitempty" gorm:"foreignKey:BuildingID;references:ID"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
