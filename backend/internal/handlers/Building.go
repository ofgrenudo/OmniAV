package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"gorm.io/gorm"
)

type BuildingHandler struct {
	DB *gorm.DB
}

func NewBuildingHandler(db *gorm.DB) *BuildingHandler {
	return &BuildingHandler{DB: db}
}

func (h *BuildingHandler) RegisterRoutes(rg *gin.RouterGroup) {
	buildings := rg.Group("/buildings")
	{
		buildings.GET("", h.List)
		buildings.GET("/:id", h.Get)
		buildings.POST("", h.Create)
		buildings.PUT("/:id", h.Update)
		buildings.DELETE("/:id", h.Delete)
	}
}

type buildingInput struct {
	Name        string `binding:"required"`
	Archived    bool
	Description *string
	Address     *string
}

// buildingSortColumns whitelists the columns clients may sort by, mapping
// the query value to the actual DB column to avoid SQL injection via ORDER BY.
var buildingSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"archived":   "archived",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func (h *BuildingHandler) List(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	sortColumn, sortOrder, err := parseSort(c, buildingSortColumns, "id")
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	query := h.DB.Model(&models.Building{})

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if archived, present, err := parseOptionalBool(c, "archived"); err != nil {
		badRequest(c, err.Error())
		return
	} else if present {
		query = query.Where("archived = ?", archived)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		internalError(c, err)
		return
	}

	var buildings []models.Building
	offset := (page - 1) * pageSize
	if err := query.
		Order(sortColumn + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&buildings).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, listResponse{
		Data: buildings,
		Meta: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages(total, pageSize),
		},
	})
}

func (h *BuildingHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid building id")
		return
	}

	var building models.Building
	if err := h.DB.First(&building, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "building not found")
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, building)
}

func (h *BuildingHandler) Create(c *gin.Context) {
	var input buildingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	building := models.Building{
		Name:        input.Name,
		Archived:    input.Archived,
		Description: input.Description,
		Address:     input.Address,
	}

	if err := h.DB.Create(&building).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, building)
}

func (h *BuildingHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid building id")
		return
	}

	var building models.Building
	if err := h.DB.First(&building, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "building not found")
			return
		}
		internalError(c, err)
		return
	}

	var input buildingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	building.Name = input.Name
	building.Archived = input.Archived
	building.Description = input.Description
	building.Address = input.Address

	if err := h.DB.Save(&building).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, building)
}

func (h *BuildingHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid building id")
		return
	}

	result := h.DB.Model(&models.Building{}).Where("id = ?", id).Update("archived", true)
	if result.Error != nil {
		internalError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		notFound(c, "building not found")
		return
	}

	c.Status(http.StatusNoContent)
}
