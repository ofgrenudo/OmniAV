package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	code, err := runWithContainer(m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runWithContainer(m *testing.M) (int, error) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("omniav_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		return 1, fmt.Errorf("start postgres container: %w", err)
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return 1, fmt.Errorf("get connection string: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return 1, fmt.Errorf("connect gorm: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Building{},
		&models.BuildingRoom{},
		&models.EquipmentGroup{},
		&models.Equipment{},
		&models.Request{},
		&models.RequestedEquipment{},
	); err != nil {
		return 1, fmt.Errorf("migrate: %w", err)
	}

	testDB = db

	return m.Run(), nil
}

func resetTables(t *testing.T) {
	t.Helper()
	if err := testDB.Exec(
		"TRUNCATE TABLE requested_equipments, requests, equipment, equipment_groups, buildings RESTART IDENTITY CASCADE",
	).Error; err != nil {
		t.Fatalf("failed to reset tables: %v", err)
	}
}

func clockTime(hour, minute int) time.Time {
	return time.Date(2000, 1, 1, hour, minute, 0, 0, time.UTC)
}

var goWeekdayToModel = map[time.Weekday]models.DaysOfWeekT{
	time.Monday:    models.Monday,
	time.Tuesday:   models.Tuesday,
	time.Wednesday: models.Wednesday,
	time.Thursday:  models.Thursday,
	time.Friday:    models.Friday,
	time.Saturday:  models.Saturday,
	time.Sunday:    models.Sunday,
}

func seedBuilding(t *testing.T) models.Building {
	t.Helper()
	b := models.Building{Name: "Test Hall"}
	if err := testDB.Create(&b).Error; err != nil {
		t.Fatalf("seed building: %v", err)
	}
	return b
}

func seedGroup(t *testing.T, name string, disabled, archived bool) models.EquipmentGroup {
	t.Helper()
	g := models.EquipmentGroup{Name: name, Disabled: disabled, Archived: archived}
	if err := testDB.Create(&g).Error; err != nil {
		t.Fatalf("seed group %q: %v", name, err)
	}
	return g
}

func seedEquipment(t *testing.T, groupID uint, name string, disabled, archived bool) models.Equipment {
	t.Helper()
	e := models.Equipment{Name: name, GroupID: groupID, Disabled: disabled, Archived: archived}
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
}

func seedRequest(t *testing.T, o requestOpts) models.Request {
	t.Helper()
	weeks := o.weeks
	if weeks == 0 {
		weeks = 1
	}
	r := models.Request{
		Name:            o.name,
		FirstDateNeeded: o.firstDate,
		StartTime:       o.startTime,
		EndTime:         o.endTime,
		NumberOfWeeks:   weeks,
		DaysOfWeek:      goWeekdayToModel[o.firstDate.Weekday()],
		BuildingID:      o.buildingID,
		Room:            "101",
	}
	if err := testDB.Create(&r).Error; err != nil {
		t.Fatalf("seed request %q: %v", o.name, err)
	}
	return r
}

// --- Unit tests: pure overlap math, no DB involved ---

func TestRequestsOverlap(t *testing.T) {
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // a Monday
	if monday.Weekday() != time.Monday {
		t.Fatalf("test fixture bug: %v is not a Monday", monday)
	}

	base := func() models.Request {
		return models.Request{
			FirstDateNeeded: monday,
			StartTime:       clockTime(9, 0),
			EndTime:         clockTime(11, 0),
			NumberOfWeeks:   1,
			DaysOfWeek:      models.Monday,
		}
	}

	t.Run("different weekday never conflicts", func(t *testing.T) {
		a := base()
		b := base()
		b.DaysOfWeek = models.Tuesday
		if RequestsOverlap(&a, &b) {
			t.Error("expected no overlap for different weekdays")
		}
	})

	t.Run("same day overlapping times conflict", func(t *testing.T) {
		a := base()
		b := base()
		b.StartTime = clockTime(10, 0)
		b.EndTime = clockTime(12, 0)
		if !RequestsOverlap(&a, &b) {
			t.Error("expected overlap for 9-11 vs 10-12")
		}
	})

	t.Run("back-to-back times do not conflict", func(t *testing.T) {
		a := base()
		b := base()
		b.StartTime = clockTime(11, 0)
		b.EndTime = clockTime(13, 0)
		if RequestsOverlap(&a, &b) {
			t.Error("expected no overlap for 9-11 vs 11-13 (touching boundary)")
		}
	})

	t.Run("disjoint times do not conflict", func(t *testing.T) {
		a := base()
		b := base()
		b.StartTime = clockTime(14, 0)
		b.EndTime = clockTime(15, 0)
		if RequestsOverlap(&a, &b) {
			t.Error("expected no overlap for 9-11 vs 14-15")
		}
	})

	t.Run("disjoint week ranges do not conflict", func(t *testing.T) {
		a := base()
		a.NumberOfWeeks = 2 // occurs on monday, monday+7
		b := base()
		b.FirstDateNeeded = monday.AddDate(0, 0, 21) // three weeks later
		b.NumberOfWeeks = 1
		if RequestsOverlap(&a, &b) {
			t.Error("expected no overlap for non-intersecting week ranges")
		}
	})

	t.Run("overlapping week ranges conflict on the shared occurrence", func(t *testing.T) {
		a := base()
		a.NumberOfWeeks = 4 // monday, +7, +14, +21
		b := base()
		b.FirstDateNeeded = monday.AddDate(0, 0, 21) // shares the last occurrence of a
		b.NumberOfWeeks = 2
		if !RequestsOverlap(&a, &b) {
			t.Error("expected overlap where week ranges share an occurrence date")
		}
	})

	t.Run("zero or negative NumberOfWeeks is treated as a single occurrence", func(t *testing.T) {
		a := base()
		a.NumberOfWeeks = 0
		b := base()
		b.FirstDateNeeded = monday.AddDate(0, 0, 7)
		b.NumberOfWeeks = -3
		if RequestsOverlap(&a, &b) {
			t.Error("expected no overlap: each request should span only its own single day")
		}
	})
}

// --- Integration tests: assignment, conflicts, and availability against a real Postgres ---

func TestCreateRequestedEquipment(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart B", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)

	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Bio Lecture",
		firstDate:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), // a Monday
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
		weeks:      1,
	})

	t.Run("assigns the alphabetically-first available unit", func(t *testing.T) {
		got, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var assigned models.Equipment
		if err := testDB.First(&assigned, got.EquipmentID).Error; err != nil {
			t.Fatalf("failed to load assigned equipment: %v", err)
		}
		if assigned.Name != "Cow Cart A" {
			t.Errorf("assigned %q, want Cow Cart A", assigned.Name)
		}
		if got.GroupID != group.ID {
			t.Errorf("GroupID = %d, want %d", got.GroupID, group.ID)
		}
	})

	t.Run("a second unit for the same request gets a different one", func(t *testing.T) {
		got, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var assigned models.Equipment
		if err := testDB.First(&assigned, got.EquipmentID).Error; err != nil {
			t.Fatalf("failed to load assigned equipment: %v", err)
		}
		if assigned.Name != "Cow Cart B" {
			t.Errorf("assigned %q, want Cow Cart B (the other unit already used by this request)", assigned.Name)
		}
	})

	t.Run("a third unit for the same request has nothing left", func(t *testing.T) {
		_, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
		if !errors.Is(err, ErrNoEquipmentAvailable) {
			t.Fatalf("expected ErrNoEquipmentAvailable, got %v", err)
		}
	})
}

