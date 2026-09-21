package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/ofgrenudo/OmniAv/internal/services"
)

func TestEquipmentGroupResponseContract(t *testing.T) {
	r := newTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/equipment-groups", equipmentGroupInput{Name: "Cow Cart"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON[map[string]any](t, w)
	assertExactKeys(t, "create response", body, "id", "name", "description", "disabled", "archived", "createdAt", "updatedAt")
}

func TestEquipmentGroupCreate(t *testing.T) {
	r := newTestRouter(t)

	t.Run("creates a group", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment-groups", equipmentGroupInput{Name: "Cow Cart"})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.EquipmentGroup](t, w)
		if got.ID == 0 || got.Name != "Cow Cart" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment-groups", equipmentGroupInput{Name: "   "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGroupGet(t *testing.T) {
	r := newTestRouter(t)
	g := seedGroup(t, "Cow Cart", false, false)

	t.Run("returns an existing group", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment-groups/%d", g.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("400s for a non-numeric id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups/nope", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGroupUpdate(t *testing.T) {
	r := newTestRouter(t)
	g := seedGroup(t, "Cow Cart", false, false)

	t.Run("updates an existing group", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment-groups/%d", g.ID), equipmentGroupInput{
			Name:     "Cow Cart (renamed)",
			Disabled: true,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.EquipmentGroup](t, w)
		if got.Name != "Cow Cart (renamed)" || !got.Disabled {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment-groups/%d", g.ID), equipmentGroupInput{Name: ""})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, "/api/equipment-groups/999999", equipmentGroupInput{Name: "Nope"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGroupDelete(t *testing.T) {
	r := newTestRouter(t)
	g := seedGroup(t, "Cow Cart", false, false)

	t.Run("archives instead of hard-deleting", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, fmt.Sprintf("/api/equipment-groups/%d", g.ID), nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		var reloaded models.EquipmentGroup
		if err := testDB.First(&reloaded, g.ID).Error; err != nil {
			t.Fatalf("expected group to still exist: %v", err)
		}
		if !reloaded.Archived {
			t.Error("expected group to be archived")
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, "/api/equipment-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type equipmentGroupListEnvelope struct {
	Data []models.EquipmentGroup `json:"data"`
	Meta paginationMeta          `json:"meta"`
}

func TestEquipmentGroupListSearchAndFilter(t *testing.T) {
	r := newTestRouter(t)
	seedGroup(t, "Cow Cart", false, false)
	seedGroup(t, "Projector Kit", false, false)
	seedGroup(t, "Old Kit", true, true)

	t.Run("searches case-insensitively by name", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups?q=cow", nil)
		got := decodeJSON[equipmentGroupListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Cow Cart" {
			t.Errorf("expected only Cow Cart, got %+v", got.Data)
		}
	})

	t.Run("filters by archived", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups?archived=true", nil)
		got := decodeJSON[equipmentGroupListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Old Kit" {
			t.Errorf("expected only Old Kit, got %+v", got.Data)
		}
	})

	t.Run("filters by disabled", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups?disabled=false", nil)
		got := decodeJSON[equipmentGroupListEnvelope](t, w)
		if len(got.Data) != 2 {
			t.Errorf("expected 2 non-disabled groups, got %d", len(got.Data))
		}
	})
}

func TestEquipmentGroupAvailability(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Main Hall", false)
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart B", false, false)

	date := formatDate(futureDate(14))

	t.Run("reports the full count with no bookings", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00&weeks=1&buildingId=%d", group.ID, date, building.ID,
		), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[map[string]int](t, w)
		if got["available"] != 2 {
			t.Errorf("available = %d, want 2", got["available"])
		}
	})

	t.Run("404s for a nonexistent group", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/999999/availability?firstDate=%s&startTime=09:00&endTime=10:00&buildingId=%d", date, building.ID,
		), nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	invalidCases := []string{
		fmt.Sprintf("/api/equipment-groups/%d/availability", group.ID),
		fmt.Sprintf("/api/equipment-groups/%d/availability?firstDate=not-a-date&startTime=09:00&endTime=10:00", group.ID),
		fmt.Sprintf("/api/equipment-groups/%d/availability?firstDate=%s&startTime=nope&endTime=10:00", group.ID, date),
		fmt.Sprintf("/api/equipment-groups/%d/availability?firstDate=%s&startTime=10:00&endTime=09:00", group.ID, date),
		fmt.Sprintf("/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00&weeks=0", group.ID, date),
	}
	for _, target := range invalidCases {
		t.Run("rejects invalid query: "+target, func(t *testing.T) {
			w := doRequest(r, http.MethodGet, target, nil)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %s, got %d: %s", target, w.Code, w.Body.String())
			}
		})
	}
}

