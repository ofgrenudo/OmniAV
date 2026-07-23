package models

import (
	"time"
)

type Request struct {
	ID              uint        `json:"id"`
	Name            string      `json:"name"`
	FirstDateNeeded time.Time   `json:"firstDateNeeded"`
	StartTime       time.Time   `json:"startTime"`
	EndTime         time.Time   `json:"endTime"`
	NumberOfWeeks   int         `json:"numberOfWeeks"`
	DaysOfWeek      DaysOfWeekT `json:"daysOfWeek"`
	BuildingID      uint        `json:"buildingId"`
	Room            string      `json:"room"`
	// Building is a pointer so it's omitted from JSON entirely when not Preloaded, instead of
	// serializing as a misleading zero-value record.
	Building *Building `json:"building,omitempty" gorm:"foreignKey:BuildingID;references:ID"`
	// -:migration: BuildingID is also the FK for Building above, which makes GORM misinfer this
	// composite association's direction and try to add an invalid FK on building_rooms referencing
	// requests(building_id, room) — a column pair that isn't unique on requests. Kept for Preload/queries,
	// skipped at migration time.
	BuildingRoom BuildingRoom `json:"-" gorm:"foreignKey:BuildingID,Room;references:BuildingID,Room;-:migration"`
	Comments     *string      `json:"comments"`
	AttachmentID *uint        `json:"attachmentId"` // todo(jwb): set with fk once an Attachment model exists
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`

	RequestedEquipment []RequestedEquipment `json:"requestedEquipment,omitempty" gorm:"foreignKey:RequestID"`
}
