package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"gorm.io/gorm"
)

type EquipmentHandler struct {
	DB *gorm.DB
}

func NewEquipmentHandler(db *gorm.DB) *EquipmentHandler {
	return &EquipmentHandler{DB: db}
}

func (h *EquipmentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	equipment := rg.Group("/equipment")
	{
		equipment.GET("", h.List)
		equipment.GET("/:id", h.Get)
		equipment.POST("", h.Create)
		equipment.PUT("/:id", h.Update)
		equipment.DELETE("/:id", h.Delete)
	}
}

type equipmentInput struct {
	Name        string `binding:"required"`
	Description *string
	Disabled    bool
	Archived    bool
	GroupID     uint `binding:"required"`
}

var equipmentSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"disabled":   "disabled",
	"archived":   "archived",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// equipmentGroupExists is used to validate a client-supplied GroupID on create/update; it's a
// 400 (bad input), not a 404, since the group isn't the resource the URL identifies.
func (h *EquipmentHandler) equipmentGroupExists(groupID uint) (bool, error) {
	var count int64
	if err := h.DB.Model(&models.EquipmentGroup{}).Where("id = ?", groupID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *EquipmentHandler) List(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	sortColumn, sortOrder, err := parseSort(c, equipmentSortColumns, "name")
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	query := h.DB.Model(&models.Equipment{})

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	if v := strings.TrimSpace(c.Query("groupId")); v != "" {
		groupID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			badRequest(c, "invalid groupId filter")
			return
		}
		query = query.Where("group_id = ?", uint(groupID))
	}

	if archived, present, err := parseOptionalBool(c, "archived"); err != nil {
		badRequest(c, err.Error())
		return
	} else if present {
		query = query.Where("archived = ?", archived)
	}

	if disabled, present, err := parseOptionalBool(c, "disabled"); err != nil {
		badRequest(c, err.Error())
		return
	} else if present {
		query = query.Where("disabled = ?", disabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		internalError(c, err)
		return
	}

	var equipment []models.Equipment
	offset := (page - 1) * pageSize
	if err := query.
		Order(sortColumn + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&equipment).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, listResponse{
		Data: equipment,
		Meta: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages(total, pageSize),
		},
	})
}

func (h *EquipmentHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment id")
		return
	}

	var equipment models.Equipment
	if err := h.DB.First(&equipment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "equipment not found")
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, equipment)
}

func (h *EquipmentHandler) Create(c *gin.Context) {
	var input equipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	exists, err := h.equipmentGroupExists(input.GroupID)
	if err != nil {
		internalError(c, err)
		return
	}
	if !exists {
		badRequest(c, "equipment group not found")
		return
	}

	equipment := models.Equipment{
		Name:        input.Name,
		Description: input.Description,
		Disabled:    input.Disabled,
		Archived:    input.Archived,
		GroupID:     input.GroupID,
	}

	if err := h.DB.Create(&equipment).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, equipment)
}

func (h *EquipmentHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment id")
		return
	}

	var equipment models.Equipment
	if err := h.DB.First(&equipment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "equipment not found")
			return
		}
		internalError(c, err)
		return
	}

	var input equipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	exists, err := h.equipmentGroupExists(input.GroupID)
	if err != nil {
		internalError(c, err)
		return
	}
	if !exists {
		badRequest(c, "equipment group not found")
		return
	}

	equipment.Name = input.Name
	equipment.Description = input.Description
	equipment.Disabled = input.Disabled
	equipment.Archived = input.Archived
	equipment.GroupID = input.GroupID

	if err := h.DB.Save(&equipment).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, equipment)
}

func (h *EquipmentHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment id")
		return
	}

	result := h.DB.Model(&models.Equipment{}).Where("id = ?", id).Update("archived", true)
	if result.Error != nil {
		internalError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		notFound(c, "equipment not found")
		return
	}

	c.Status(http.StatusNoContent)
}
