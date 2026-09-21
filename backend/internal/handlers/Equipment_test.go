package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/ofgrenudo/OmniAv/internal/services"
)

func TestEquipmentResponseContract(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	building := seedBuilding(t, "Main Hall", false)

	w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Cow Cart A", GroupID: group.ID, BuildingID: building.ID})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON[map[string]any](t, w)
	assertExactKeys(t, "create response", body,
		"id", "name", "description", "disabled", "archived", "groupId", "buildingId", "createdAt", "updatedAt")
}

func TestEquipmentCreate(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	building := seedBuilding(t, "Main Hall", false)

	t.Run("creates equipment under a group in a building", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Cow Cart A", GroupID: group.ID, BuildingID: building.ID})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Equipment](t, w)
		if got.GroupID != group.ID {
			t.Errorf("GroupID = %d, want %d", got.GroupID, group.ID)
		}
		if got.BuildingID != building.ID {
			t.Errorf("BuildingID = %d, want %d", got.BuildingID, building.ID)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: " ", GroupID: group.ID, BuildingID: building.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a missing groupId", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Orphan Unit", BuildingID: building.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a nonexistent groupId", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Orphan Unit", GroupID: 999999, BuildingID: building.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a missing buildingId (every unit must have a permanent home)", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Homeless Unit", GroupID: group.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a nonexistent buildingId", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Orphan Unit", GroupID: group.ID, BuildingID: 999999})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGet(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	building := seedBuilding(t, "Main Hall", false)
	e := seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	t.Run("returns existing equipment without a nested group by default", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment/%d", e.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeJSON[map[string]any](t, w)
		if _, present := body["group"]; present {
			t.Errorf(`expected "group" to be omitted when not preloaded, got %v`, body["group"])
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentUpdate(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	otherGroup := seedGroup(t, "Projector Kit", false, false)
	building := seedBuilding(t, "Main Hall", false)
	e := seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	t.Run("updates and can move to another group", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment/%d", e.ID), equipmentInput{
			Name:       "Projector 1",
			GroupID:    otherGroup.ID,
			BuildingID: building.ID,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Equipment](t, w)
		if got.GroupID != otherGroup.ID {
			t.Errorf("GroupID = %d, want %d", got.GroupID, otherGroup.ID)
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, "/api/equipment/999999", equipmentInput{Name: "Nope", GroupID: group.ID, BuildingID: building.ID})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("moves a unit to a different building, replacing the prior one", func(t *testing.T) {
		buildingA := seedBuilding(t, "Building A", false)
		buildingB := seedBuilding(t, "Building B", false)

		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment/%d", e.ID), equipmentInput{
			Name:       e.Name,
			GroupID:    group.ID,
			BuildingID: buildingA.ID,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		w = doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment/%d", e.ID), equipmentInput{
			Name:       e.Name,
			GroupID:    group.ID,
			BuildingID: buildingB.ID,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Equipment](t, w)
		if got.BuildingID != buildingB.ID {
			t.Errorf("BuildingID = %d, want %d (only the most recent building)", got.BuildingID, buildingB.ID)
		}
	})

	t.Run("rejects a nonexistent buildingId", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment/%d", e.ID), equipmentInput{
			Name:       e.Name,
			GroupID:    group.ID,
			BuildingID: 999999,
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentDelete(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	building := seedBuilding(t, "Main Hall", false)
	e := seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	t.Run("archives instead of hard-deleting", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, fmt.Sprintf("/api/equipment/%d", e.ID), nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		var reloaded models.Equipment
		if err := testDB.First(&reloaded, e.ID).Error; err != nil {
			t.Fatalf("expected equipment to still exist: %v", err)
		}
		if !reloaded.Archived {
			t.Error("expected equipment to be archived")
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, "/api/equipment/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type equipmentListEnvelope struct {
	Data []models.Equipment `json:"data"`
	Meta paginationMeta     `json:"meta"`
}

func TestEquipmentListFilters(t *testing.T) {
	r := newTestRouter(t)
	cowCart := seedGroup(t, "Cow Cart", false, false)
	projector := seedGroup(t, "Projector Kit", false, false)
	building := seedBuilding(t, "Main Hall", false)
	seedEquipment(t, cowCart.ID, building.ID, "Cow Cart A", false, false)
	seedEquipment(t, cowCart.ID, building.ID, "Cow Cart B (disabled)", true, false)
	seedEquipment(t, projector.ID, building.ID, "Projector 1", false, false)

	t.Run("filters by groupId", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment?groupId=%d", cowCart.ID), nil)
		got := decodeJSON[equipmentListEnvelope](t, w)
		if len(got.Data) != 2 {
			t.Errorf("expected 2 units in Cow Cart group, got %d", len(got.Data))
		}
	})

	t.Run("filters by disabled", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment?groupId=%d&disabled=false", cowCart.ID), nil)
		got := decodeJSON[equipmentListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Cow Cart A" {
			t.Errorf("expected only Cow Cart A, got %+v", got.Data)
		}
	})

	t.Run("searches by name", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment?q=projector", nil)
		got := decodeJSON[equipmentListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Projector 1" {
			t.Errorf("expected only Projector 1, got %+v", got.Data)
		}
	})

	t.Run("rejects an invalid groupId filter", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment?groupId=nope", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("filters by buildingId", func(t *testing.T) {
		library := seedBuilding(t, "Library", false)
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{
			Name:       "Cow Cart In Library",
			GroupID:    cowCart.ID,
			BuildingID: library.ID,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		w = doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment?buildingId=%d", library.ID), nil)
		got := decodeJSON[equipmentListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Cow Cart In Library" {
			t.Errorf("expected only Cow Cart In Library, got %+v", got.Data)
		}
	})

	t.Run("rejects an invalid buildingId filter", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment?buildingId=nope", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type bookingsEnvelope struct {
	EquipmentID   uint           `json:"equipmentId"`
	EquipmentName string         `json:"equipmentName"`
	Data          []bookingEntry `json:"data"`
}

func TestEquipmentBookings(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	group := seedGroup(t, "Cow Cart", false, false)
	unit := seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	// Two bookings for the same unit, seeded out of chronological order.
	later := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Chem Lab",
		firstDate:  futureDate(21),
		startTime:  clockTime(13, 0),
		endTime:    clockTime(15, 0),
		room:       "112b",
	})
	earlier := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Bio Lecture",
		firstDate:  futureDate(14),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
		room:       "204",
	})
	for _, requestID := range []uint{later.ID, earlier.ID} {
		if _, err := services.CreateRequestedEquipment(testDB, requestID, group.ID); err != nil {
			t.Fatalf("seed booking for request %d: %v", requestID, err)
		}
	}

	t.Run("returns the unit's bookings oldest first", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment/%d/bookings", unit.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		got := decodeJSON[bookingsEnvelope](t, w)
		if got.EquipmentName != "Cow Cart A" {
			t.Errorf("equipmentName = %q, want Cow Cart A", got.EquipmentName)
		}
		if len(got.Data) != 2 {
			t.Fatalf("expected 2 bookings, got %d: %+v", len(got.Data), got.Data)
		}
		if got.Data[0].RequestName != "Bio Lecture" || got.Data[1].RequestName != "Chem Lab" {
			t.Fatalf("expected Bio Lecture then Chem Lab, got %q then %q", got.Data[0].RequestName, got.Data[1].RequestName)
		}
		first := got.Data[0]
		if first.Room != "204" || first.BuildingName != "Anna Whitten Hall" {
			t.Errorf("first booking location = %s %s, want Anna Whitten Hall 204", first.BuildingName, first.Room)
		}
		if first.StartTime != "09:00" || first.EndTime != "10:00" {
			t.Errorf("first booking times = %s-%s, want 09:00-10:00", first.StartTime, first.EndTime)
		}
		if first.FirstDateNeeded != formatDate(futureDate(14)) {
			t.Errorf("first booking date = %s, want %s", first.FirstDateNeeded, formatDate(futureDate(14)))
		}
		if first.RequestID != earlier.ID {
			t.Errorf("first booking requestId = %d, want %d", first.RequestID, earlier.ID)
		}
	})

	t.Run("returns an empty list for a unit that has never been booked", func(t *testing.T) {
		idle := seedEquipment(t, group.ID, building.ID, "Cow Cart Z", false, false)
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/equipment/%d/bookings", idle.ID), nil)
		got := decodeJSON[bookingsEnvelope](t, w)
		if len(got.Data) != 0 {
			t.Fatalf("expected no bookings, got %+v", got.Data)
		}
	})

	t.Run("404s for a nonexistent unit", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/equipment/999999/bookings", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
