package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

// TestMain spins up a single real Postgres container for the whole package,
// so handler tests exercise the actual SQL (ILIKE, ORDER BY, LIMIT/OFFSET)
// instead of a mocked or dialect-mismatched substitute.
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
	gin.SetMode(gin.TestMode)

	return m.Run(), nil
}

// resetTables truncates every table shared by this package's handler tests, in FK-safe order.
func resetTables(t *testing.T) {
	t.Helper()
	if err := testDB.Exec(
		"TRUNCATE TABLE requested_equipments, requests, equipment, equipment_groups, buildings RESTART IDENTITY CASCADE",
	).Error; err != nil {
		t.Fatalf("failed to reset tables: %v", err)
	}
}

// newTestRouter wires every handler onto one router, mirroring how main.go mounts them, so route
// registration (shared ":id" param names, etc.) is exercised the same way it runs in production.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	resetTables(t)

	r := gin.New()
	api := r.Group("/api")
	NewBuildingHandler(testDB).RegisterRoutes(api)
	NewEquipmentGroupHandler(testDB).RegisterRoutes(api)
	NewEquipmentHandler(testDB).RegisterRoutes(api)
	NewRequestHandler(testDB).RegisterRoutes(api)
	NewRequestedEquipmentHandler(testDB).RegisterRoutes(api)
	return r
}

func seedBuilding(t *testing.T, name string, archived bool) models.Building {
	t.Helper()
	desc := "description for " + name
	addr := "address for " + name
	b := models.Building{Name: name, Archived: archived, Description: &desc, Address: &addr}
	if err := testDB.Create(&b).Error; err != nil {
		t.Fatalf("failed to seed building %q: %v", name, err)
	}
	return b
}

func doRequest(r *gin.Engine, method, target string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		buf, _ := json.Marshal(body)
		reader = bytes.NewReader(buf)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response body %q: %v", w.Body.String(), err)
	}
	return out
}

func keySet(t *testing.T, m map[string]any) map[string]bool {
	t.Helper()
	set := make(map[string]bool, len(m))
	for k := range m {
		set[k] = true
	}
	return set
}

func assertExactKeys(t *testing.T, label string, got map[string]any, want ...string) {
	t.Helper()
	gotSet := keySet(t, got)
	wantSet := make(map[string]bool, len(want))
	for _, k := range want {
		wantSet[k] = true
		if !gotSet[k] {
			t.Errorf("%s: missing expected key %q (got keys: %v)", label, k, gotSet)
		}
	}
	for k := range gotSet {
		if !wantSet[k] {
			t.Errorf("%s: unexpected extra key %q the frontend does not expect (got keys: %v)", label, k, gotSet)
		}
	}
}

// buildingContractKeys is the exact set of JSON fields the frontend Building
// type binds to. If this test fails, the frontend's src/types/Building.ts and
// src/services/buildingsApi.ts need to be updated to match.
var buildingContractKeys = []string{"id", "name", "archived", "description", "address", "createdAt", "updatedAt"}

func TestBuildingResponseContract(t *testing.T) {
	r := newTestRouter(t)

	t.Run("create response matches the frontend contract", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/buildings", buildingInput{Name: "Contract Hall"})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeJSON[map[string]any](t, w)
		assertExactKeys(t, "create response", body, buildingContractKeys...)
	})

	t.Run("list response wraps data/meta with the expected shape", func(t *testing.T) {
		seedBuilding(t, "Listed Hall", false)
		w := doRequest(r, http.MethodGet, "/api/buildings", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		envelope := decodeJSON[map[string]any](t, w)
		assertExactKeys(t, "list envelope", envelope, "data", "meta")

		meta, ok := envelope["meta"].(map[string]any)
		if !ok {
			t.Fatalf("expected meta to be an object, got %T", envelope["meta"])
		}
		assertExactKeys(t, "list meta", meta, "page", "pageSize", "totalItems", "totalPages")

		data, ok := envelope["data"].([]any)
		if !ok {
			t.Fatalf("expected data to be an array, got %T", envelope["data"])
		}
		if len(data) == 0 {
			t.Fatal("expected at least one building in data")
		}
		first, ok := data[0].(map[string]any)
		if !ok {
			t.Fatalf("expected data[0] to be an object, got %T", data[0])
		}
		assertExactKeys(t, "list item", first, buildingContractKeys...)
	})
}

