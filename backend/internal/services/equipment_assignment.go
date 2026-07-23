// Package services holds business logic that spans multiple models and doesn't belong to any
// single HTTP handler.
package services

import (
	"errors"
	"time"

	"github.com/ofgrenudo/OmniAv/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrEquipmentGroupNotFound    = errors.New("equipment group not found")
	ErrEquipmentGroupUnavailable = errors.New("equipment group is disabled or archived")
	ErrNoEquipmentAvailable      = errors.New("no equipment in this group is available for the requested time")
	ErrRequestNotFound           = errors.New("request not found")
)

// RequestsOverlap reports whether two requests would ever occupy the same equipment at the same
// wall-clock moment. Request.DaysOfWeek is a single weekday, so two requests can only conflict if
// they recur on the same weekday, their weekly occurrence ranges intersect, and their time-of-day
// windows intersect.
func RequestsOverlap(a, b *models.Request) bool {
	if a.DaysOfWeek != b.DaysOfWeek {
		return false
	}

	aStart, aEnd := occurrenceRange(a)
	bStart, bEnd := occurrenceRange(b)
	if aStart.After(bEnd) || bStart.After(aEnd) {
		return false
	}

	return timeOfDayOverlaps(a.StartTime, a.EndTime, b.StartTime, b.EndTime)
}

func occurrenceRange(r *models.Request) (time.Time, time.Time) {
	start := dateOnly(r.FirstDateNeeded)
	weeks := r.NumberOfWeeks
	if weeks < 1 {
		weeks = 1
	}
	return start, start.AddDate(0, 0, 7*(weeks-1))
}

// dateOnly normalizes to UTC before truncating to a calendar day. Postgres round-trips
// timestamptz values converted to the server's local zone, not UTC, so comparing Year/Month/Day
// straight off a freshly-scanned time.Time can land on the wrong calendar day; normalizing first
// keeps this stable regardless of the process's local timezone.
func dateOnly(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// timeOfDayOverlaps compares only the hour/minute/second of each time.Time, since StartTime and
// EndTime carry a time-of-day that repeats every occurrence; the calendar date they're stored
// against is irrelevant. Touching endpoints (one ends exactly when the other starts) do not count
// as a conflict, matching normal back-to-back scheduling.
func timeOfDayOverlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	as, ae := secondOfDay(aStart), secondOfDay(aEnd)
	bs, be := secondOfDay(bStart), secondOfDay(bEnd)
	return as < be && bs < ae
}

func secondOfDay(t time.Time) int {
	t = t.UTC()
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}

// CreateRequestedEquipment resolves groupID (e.g. "Cow Cart") to a specific available Equipment
// unit (e.g. "Cow Cart B") for request's schedule and persists the assignment. Call it once per
// unit needed; calling it twice for the same request and group assigns two distinct units.
func CreateRequestedEquipment(db *gorm.DB, requestID, groupID uint) (*models.RequestedEquipment, error) {
	var created models.RequestedEquipment

	err := db.Transaction(func(tx *gorm.DB) error {
		var request models.Request
		if err := tx.First(&request, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRequestNotFound
			}
			return err
		}

		equipment, err := assignEquipment(tx, groupID, &request)
		if err != nil {
			return err
		}

		created = models.RequestedEquipment{
			GroupID:     groupID,
			EquipmentID: equipment.ID,
			RequestID:   requestID,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// assignEquipment picks the alphabetically-first non-disabled, non-archived Equipment unit in
// groupID with no conflicting booking for request's schedule. Candidate rows are locked with
// FOR UPDATE so concurrent assignments for the same group serialize instead of two callers
// racing to hand out the same physical unit; callers must run this inside a transaction that
// also performs the resulting insert.
func assignEquipment(tx *gorm.DB, groupID uint, request *models.Request) (*models.Equipment, error) {
	group, err := loadGroup(tx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Disabled || group.Archived {
		return nil, ErrEquipmentGroupUnavailable
	}

	candidates, err := loadCandidates(tx, groupID, true)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, ErrNoEquipmentAvailable
	}

	conflicted, err := conflictedEquipmentIDs(tx, equipmentIDs(candidates), request)
	if err != nil {
		return nil, err
	}

	for i := range candidates {
		if !conflicted[candidates[i].ID] {
			return &candidates[i], nil
		}
	}

	return nil, ErrNoEquipmentAvailable
}

// AvailableCount reports how many units in groupID could be assigned to request right now,
// without committing to one. Useful for showing remaining availability (e.g. the frontend's
// EquipmentItem.available) before the user submits a request.
func AvailableCount(db *gorm.DB, groupID uint, request *models.Request) (int, error) {
	group, err := loadGroup(db, groupID)
	if err != nil {
		return 0, err
	}
	if group.Disabled || group.Archived {
		return 0, nil
	}

	candidates, err := loadCandidates(db, groupID, false)
	if err != nil {
		return 0, err
	}
	if len(candidates) == 0 {
		return 0, nil
	}

	conflicted, err := conflictedEquipmentIDs(db, equipmentIDs(candidates), request)
	if err != nil {
		return 0, err
	}

	available := 0
	for _, c := range candidates {
		if !conflicted[c.ID] {
			available++
		}
	}
	return available, nil
}

func loadGroup(tx *gorm.DB, groupID uint) (*models.EquipmentGroup, error) {
	var group models.EquipmentGroup
	if err := tx.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEquipmentGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

func loadCandidates(tx *gorm.DB, groupID uint, lock bool) ([]models.Equipment, error) {
	q := tx.Where("group_id = ? AND disabled = ? AND archived = ?", groupID, false, false).Order("name ASC")
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var candidates []models.Equipment
	if err := q.Find(&candidates).Error; err != nil {
		return nil, err
	}
	return candidates, nil
}

// conflictedEquipmentIDs returns the subset of candidateIDs that already have a RequestedEquipment
// whose parent Request overlaps request's schedule. A booking made by request itself (e.g. a
// second unit requested for the same group) still counts, so the same request never gets handed
// the same physical unit twice.
func conflictedEquipmentIDs(tx *gorm.DB, candidateIDs []uint, request *models.Request) (map[uint]bool, error) {
	var booked []models.RequestedEquipment
	if err := tx.Preload("Request").Where("equipment_id IN ?", candidateIDs).Find(&booked).Error; err != nil {
		return nil, err
	}

	conflicted := make(map[uint]bool, len(booked))
	for _, re := range booked {
		if conflicted[re.EquipmentID] || re.Request == nil {
			continue
		}
		if RequestsOverlap(re.Request, request) {
			conflicted[re.EquipmentID] = true
		}
	}
	return conflicted, nil
}

func equipmentIDs(candidates []models.Equipment) []uint {
	ids := make([]uint, len(candidates))
	for i, c := range candidates {
		ids[i] = c.ID
	}
	return ids
}