func TestCreateRequestedEquipment_SkipsDisabledAndArchived(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A (disabled)", true, false)
	seedEquipment(t, group.ID, "Cow Cart B (archived)", false, true)
	seedEquipment(t, group.ID, "Cow Cart C", false, false)

	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Lab Session",
		firstDate:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	got, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var assigned models.Equipment
	if err := testDB.First(&assigned, got.EquipmentID).Error; err != nil {
		t.Fatalf("failed to load assigned equipment: %v", err)
	}
	if assigned.Name != "Cow Cart C" {
		t.Errorf("assigned %q, want Cow Cart C (only non-disabled, non-archived unit)", assigned.Name)
	}
}

func TestCreateRequestedEquipment_NoUnitsInGroup(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Empty Group", false, false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Whatever",
		firstDate:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	_, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
	if !errors.Is(err, ErrNoEquipmentAvailable) {
		t.Fatalf("expected ErrNoEquipmentAvailable, got %v", err)
	}
}

func TestCreateRequestedEquipment_GroupNotFound(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Whatever",
		firstDate:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	_, err := CreateRequestedEquipment(testDB, request.ID, 999999)
	if !errors.Is(err, ErrEquipmentGroupNotFound) {
		t.Fatalf("expected ErrEquipmentGroupNotFound, got %v", err)
	}
}

