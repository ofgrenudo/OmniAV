package handlers

import (
	"errors"
	"net/http"
	"sort"
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
		groups.GET("/:id/schedule", h.Schedule)
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

	// buildingId narrows the list to groups that actually have a usable unit stocked in that
	// building, so request flows never offer a group whose units all live somewhere else.
	if buildingID, present, err := parseOptionalUint(c, "buildingId"); err != nil {
		badRequest(c, err.Error())
		return
	} else if present {
		stockedHere := h.DB.Model(&models.Equipment{}).
			Select("1").
			Where("equipment.group_id = equipment_groups.id").
			Where("equipment.building_id = ? AND equipment.disabled = ? AND equipment.archived = ?", buildingID, false, false)
		query = query.Where("EXISTS (?)", stockedHere)
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

	if exists, err := buildingExists(h.DB, candidate.BuildingID); err != nil {
		internalError(c, err)
		return
	} else if !exists {
		badRequest(c, "building not found")
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

// scheduleEntry is one unit's booking on the requested day: where it has to be, when, and for whom.
// Times are formatted HH:MM so the client can lay them out directly without re-parsing timestamps.
type scheduleEntry struct {
	EquipmentID   uint    `json:"equipmentId"`
	EquipmentName string  `json:"equipmentName"`
	RequestID     uint    `json:"requestId"`
	RequestName   string  `json:"requestName"`
	BuildingID    uint    `json:"buildingId"`
	BuildingName  string  `json:"buildingName"`
	Room          string  `json:"room"`
	StartTime     string  `json:"startTime"`
	EndTime       string  `json:"endTime"`
	Comments      *string `json:"comments"`
}

// Schedule lists every booking for this group's units on a single calendar day — the "where has
// this equipment been / where is it going" view. Requests recur weekly, so a booking shows up on
// each date its weekly run covers, not just its first date.
func (h *EquipmentGroupHandler) Schedule(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid equipment group id")
		return
	}

	dateStr := strings.TrimSpace(c.Query("date"))
	if dateStr == "" {
		badRequest(c, "date is required")
		return
	}
	date, err := parseDateOnly(dateStr)
	if err != nil {
		badRequest(c, "invalid date, expected YYYY-MM-DD")
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

	// Narrow to requests that can possibly land on this exact date (same weekday, and date falls
	// inside the request's weekly run) at the SQL level, rather than fetching a group's entire
	// booking history and filtering it in Go — this table only grows over time.
	var booked []models.RequestedEquipment
	if err := h.DB.
		Preload("Equipment").
		Preload("Request").
		Preload("Request.Building").
		Joins("JOIN requests ON requests.id = requested_equipments.request_id").
		Where("requested_equipments.group_id = ?", id).
		Where("requests.days_of_week = ?", models.FromGoWeekday(date.Weekday())).
		Where("requests.first_date_needed <= ?", date).
		Where("requests.first_date_needed >= ? - (requests.number_of_weeks - 1) * INTERVAL '7 days'", date).
		Find(&booked).Error; err != nil {
		internalError(c, err)
		return
	}

	entries := make([]scheduleEntry, 0, len(booked))
	for _, re := range booked {
		if re.Request == nil || !services.RequestOccursOn(re.Request, date) {
			continue
		}

		entry := scheduleEntry{
			EquipmentID: re.EquipmentID,
			RequestID:   re.RequestID,
			RequestName: re.Request.Name,
			BuildingID:  re.Request.BuildingID,
			Room:        re.Request.Room,
			StartTime:   re.Request.StartTime.UTC().Format(clockLayout),
			EndTime:     re.Request.EndTime.UTC().Format(clockLayout),
			Comments:    re.Request.Comments,
		}
		if re.Equipment != nil {
			entry.EquipmentName = re.Equipment.Name
		}
		if re.Request.Building != nil {
			entry.BuildingName = re.Request.Building.Name
		}
		entries = append(entries, entry)
	}

	// Chronological, then by unit, so the client can render rows in a stable order.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].StartTime != entries[j].StartTime {
			return entries[i].StartTime < entries[j].StartTime
		}
		return entries[i].EquipmentName < entries[j].EquipmentName
	})

	c.JSON(http.StatusOK, gin.H{"date": dateStr, "data": entries})
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

	// buildingId is required: every unit has a permanent home, so an availability probe
	// without a building can't correspond to any real request and always returns 0. Require
	// it up front so callers get a clear 400 instead of a silent zero.
	buildingID, present, err := parseOptionalUint(c, "buildingId")
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, errors.New("buildingId is required")
	}

	return &models.Request{
		FirstDateNeeded: firstDate,
		StartTime:       startTime,
		EndTime:         endTime,
		NumberOfWeeks:   weeks,
		DaysOfWeek:      models.FromGoWeekday(firstDate.Weekday()),
		BuildingID:      buildingID,
	}, nil
}
