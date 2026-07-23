package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"gorm.io/gorm"
)

type RequestHandler struct {
	DB *gorm.DB
}

func NewRequestHandler(db *gorm.DB) *RequestHandler {
	return &RequestHandler{DB: db}
}

func (h *RequestHandler) RegisterRoutes(rg *gin.RouterGroup) {
	requests := rg.Group("/requests")
	{
		requests.GET("", h.List)
		requests.GET("/:id", h.Get)
		requests.POST("", h.Create)
		requests.DELETE("/:id", h.Delete)
	}
}

// requestInput deliberately has no DaysOfWeek field: it's derived from FirstDateNeeded server-side
// (see models.FromGoWeekday) so it can never disagree with the actual calendar date, which the
// equipment-conflict logic in the services package depends on.
type requestInput struct {
	Name            string `binding:"required"`
	FirstDateNeeded string `binding:"required"`
	StartTime       string `binding:"required"`
	EndTime         string `binding:"required"`
	NumberOfWeeks   int    `binding:"required,min=1"`
	BuildingID      uint   `binding:"required"`
	Room            string `binding:"required"`
	Comments        *string
}

var requestSortColumns = map[string]string{
	"id":                "id",
	"name":              "name",
	"first_date_needed": "first_date_needed",
	"created_at":        "created_at",
	"updated_at":        "updated_at",
}

// minAdvanceNotice mirrors the frontend's "requests must be made at least 24 hours in advance"
// rule (see frontend/src/utils/time.ts isAtLeast24HoursOut); enforced here too since a client-side
// check alone is trivially bypassed by calling the API directly.
const minAdvanceNotice = 24 * time.Hour

func (h *RequestHandler) List(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	sortColumn, sortOrder, err := parseSort(c, requestSortColumns, "first_date_needed")
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	query := h.DB.Model(&models.Request{})

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR room ILIKE ?", like, like)
	}

	if v := strings.TrimSpace(c.Query("buildingId")); v != "" {
		buildingID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			badRequest(c, "invalid buildingId filter")
			return
		}
		query = query.Where("building_id = ?", uint(buildingID))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		internalError(c, err)
		return
	}

	var requests []models.Request
	offset := (page - 1) * pageSize
	if err := query.
		Preload("Building").
		Order(sortColumn + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&requests).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, listResponse{
		Data: requests,
		Meta: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages(total, pageSize),
		},
	})
}

func (h *RequestHandler) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}

	var request models.Request
	if err := h.DB.
		Preload("Building").
		Preload("RequestedEquipment").
		Preload("RequestedEquipment.Equipment").
		Preload("RequestedEquipment.Group").
		First(&request, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "request not found")
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, request)
}

func (h *RequestHandler) Create(c *gin.Context) {
	var input requestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequest(c, "name must not be blank")
		return
	}
	input.Room = strings.TrimSpace(input.Room)
	if input.Room == "" {
		badRequest(c, "room must not be blank")
		return
	}

	firstDate, err := parseDateOnly(input.FirstDateNeeded)
	if err != nil {
		badRequest(c, "invalid firstDateNeeded, expected YYYY-MM-DD")
		return
	}
	startTime, err := parseClockTime(input.StartTime)
	if err != nil {
		badRequest(c, "invalid startTime, expected HH:MM")
		return
	}
	endTime, err := parseClockTime(input.EndTime)
	if err != nil {
		badRequest(c, "invalid endTime, expected HH:MM")
		return
	}
	if !endTime.After(startTime) {
		badRequest(c, "endTime must be after startTime")
		return
	}

	requestedAt := firstDate.Add(time.Duration(startTime.Hour())*time.Hour + time.Duration(startTime.Minute())*time.Minute)
	if time.Until(requestedAt) < minAdvanceNotice {
		badRequest(c, "requests must be made at least 24 hours in advance")
		return
	}

	var building models.Building
	if err := h.DB.First(&building, input.BuildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			badRequest(c, "building not found")
			return
		}
		internalError(c, err)
		return
	}

	request := models.Request{
		Name:            input.Name,
		FirstDateNeeded: firstDate,
		StartTime:       startTime,
		EndTime:         endTime,
		NumberOfWeeks:   input.NumberOfWeeks,
		DaysOfWeek:      models.FromGoWeekday(firstDate.Weekday()),
		BuildingID:      input.BuildingID,
		Room:            input.Room,
		Comments:        input.Comments,
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where(models.BuildingRoom{BuildingID: input.BuildingID, Room: input.Room}).
			FirstOrCreate(&models.BuildingRoom{}).Error; err != nil {
			return err
		}
		return tx.Create(&request).Error
	})
	if err != nil {
		internalError(c, err)
		return
	}

	request.Building = &building
	c.JSON(http.StatusCreated, request)
}

func (h *RequestHandler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("request_id = ?", id).Delete(&models.RequestedEquipment{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.Request{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound(c, "request not found")
			return
		}
		internalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
