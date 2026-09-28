package models

import "gorm.io/gorm"

const requestBuildingRoomConstraint = "fk_requests_building_room"

// AddRequestRoomConstraint adds the composite FK requests(building_id, room) ->
// building_rooms(building_id, room). It's done with raw SQL because GORM can't model it from
// struct tags: Request.BuildingID is already the FK for Request.Building, which makes GORM
// misinfer the composite association's direction. Idempotent; call after both tables exist.
func AddRequestRoomConstraint(db *gorm.DB) error {
	if db.Migrator().HasConstraint("requests", requestBuildingRoomConstraint) {
		return nil
	}
	return db.Exec(
		"ALTER TABLE requests ADD CONSTRAINT " + requestBuildingRoomConstraint +
			" FOREIGN KEY (building_id, room) REFERENCES building_rooms (building_id, room)" +
			" ON UPDATE CASCADE ON DELETE RESTRICT",
	).Error
}