func TestCreateRequestedEquipment_GroupUnavailable(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Whatever",
		firstDate:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	t.Run("disabled group", func(t *testing.T) {
		group := seedGroup(t, "Disabled Group", true, false)
		seedEquipment(t, group.ID, "Unit A", false, false)

		_, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
		if !errors.Is(err, ErrEquipmentGroupUnavailable) {
			t.Fatalf("expected ErrEquipmentGroupUnavailable, got %v", err)
		}
	})

	t.Run("archived group", func(t *testing.T) {
		group := seedGroup(t, "Archived Group", false, true)
		seedEquipment(t, group.ID, "Unit A", false, false)

		_, err := CreateRequestedEquipment(testDB, request.ID, group.ID)
		if !errors.Is(err, ErrEquipmentGroupUnavailable) {
			t.Fatalf("expected ErrEquipmentGroupUnavailable, got %v", err)
		}
	})
}

func TestCreateRequestedEquipment_RequestNotFound(t *testing.T) {
	resetTables(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)

	_, err := CreateRequestedEquipment(testDB, 999999, group.ID)
	if !errors.Is(err, ErrRequestNotFound) {
		t.Fatalf("expected ErrRequestNotFound, got %v", err)
	}
}

func TestCreateRequestedEquipment_RespectsOverlappingBookings(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)

	monday := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	firstRequest := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "First booking",
		firstDate:  monday,
		startTime:  clockTime(9, 0),
		endTime:    clockTime(11, 0),
		weeks:      4,
	})
	if _, err := CreateRequestedEquipment(testDB, firstRequest.ID, group.ID); err != nil {
		t.Fatalf("failed to seed the first booking: %v", err)
	}

	t.Run("an overlapping request cannot get the only unit", func(t *testing.T) {
		overlapping := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "Conflicting booking",
			firstDate:  monday,
			startTime:  clockTime(10, 0),
			endTime:    clockTime(12, 0),
			weeks:      1,
		})

		_, err := CreateRequestedEquipment(testDB, overlapping.ID, group.ID)
		if !errors.Is(err, ErrNoEquipmentAvailable) {
			t.Fatalf("expected ErrNoEquipmentAvailable, got %v", err)
		}
	})

	t.Run("a non-overlapping request (different time) can still get the unit", func(t *testing.T) {
		nonOverlapping := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "Later that day",
			firstDate:  monday,
			startTime:  clockTime(13, 0),
			endTime:    clockTime(14, 0),
			weeks:      1,
		})

		got, err := CreateRequestedEquipment(testDB, nonOverlapping.ID, group.ID)
		if err != nil {
			t.Fatalf("expected the unit to be available for a non-overlapping time, got error: %v", err)
		}
		if got.EquipmentID == 0 {
			t.Error("expected an assigned equipment id")
		}
	})

	t.Run("a non-overlapping request (different weekday) can still get the unit", func(t *testing.T) {
		nonOverlapping := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "Different weekday",
			firstDate:  monday.AddDate(0, 0, 1), // Tuesday
			startTime:  clockTime(9, 0),
			endTime:    clockTime(11, 0),
			weeks:      4,
		})

		_, err := CreateRequestedEquipment(testDB, nonOverlapping.ID, group.ID)
		if err != nil {
			t.Fatalf("expected the unit to be available on a different weekday, got error: %v", err)
		}
	})

	t.Run("a non-overlapping request (different week range) can still get the unit", func(t *testing.T) {
		nonOverlapping := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "Much later",
			firstDate:  monday.AddDate(0, 0, 90),
			startTime:  clockTime(9, 0),
			endTime:    clockTime(11, 0),
			weeks:      1,
		})

		_, err := CreateRequestedEquipment(testDB, nonOverlapping.ID, group.ID)
		if err != nil {
			t.Fatalf("expected the unit to be available far outside the booked week range, got error: %v", err)
		}
	})
}

