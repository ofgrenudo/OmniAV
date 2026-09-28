package models

// BuildingRoom is the catalog behind the room picker: one row per (building, room). Requests
// reference it through the composite FK added by AddRequestRoomConstraint, so a request's room
// is guaranteed to exist in the request's building.
type BuildingRoom struct {
	BuildingID uint      `gorm:"primaryKey"`
	Room       string    `gorm:"primaryKey"`
	Building   *Building `json:"-" gorm:"foreignKey:BuildingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
