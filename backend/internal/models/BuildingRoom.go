package models

type BuildingRoom struct {
	// BUG: GORM does not support fk as a composite primary key. Thats fine, but that means that if someone creates a room without a proper BuildingID,
	// we will end up with a bunch of orphaned rooms. Which is probably okay.
	BuildingID uint   `gorm:"primaryKey"`
	Room       string `gorm:"primaryKey"`
}
