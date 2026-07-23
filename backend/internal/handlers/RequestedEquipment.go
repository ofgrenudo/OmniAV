package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"github.com/ofgrenudo/OmniAv/internal/services"
	"gorm.io/gorm"
)

type RequestedEquipmentHandler struct {
	DB *gorm.DB
}

func NewRequestedEquipmentHandler(db *gorm.DB) *RequestedEquipmentHandler {
	return &RequestedEquipmentHandler{DB: db}
}

// RegisterRoutes nests under /requests/:id/equipment, reusing RequestHandler's ":id" param name
// at the same tree position — gin requires the same wildcard name wherever two handlers share a
// route segment.
func (h *RequestedEquipmentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	requestEquipment := rg.Group("/requests/:id/equipment")
	{
		requestEquipment.GET("", h.List)
		requestEquipment.POST("", h.Create)
		requestEquipment.DELETE("/:equipmentId", h.Delete)
	}
}

type requestedEquipmentInput struct {
	GroupID uint `json:"groupId" binding:"required"`
}

func (h *RequestedEquipmentHandler) List(c *gin.Context) {
	requestID, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}

	var items []models.RequestedEquipment
	if err := h.DB.
		Preload("Equipment").
		Preload("Group").
		Where("request_id = ?", requestID).
		Order("id ASC").
		Find(&items).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

// Create assigns one unit from the requested group to the request. The frontend only ever
// presents a group (e.g. "Cow Cart"); services.CreateRequestedEquipment resolves it to a specific
// available unit (e.g. "Cow Cart B") and rejects the call if none is free for this request's
// schedule. Call it once per unit needed.
func (h *RequestedEquipmentHandler) Create(c *gin.Context) {
	requestID, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}

	var input requestedEquipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err.Error())
		return
	}

	created, err := services.CreateRequestedEquipment(h.DB, requestID, input.GroupID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRequestNotFound):
			notFound(c, err.Error())
		case errors.Is(err, services.ErrEquipmentGroupNotFound):
			notFound(c, err.Error())
		case errors.Is(err, services.ErrEquipmentGroupUnavailable), errors.Is(err, services.ErrNoEquipmentAvailable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			internalError(c, err)
		}
		return
	}

	if err := h.DB.Preload("Equipment").Preload("Group").First(created, created.ID).Error; err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Delete removes a specific assignment, freeing that unit for other requests' overlapping times.
func (h *RequestedEquipmentHandler) Delete(c *gin.Context) {
	requestID, err := parseIDParam(c)
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}

	equipmentRequestID, err := strconv.ParseUint(c.Param("equipmentId"), 10, 64)
	if err != nil {
		badRequest(c, "invalid requested equipment id")
		return
	}

	result := h.DB.
		Where("id = ? AND request_id = ?", uint(equipmentRequestID), requestID).
		Delete(&models.RequestedEquipment{})
	if result.Error != nil {
		internalError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		notFound(c, "requested equipment not found")
		return
	}

	c.Status(http.StatusNoContent)
}
