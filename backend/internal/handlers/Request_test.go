package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
)

func validRequestInput(buildingID uint, name string, daysFromNow int) requestInput {
	return requestInput{
		Name:            name,
		FirstDateNeeded: formatDate(futureDate(daysFromNow)),
		StartTime:       "09:00",
		EndTime:         "10:00",
		NumberOfWeeks:   1,
		BuildingID:      buildingID,
		Room:            "101",
	}
}

func TestRequestResponseContract(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)

	w := doRequest(r, http.MethodPost, "/api/requests", validRequestInput(building.ID, "Bio Lecture", 10))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON[map[string]any](t, w)
	assertExactKeys(t, "create response", body,
		"id", "name", "firstDateNeeded", "startTime", "endTime", "numberOfWeeks", "daysOfWeek",
		"buildingId", "room", "building", "comments", "attachmentId", "createdAt", "updatedAt")
}

func TestRequestCreate(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)

	t.Run("creates a request and derives DaysOfWeek from the date", func(t *testing.T) {
		date := futureDate(10)
		w := doRequest(r, http.MethodPost, "/api/requests", validRequestInput(building.ID, "Bio Lecture", 10))
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Request](t, w)
		want := models.FromGoWeekday(date.Weekday())
		if got.DaysOfWeek != want {
			t.Errorf("DaysOfWeek = %v, want %v (derived from %v)", got.DaysOfWeek, want, date.Weekday())
		}
		if got.Building == nil || got.Building.ID != building.ID {
			t.Errorf("expected Building to be populated, got %+v", got.Building)
		}
	})

	t.Run("upserts a BuildingRoom instead of duplicating it", func(t *testing.T) {
		input := validRequestInput(building.ID, "Second booking, same room", 11)
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		var count int64
		if err := testDB.Model(&models.BuildingRoom{}).
			Where("building_id = ? AND room = ?", building.ID, "101").
			Count(&count).Error; err != nil {
			t.Fatalf("failed to count building_rooms: %v", err)
		}
		if count != 1 {
			t.Errorf("expected exactly 1 building_rooms row for building %d room 101, got %d", building.ID, count)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		input := validRequestInput(building.ID, "  ", 10)
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a blank room", func(t *testing.T) {
		input := validRequestInput(building.ID, "Whatever", 10)
		input.Room = "  "
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("accepts rooms made of digits and the wing letters a-c", func(t *testing.T) {
		for i, room := range []string{"204", "123a", "12C", "3b"} {
			input := validRequestInput(building.ID, "Room "+room, 10+i)
			input.Room = room
			w := doRequest(r, http.MethodPost, "/api/requests", input)
			if w.Code != http.StatusCreated {
				t.Errorf("expected 201 for room %q, got %d: %s", room, w.Code, w.Body.String())
			}
		}
	})

	t.Run("rejects a room with characters outside 0-9 and a-c", func(t *testing.T) {
		for _, room := range []string{"204d", "lobby", "12 4", "204-A", "204.1", "Gym"} {
			input := validRequestInput(building.ID, "Whatever", 10)
			input.Room = room
			w := doRequest(r, http.MethodPost, "/api/requests", input)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for room %q, got %d: %s", room, w.Code, w.Body.String())
			}
		}
	})

	t.Run("accepts times at the edges of service hours", func(t *testing.T) {
		input := validRequestInput(building.ID, "All day", 20)
		input.StartTime, input.EndTime = "07:30", "22:00"
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects times outside service hours", func(t *testing.T) {
		cases := [][2]string{
			{"07:00", "09:00"}, // starts before staff arrive
			{"07:20", "08:00"},
			{"21:00", "22:30"}, // ends after staff leave
			{"22:30", "23:00"},
		}
		for _, tc := range cases {
			input := validRequestInput(building.ID, "Whatever", 10)
			input.StartTime, input.EndTime = tc[0], tc[1]
			w := doRequest(r, http.MethodPost, "/api/requests", input)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %s-%s, got %d: %s", tc[0], tc[1], w.Code, w.Body.String())
			}
		}
	})

	t.Run("rejects an invalid date", func(t *testing.T) {
		input := validRequestInput(building.ID, "Whatever", 10)
		input.FirstDateNeeded = "not-a-date"
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects endTime not after startTime", func(t *testing.T) {
		input := validRequestInput(building.ID, "Whatever", 10)
		input.StartTime, input.EndTime = "10:00", "09:00"
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects fewer than 1 week", func(t *testing.T) {
		input := validRequestInput(building.ID, "Whatever", 10)
		input.NumberOfWeeks = 0
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects less than 24 hours advance notice", func(t *testing.T) {
		input := validRequestInput(building.ID, "Too soon", 0) // today
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a nonexistent building", func(t *testing.T) {
		input := validRequestInput(999999, "Whatever", 10)
		w := doRequest(r, http.MethodPost, "/api/requests", input)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestRequestGet(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Bio Lecture",
		firstDate:  futureDate(10),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})

	t.Run("returns the request with its building and assignments preloaded", func(t *testing.T) {
		group := seedGroup(t, "Cow Cart", false, false)
		seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)
		if w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
			requestedEquipmentInput{GroupID: group.ID}); w.Code != http.StatusCreated {
			t.Fatalf("failed to seed an assignment: %s", w.Body.String())
		}

		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/requests/%d", request.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Request](t, w)
		if got.Building == nil || got.Building.ID != building.ID {
			t.Errorf("expected Building preloaded, got %+v", got.Building)
		}
		if len(got.RequestedEquipment) != 1 {
			t.Fatalf("expected 1 requested equipment, got %d", len(got.RequestedEquipment))
		}
		if got.RequestedEquipment[0].Equipment == nil || got.RequestedEquipment[0].Equipment.Name != "Cow Cart A" {
			t.Errorf("expected the assigned equipment preloaded, got %+v", got.RequestedEquipment[0].Equipment)
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/requests/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("400s for a non-numeric id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/requests/nope", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type requestListEnvelope struct {
	Data []models.Request `json:"data"`
	Meta paginationMeta   `json:"meta"`
}

func TestRequestListFilters(t *testing.T) {
	r := newTestRouter(t)
	buildingA := seedBuilding(t, "Anna Whitten Hall", false)
	buildingB := seedBuilding(t, "Groves", false)
	seedRequest(t, requestOpts{buildingID: buildingA.ID, name: "Bio Lecture", firstDate: futureDate(10), startTime: clockTime(9, 0), endTime: clockTime(10, 0)})
	seedRequest(t, requestOpts{buildingID: buildingB.ID, name: "Chem Lab", firstDate: futureDate(11), startTime: clockTime(9, 0), endTime: clockTime(10, 0)})

	t.Run("filters by buildingId", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/requests?buildingId=%d", buildingA.ID), nil)
		got := decodeJSON[requestListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Bio Lecture" {
			t.Errorf("expected only Bio Lecture, got %+v", got.Data)
		}
	})

	t.Run("searches by name", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/requests?q=chem", nil)
		got := decodeJSON[requestListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Chem Lab" {
			t.Errorf("expected only Chem Lab, got %+v", got.Data)
		}
	})
}

func TestRequestDelete(t *testing.T) {
	r := newTestRouter(t)
	building := seedBuilding(t, "Anna Whitten Hall", false)
	request := seedRequest(t, requestOpts{
		buildingID: building.ID,
		name:       "Bio Lecture",
		firstDate:  futureDate(10),
		startTime:  clockTime(9, 0),
		endTime:    clockTime(10, 0),
	})
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, building.ID, "Cow Cart A", false, false)
	if w := doRequest(r, http.MethodPost, fmt.Sprintf("/api/requests/%d/equipment", request.ID),
		requestedEquipmentInput{GroupID: group.ID}); w.Code != http.StatusCreated {
		t.Fatalf("failed to seed an assignment: %s", w.Body.String())
	}

	t.Run("cascades to remove requested equipment", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, fmt.Sprintf("/api/requests/%d", request.ID), nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		var requestCount, assignmentCount int64
		testDB.Model(&models.Request{}).Where("id = ?", request.ID).Count(&requestCount)
		testDB.Model(&models.RequestedEquipment{}).Where("request_id = ?", request.ID).Count(&assignmentCount)
		if requestCount != 0 {
			t.Errorf("expected the request to be gone, count = %d", requestCount)
		}
		if assignmentCount != 0 {
			t.Errorf("expected its requested_equipments rows to be gone, count = %d", assignmentCount)
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, "/api/requests/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
