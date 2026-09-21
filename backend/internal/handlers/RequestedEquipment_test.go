package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
)

func TestRequestedEquipmentResponseContract(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID, name: "Bio Lecture",
		firstDate: futureDate(10), startTime: clockTime(9, 0), endTime: clockTime(10, 0),
	})
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
		requestedEquipmentInput{GroupID: group.ID})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON[map[string]any](t, w)
	assertExactKeys(t, "create response", body, "id", "groupId", "group", "equipmentId", "equipment", "requestId")
	if _, present := body["request"]; present {
		t.Errorf(`expected "request" to be omitted (not preloaded here), got %v`, body["request"])
	}
}

func TestRequestedEquipmentCreate(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID, name: "Bio Lecture",
		firstDate: futureDate(10), startTime: clockTime(9, 0), endTime: clockTime(10, 0),
	})
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart B", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	t.Run("assigns the alphabetically-first available unit", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: group.ID})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.RequestedEquipment](t, w)
		if got.Equipment == nil || got.Equipment.Name != "Cow Cart A" {
			t.Errorf("expected Cow Cart A assigned, got %+v", got.Equipment)
		}
	})

	t.Run("a second call assigns the other unit", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: group.ID})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.RequestedEquipment](t, w)
		if got.Equipment == nil || got.Equipment.Name != "Cow Cart B" {
			t.Errorf("expected Cow Cart B assigned, got %+v", got.Equipment)
		}
	})

	t.Run("a third call has nothing left", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: group.ID})
		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s for a nonexistent request", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/requests/999999/equipment", requestedEquipmentInput{GroupID: group.ID})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s for a nonexistent group", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: 999999})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("409s for a disabled group", func(t *testing.T) {
		disabledGroup := seedGroup(t, "Disabled Group", true, false)
		seedEquipment(t, disabledGroup.ID, building.ID, "Unit A", false, false)
		w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: disabledGroup.ID})
		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestRequestedEquipmentList(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID, name: "Bio Lecture",
		firstDate: futureDate(10), startTime: clockTime(9, 0), endTime: clockTime(10, 0),
	})
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	if w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
		requestedEquipmentInput{GroupID: group.ID}); w.Code != http.StatusCreated {
		t.Fatalf("failed to seed an assignment: %s", w.Body.String())
	}

	w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/requests/%d/equipment", request.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	got := decodeJSON[[]models.RequestedEquipment](t, w)
	if len(got) != 1 || got[0].Equipment == nil || got[0].Equipment.Name != "Cow Cart A" {
		t.Errorf("expected 1 assignment for Cow Cart A, got %+v", got)
	}
}

func TestRequestedEquipmentDelete(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID, name: "Bio Lecture",
		firstDate: futureDate(10), startTime: clockTime(9, 0), endTime: clockTime(10, 0),
	})
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)

	assignResp := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
		requestedEquipmentInput{GroupID: group.ID})
	if assignResp.Code != http.StatusCreated {
		t.Fatalf("failed to seed an assignment: %s", assignResp.Body.String())
	}
	assignment := decodeJSON[models.RequestedEquipment](t, assignResp)

	t.Run("unassigning frees the unit for reuse", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete,
			fmt.Sprintf("/api/requests/%d/equipment/%d", request.ID, assignment.ID), nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		// the same request can now be assigned Cow Cart A again since the prior booking is gone
		reassign := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: group.ID})
		if reassign.Code != http.StatusCreated {
			t.Fatalf("expected the freed unit to be assignable again, got %d: %s", reassign.Code, reassign.Body.String())
		}
	})

	t.Run("404s for a nonexistent assignment id", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, fmt.Sprintf("/api/requests/%d/equipment/999999", request.ID), nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s when the assignment belongs to a different request", func(t *testing.T) {
		// Isolated fixtures: `request` already holds the group's only unit from the subtest
		// above, so reuse would spuriously conflict with itself rather than testing the mismatch.
		otherGroup := seedGroup(t, "Other Cart", false, false)
		seedEquipment(t, otherGroup.ID, building.ID, "Other Cart A", false, false)
		otherRequest := seedRequest(t, requestOpts{
			buildingID: building.ID, name: "Other Lecture",
			firstDate: futureDate(20), startTime: clockTime(9, 0), endTime: clockTime(10, 0),
		})
		reassign := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: otherGroup.ID})
		if reassign.Code != http.StatusCreated {
			t.Fatalf("failed to seed an assignment for the mismatch test: %s", reassign.Body.String())
		}
		other := decodeJSON[models.RequestedEquipment](t, reassign)

		w := doRequest(r, http.MethodDelete,
			fmt.Sprintf("/api/requests/%d/equipment/%d", otherRequest.ID, other.ID), nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