func TestCreateRequestedEquipment_ConcurrentRequestsForLastUnit(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)

	monday := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	requestA := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Racer A",
		firstDate:  monday,
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})
	requestB := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Racer B",
		firstDate:  monday,
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	var wg sync.WaitGroup
	results := make([]error, 2)
	ids := []uint{requestA.ID, requestB.ID}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := CreateRequestedEquipment(testDB, ids[i], group.ID)
			results[i] = err
		}(i)
	}
	wg.Wait()

	successes, failures := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrNoEquipmentAvailable):
			failures++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 1 || failures != 1 {
		t.Fatalf("expected exactly one success and one ErrNoEquipmentAvailable, got %d successes and %d failures", successes, failures)
	}

	var count int64
	if err := testDB.Model(&models.RequestedEquipment{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count requested_equipment rows: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 requested_equipment row, got %d", count)
	}
}

func TestAvailableCount(t *testing.T) {
	resetTables(t)
	building := seedBuilding(t)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)
	seedEquipment(t, group.ID, "Cow Cart B", false, false)
	seedEquipment(t, group.ID, "Cow Cart C (disabled)", true, false)

	monday := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	existing := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Existing booking",
		firstDate:  monday,
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})
	if _, err := CreateRequestedEquipment(testDB, existing.ID, group.ID); err != nil {
		t.Fatalf("failed to seed existing booking: %v", err)
	}

	t.Run("counts down for an overlapping request", func(t *testing.T) {
		newRequest := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "New request",
			firstDate:  monday,
			startTime:  clockTime(9, 30),
			endTime:    clockTime(10, 30),
		})

		got, err := AvailableCount(testDB, group.ID, &newRequest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("AvailableCount = %d, want 1 (one of two enabled units is booked)", got)
		}
	})

	t.Run("full count for a non-overlapping request", func(t *testing.T) {
		newRequest := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "New request, different time",
			firstDate:  monday,
			startTime:  clockTime(14, 0),
			endTime:    clockTime(15, 0),
		})

		got, err := AvailableCount(testDB, group.ID, &newRequest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("AvailableCount = %d, want 2 (both enabled units are free at this time)", got)
		}
	})

	t.Run("zero for a disabled group", func(t *testing.T) {
		disabledGroup := seedGroup(t, "Disabled Group", true, false)
		seedEquipment(t, disabledGroup.ID, "Unit A", false, false)
		newRequest := seedRequest(t, requestOpts{
			buildingID: building.ID,
			name:       "Whatever",
			firstDate:  monday,
			startTime:  clockTime(14, 0),
			endTime:    clockTime(15, 0),
		})

		got, err := AvailableCount(testDB, disabledGroup.ID, &newRequest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("AvailableCount = %d, want 0 for a disabled group", got)
		}
	})
}