func TestListEquipmentGroups_FiltersByBuilding(t *testing.T) {
	r := newTestRouter(t)
	buildingA := seedBuilding(t, "Anna Whitten Hall", false)
	buildingB := seedBuilding(t, "Groves", false)

	stocked := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, stocked.ID, buildingA.ID, "Cow Cart A", false, false)

	elsewhere := seedGroup(t, "Projector Kit", false, false)
	seedEquipment(t, elsewhere.ID, buildingB.ID, "Projector 1", false, false)

	unusable := seedGroup(t, "Broken Kit", false, false)
	seedEquipment(t, unusable.ID, buildingA.ID, "Broken Unit", true, false)

	seedGroup(t, "Empty Group", false, false) // no units at all

	w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment-groups?buildingId=%d", buildingA.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	got := decodeJSON[equipmentGroupListEnvelope](t, w)
	if len(got.Data) != 1 || got.Data[0].Name != "Cow Cart" {
		t.Fatalf("expected only Cow Cart, got %+v", got.Data)
	}

	t.Run("rejects an invalid buildingId", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups?buildingId=nope", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGroupAvailability_ScopedToBuilding(t *testing.T) {
	r := newTestRouter(t)
	buildingA := seedBuilding(t, "Anna Whitten Hall", false)
	buildingB := seedBuilding(t, "Groves", false)

	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, buildingA.ID, "Cow Cart A", false, false)
	seedEquipment(t, group.ID, buildingB.ID, "Cow Cart B", false, false)

	date := formatDate(futureDate(14))

	t.Run("counts only units stocked in the building", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00&buildingId=%d",
			group.ID, date, buildingA.ID,
		), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[map[string]int](t, w)
		if got["available"] != 1 {
			t.Errorf("available = %d, want 1", got["available"])
		}
	})

	t.Run("rejects a missing buildingId (every unit has a permanent home)", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00", group.ID, date,
		), nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects an invalid buildingId", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00&buildingId=nope",
			group.ID, date,
		), nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type scheduleEnvelope struct {
	Date string          `json:"date"`
	Data []scheduleEntry `json:"data"`
}

func TestEquipmentGroupSchedule(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	group := seedGroup(t, "Cow Cart", false, false)
	unitA := seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart B", false, false)

	// A weekly booking that runs for 3 weeks, and a same-day booking for the other unit.
	firstDay := futureDate(7)
	morning := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Bio Lecture",
		firstDate:  firstDay,
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
		weeks:      3,
		room:       "204",
	})
	afternoon := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Chem Lab",
		firstDate:  firstDay,
		startTime:  clockTime(13, 0),
		endTime:    clockTime(15, 0),
		room:       "112b",
	})
	for _, requestID := range []uint{morning.ID, afternoon.ID} {
		if _, err := services.CreateRequestedEquipment(testDB, requestID, group.ID); err != nil {
			t.Fatalf("seed booking for request %d: %v", requestID, err)
		}
	}

	t.Run("lists bookings for the day, earliest first", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/schedule?date=%s", group.ID, formatDate(firstDay),
		), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		got := decodeJSON[scheduleEnvelope](t, w)
		if len(got.Data) != 2 {
			t.Fatalf("expected 2 entries, got %d: %+v", len(got.Data), got.Data)
		}
		first, second := got.Data[0], got.Data[1]
		if first.StartTime != "09:00" || first.EndTime != "10:00" {
			t.Errorf("first entry times = %s-%s, want 09:00-10:00", first.StartTime, first.EndTime)
		}
		if first.Room != "204" || first.BuildingName != "Anna Whitten Hall" {
			t.Errorf("first entry location = %s %s, want Anna Whitten Hall 204", first.BuildingName, first.Room)
		}
		if first.EquipmentID != unitA.ID || first.EquipmentName != "Cow Cart A" {
			t.Errorf("first entry unit = %d/%s, want %d/Cow Cart A", first.EquipmentID, first.EquipmentName, unitA.ID)
		}
		if first.RequestName != "Bio Lecture" {
			t.Errorf("first entry request = %q, want Bio Lecture", first.RequestName)
		}
		if second.StartTime != "13:00" || second.Room != "112b" {
			t.Errorf("second entry = %s in %s, want 13:00 in 112b", second.StartTime, second.Room)
		}
	})

	t.Run("includes a weekly booking on a later occurrence", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/schedule?date=%s", group.ID, formatDate(firstDay.AddDate(0, 0, 14)),
		), nil)
		got := decodeJSON[scheduleEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].RequestName != "Bio Lecture" {
			t.Fatalf("expected only the 3-week Bio Lecture booking, got %+v", got.Data)
		}
	})

	t.Run("is empty past the end of the weekly run", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/schedule?date=%s", group.ID, formatDate(firstDay.AddDate(0, 0, 21)),
		), nil)
		got := decodeJSON[scheduleEnvelope](t, w)
		if len(got.Data) != 0 {
			t.Fatalf("expected no entries, got %+v", got.Data)
		}
	})

	t.Run("is empty on a different weekday", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/schedule?date=%s", group.ID, formatDate(firstDay.AddDate(0, 0, 1)),
		), nil)
		got := decodeJSON[scheduleEnvelope](t, w)
		if len(got.Data) != 0 {
			t.Fatalf("expected no entries on a different weekday, got %+v", got.Data)
		}
	})

	t.Run("404s for a nonexistent group", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment-groups/999999/schedule?date="+formatDate(firstDay), nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a missing or invalid date", func(t *testing.T) {
		for _, target := range []string{
			fmt.Sprintf("/api/equipment-groups/%d/schedule", group.ID),
			fmt.Sprintf("/api/equipment-groups/%d/schedule?date=not-a-date", group.ID),
		} {
			w := doRequest(r, http.MethodGet, target, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %s, got %d: %s", target, w.Code, w.Body.String())
			}
		}
	})
}