func TestBuildingCreate(t *testing.T) {
	r := newTestRouter(t)

	t.Run("creates a building", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/buildings", buildingInput{Name: "Anna Whitten Hall"})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Building](t, w)
		if got.ID == 0 {
			t.Error("expected a generated id")
		}
		if got.Name != "Anna Whitten Hall" {
			t.Errorf("Name = %q, want %q", got.Name, "Anna Whitten Hall")
		}
		if got.Archived {
			t.Error("expected Archived to default false")
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/buildings", buildingInput{Name: "   "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejects a missing name", func(t *testing.T) {
		w := doRequest(r, http.MethodPost, "/api/buildings", map[string]any{})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestBuildingGet(t *testing.T) {
	r := newTestRouter(t)
	b := seedBuilding(t, "Groves", false)

	t.Run("returns an existing building", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, fmt.Sprintf("/api/buildings/%d", b.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Building](t, w)
		if got.ID != b.ID {
			t.Errorf("ID = %d, want %d", got.ID, b.ID)
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("400s for a non-numeric id", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings/not-a-number", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestBuildingUpdate(t *testing.T) {
	r := newTestRouter(t)
	b := seedBuilding(t, "Texas Township Campus", false)

	t.Run("updates an existing building", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/buildings/%d", b.ID), buildingInput{
			Name:     "Texas Township Campus (Renamed)",
			Archived: true,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[models.Building](t, w)
		if got.Name != "Texas Township Campus (Renamed)" {
			t.Errorf("Name = %q, want renamed value", got.Name)
		}
		if !got.Archived {
			t.Error("expected Archived to be true")
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/buildings/%d", b.ID), buildingInput{Name: ""})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodPut, "/api/buildings/999999", buildingInput{Name: "Nope"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestBuildingDelete(t *testing.T) {
	r := newTestRouter(t)
	b := seedBuilding(t, "Kalamazoo Valley Museum", false)

	t.Run("archives instead of hard-deleting", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, fmt.Sprintf("/api/buildings/%d", b.ID), nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		var reloaded models.Building
		if err := testDB.First(&reloaded, b.ID).Error; err != nil {
			t.Fatalf("expected building to still exist after delete: %v", err)
		}
		if !reloaded.Archived {
			t.Error("expected building to be archived, not removed")
		}
	})

	t.Run("404s for a nonexistent id", func(t *testing.T) {
		w := doRequest(r, http.MethodDelete, "/api/buildings/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

type buildingListEnvelope struct {
	Data []models.Building `json:"data"`
	Meta paginationMeta    `json:"meta"`
}

func TestBuildingListPagination(t *testing.T) {
	r := newTestRouter(t)
	names := []string{"Alpha Hall", "Bravo Hall", "Charlie Hall", "Delta Hall", "Echo Hall"}
	for _, n := range names {
		seedBuilding(t, n, false)
	}

	t.Run("defaults to page 1 with pageSize 20", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[buildingListEnvelope](t, w)
		if got.Meta.Page != 1 || got.Meta.PageSize != 20 {
			t.Errorf("Meta = %+v, want page 1 pageSize 20", got.Meta)
		}
		if got.Meta.TotalItems != int64(len(names)) {
			t.Errorf("TotalItems = %d, want %d", got.Meta.TotalItems, len(names))
		}
		if len(got.Data) != len(names) {
			t.Errorf("len(Data) = %d, want %d", len(got.Data), len(names))
		}
	})

	t.Run("paginates and sorts by name", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?page=2&pageSize=2&sort=name&order=asc", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		got := decodeJSON[buildingListEnvelope](t, w)
		if got.Meta.Page != 2 || got.Meta.PageSize != 2 {
			t.Errorf("Meta = %+v, want page 2 pageSize 2", got.Meta)
		}
		if got.Meta.TotalPages != 3 {
			t.Errorf("TotalPages = %d, want 3", got.Meta.TotalPages)
		}
		if len(got.Data) != 2 {
			t.Fatalf("len(Data) = %d, want 2", len(got.Data))
		}
		// Sorted ascending by name, page 2 of size 2 -> items 3 and 4 (Charlie, Delta).
		if got.Data[0].Name != "Charlie Hall" || got.Data[1].Name != "Delta Hall" {
			t.Errorf("Data = [%q, %q], want [Charlie Hall, Delta Hall]", got.Data[0].Name, got.Data[1].Name)
		}
	})

	t.Run("sorts descending", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?sort=name&order=desc&pageSize=1", nil)
		got := decodeJSON[buildingListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Echo Hall" {
			t.Errorf("expected first result to be Echo Hall, got %+v", got.Data)
		}
	})

	t.Run("caps pageSize at the maximum", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?pageSize=500", nil)
		got := decodeJSON[buildingListEnvelope](t, w)
		if got.Meta.PageSize != maxPageSize {
			t.Errorf("PageSize = %d, want %d", got.Meta.PageSize, maxPageSize)
		}
	})

	invalidCases := []string{
		"/api/buildings?page=0",
		"/api/buildings?page=abc",
		"/api/buildings?pageSize=0",
		"/api/buildings?pageSize=abc",
		"/api/buildings?sort=" + url.QueryEscape("id; DROP TABLE buildings"),
		"/api/buildings?order=sideways",
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

func TestBuildingListSearchAndFilter(t *testing.T) {
	r := newTestRouter(t)
	seedBuilding(t, "Anna Whitten Hall", false)
	seedBuilding(t, "Culinary Allied Health", false)
	seedBuilding(t, "Old Gym", true)

	t.Run("searches case-insensitively by name", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?q=whitten", nil)
		got := decodeJSON[buildingListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Anna Whitten Hall" {
			t.Errorf("expected only Anna Whitten Hall, got %+v", got.Data)
		}
	})

	t.Run("filters to archived only", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?archived=true", nil)
		got := decodeJSON[buildingListEnvelope](t, w)
		if len(got.Data) != 1 || got.Data[0].Name != "Old Gym" {
			t.Errorf("expected only Old Gym, got %+v", got.Data)
		}
	})

	t.Run("filters to non-archived only", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?archived=false", nil)
		got := decodeJSON[buildingListEnvelope](t, w)
		if len(got.Data) != 2 {
			t.Errorf("expected 2 non-archived buildings, got %d: %+v", len(got.Data), got.Data)
		}
	})

	t.Run("rejects a non-boolean archived value", func(t *testing.T) {
		w := doRequest(r, http.MethodGet, "/api/buildings?archived=maybe", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}
