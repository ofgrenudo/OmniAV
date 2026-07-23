package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/ofgrenudo/OmniAv/internal/services"
	"gorm.io/gorm"
)

type EquipmentGroupHandler struct {
	DB *gorm.DB
}

func NewEquipmentGroupHandler(db *gorm.DB) *EquipmentGroupHandler {
	return &EquipmentGroupHandler{DB: db}
}

func (h *EquipmentGroupHandler) RegisterRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/equipment-groups")
	{
		groups.GET("", h.List)
		groups.GET("/:id", h.Get)
		groups.GET("/:id/availability", h.Availability)
		groups.POST("", h.Create)
		groups.PUT("/:id", h.Update)
		groups.DELETE("/:id", h.Delete)
	}
}

type equipmentGroupInput struct {
	Name        string `binding:"required"`
	Description *string
	Disabled    bool
	Archived    bool
}

var equipmentGroupSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"disabled":   "disabled",
	"archived":   "archived",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func (h *EquipmentGroupHandler) List(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	sortColumn, sortOrder, err := parseSort(c, equipmentGroupSortColumns, "name")
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	query := h.DB.Model(&models.EquipmentGroup{})

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", like, like)
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

	var groups []models.EquipmentGroup
	offset := (page - 1) * pageSize
	if err := query.
		Order(sortColumn + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&groups).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, listResponse{
		Data: groups,
		Meta: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages(total, pageSize),
		},
	})
}

func (h *EquipmentGroupHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment group id")
		return
	}

	var group models.EquipmentGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "equipment group not found")
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *EquipmentGroupHandler) Create(c *gin.Context) {
	var input equipmentGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	group := models.EquipmentGroup{
		Name:        input.Name,
		Description: input.Description,
		Disabled:    input.Disabled,
		Archived:    input.Archived,
	}

	if err := h.DB.Create(&group).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *EquipmentGroupHandler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment group id")
		return
	}

	var group models.EquipmentGroup
	if err := h.DB.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "equipment group not found")
			return
		}
		internalError(c, err)
		return
	}

	var input equipmentGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}

	group.Name = input.Name
	group.Description = input.Description
	group.Disabled = input.Disabled
	group.Archived = input.Archived

	if err := h.DB.Save(&group).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *EquipmentGroupHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment group id")
		return
	}

	result := h.DB.Model(&models.EquipmentGroup{}).Where("id = ?", id).Update("archived", true)
	if result.Error != nil {
		internalError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		notFound(c, "equipment group not found")
		return
	}

	c.Status(http.StatusNoContent)
}

// Availability reports how many units in the group are free for a candidate schedule, without
// requiring a Request to exist yet — lets the frontend show a live "available: N" count as the
// user fills out the When & Where step, before anything is submitted.
func (h *EquipmentGroupHandler) Availability(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment group id")
		return
	}

	candidate, err := parseAvailabilityQuery(c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	available, err := services.AvailableCount(h.DB, id, candidate)
	if err != nil {
		if errors.Is(err, services.ErrEquipmentGroupNotFound) {
			notFound(c, "equipment group not found")
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": available})
}

func parseAvailabilityQuery(c *gin.Context) (*models.Request, error) {
	firstDateStr := c.Query("firstDate")
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	if firstDateStr == "" || startTimeStr == "" || endTimeStr == "" {
		return nil, errors.New("firstDate, startTime, and endTime are required")
	}

	firstDate, err := parseDateOnly(firstDateStr)
	if err != nil {
		return nil, errors.New("invalid firstDate, expected YYYY-MM-DD")
	}
	startTime, err := parseClockTime(startTimeStr)
	if err != nil {
		return nil, errors.New("invalid startTime, expected HH:MM")
	}
	endTime, err := parseClockTime(endTimeStr)
	if err != nil {
		return nil, errors.New("invalid endTime, expected HH:MM")
	}
	if !endTime.After(startTime) {
		return nil, errors.New("endTime must be after startTime")
	}

	weeks := 1
	if v := c.Query("weeks"); v != "" {
		weeks, err = strconv.Atoi(v)
		if err != nil || weeks < 1 {
			return nil, errors.New("invalid weeks, must be a positive integer")
		}
	}

	return &models.Request{
		FirstDateNeeded: firstDate,
		StartTime:       startTime,
		EndTime:         endTime,
		NumberOfWeeks:   weeks,
		DaysOfWeek:      models.FromGoWeekday(firstDate.Weekday()),
	}, nil
}
