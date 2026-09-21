package handlers

import (
	"testing"
	"time"

	"github.com/ofgrenudo/OmniAv/internal/models"
)

func seedGroup(t *testing.T, name string, disabled, archived bool) models.EquipmentGroup {
	t.Helper()
	g := models.EquipmentGroup{Name: name, Disabled: disabled, Archived: archived}
	if err := testDB.Create(&g).Error; err != nil {
		t.Fatalf("seed group %q: %v", name, err)
	}
	return g
}

// seedEquipment stocks a unit in buildingID. Every unit has a permanent home building, so
// buildingID is required — there is no "central storage" seeder.
func seedEquipment(t *testing.T, groupID, buildingID uint, name string, disabled, archived bool) models.Equipment {
	t.Helper()
	e := models.Equipment{Name: name, GroupID: groupID, BuildingID: buildingID, Disabled: disabled, Archived: archived}
	if err := testDB.Create(&e).Error; err != nil {
		t.Fatalf("seed equipment %q: %v", name, err)
	}
	return e
}

type requestOpts struct {
	buildingID uint
	name       string
	firstDate  time.Time
	startTime  time.Time
	endTime    time.Time
	weeks      int
	room       string
}

func seedRequest(t *testing.T, o requestOpts) models.Request {
	t.Helper()
	weeks := o.weeks
	if weeks == 0 {
		weeks = 1
	}
	room := o.room
	if room == "" {
		room = "101"
	}
	r := models.Request{
		Name:            o.name,
		FirstDateNeeded: o.firstDate,
		StartTime:       o.startTime,
		EndTime:         o.endTime,
		NumberOfWeeks:   weeks,
		DaysOfWeek:      models.FromGoWeekday(o.firstDate.Weekday()),
		BuildingID:      o.buildingID,
		Room:            room,
	}
	if err := testDB.Create(&r).Error; err != nil {
		t.Fatalf("seed request %q: %v", o.name, err)
	}
	return r
}

func clockTime(hour, minute int) time.Time {
	return time.Date(2000, 1, 1, hour, minute, 0, 0, time.UTC)
}

// futureDate returns a date daysFromNow days ahead of the real current date, at midnight UTC.
// Tests use this instead of a hardcoded literal so they stay valid (and satisfy the API's
// 24-hour-advance rule) no matter when the suite actually runs.
func futureDate(daysFromNow int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, daysFromNow)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
