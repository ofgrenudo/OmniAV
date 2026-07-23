package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
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
	group := seedGroup(t, "Cow Cart", false, false)
	seedEquipment(t, group.ID, "Cow Cart A", false, false)
	seedEquipment(t, group.ID, "Cow Cart B", false, false)

	date := formatDate(futureDate(14))

	t.Run("reports the full count with no bookings", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf(
			"/api/equipment-groups/%d/availability?firstDate=%s&startTime=09:00&endTime=10:00&weeks=1", group.ID, date,
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
			"/api/equipment-groups/999999/availability?firstDate=%s&startTime=09:00&endTime=10:00", date,
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
