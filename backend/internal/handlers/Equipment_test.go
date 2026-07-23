package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ofgrenudo/OmniAv/internal/models"
)

func TestEquipmentResponseContract(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)

	w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Cow Cart A", GroupID: group.ID})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON[map[string]any](t, w)
	assertExactKeys(t, "create response", body,
		"id", "name", "description", "disabled", "archived", "groupId", "createdAt", "updatedAt")
}

func TestEquipmentCreate(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)

	t.Run("creates equipment under a group", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Cow Cart A", GroupID: group.ID})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Equipment](t, w)
		if got.GroupID != group.ID {
			t.Errorf("GroupID = %d, want %d", got.GroupID, group.ID)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: " ", GroupID: group.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a missing groupId", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Orphan Unit"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a nonexistent groupId", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/equipment", equipmentInput{Name: "Orphan Unit", GroupID: 999999})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentGet(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	e := seedEquipment(t, group.ID, "Cow Cart A", false, false)

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
	e := seedEquipment(t, group.ID, "Cow Cart A", false, false)

	t.Run("updates and can move to another group", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/equipment/%d", e.ID), equipmentInput{
			Name:    "Projector 1",
			GroupID: otherGroup.ID,
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
		w := doRequest(r, http.MethodPut, "/api/equipment/999999", equipmentInput{Name: "Nope", GroupID: group.ID})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestEquipmentDelete(t *testing.T) {
	r := newTestRouter(t)
	group := seedGroup(t, "Cow Cart", false, false)
	e := seedEquipment(t, group.ID, "Cow Cart A", false, false)

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
	seedEquipment(t, cowCart.ID, "Cow Cart A", false, false)
	seedEquipment(t, cowCart.ID, "Cow Cart B (disabled)", true, false)
	seedEquipment(t, projector.ID, "Projector 1", false, false)

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
}
