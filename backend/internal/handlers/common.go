package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100

	dateOnlyLayout = "2006-01-02"
	clockLayout    = "15:04"
)

// parseDateOnly and parseClockTime always parse against UTC, matching how the services package
// normalizes time.Time values before comparing them (see services.dateOnly/secondOfDay).
func parseDateOnly(s string) (time.Time, error) {
	return time.ParseInLocation(dateOnlyLayout, s, time.UTC)
}

func parseClockTime(s string) (time.Time, error) {
	return time.ParseInLocation(clockLayout, s, time.UTC)
}

type paginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type listResponse struct {
	Data any            `json:"data"`
	Meta paginationMeta `json:"meta"`
}

func totalPages(total int64, pageSize int) int {
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// parseIDParam parses the ":id" route param shared by every resource's Get/Update/Delete routes.
func parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parsePagination(c *gin.Context) (page, pageSize int, err error) {
	page = 1
	if v := c.Query("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil || page < 1 {
			return 0, 0, errors.New("invalid page")
		}
	}

	pageSize = defaultPageSize
	if v := c.Query("pageSize"); v != "" {
		pageSize, err = strconv.Atoi(v)
		if err != nil || pageSize < 1 {
			return 0, 0, errors.New("invalid pageSize")
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize, nil
}

// parseSort validates the "sort"/"order" query params against allowed, a whitelist mapping the
// client-facing sort key to the actual DB column (preventing SQL injection via ORDER BY).
func parseSort(c *gin.Context, allowed map[string]string, defaultColumn string) (column, direction string, err error) {
	column = defaultColumn
	if v := c.Query("sort"); v != "" {
		col, ok := allowed[v]
		if !ok {
			return "", "", errors.New("invalid sort column")
		}
		column = col
	}

	direction = "ASC"
	if v := strings.ToLower(c.Query("order")); v != "" {
		if v != "asc" && v != "desc" {
			return "", "", errors.New("invalid order, must be asc or desc")
		}
		direction = strings.ToUpper(v)
	}

	return column, direction, nil
}

func parseOptionalBool(c *gin.Context, name string) (value bool, present bool, err error) {
	v := c.Query(name)
	if v == "" {
		return false, false, nil
	}
	value, err = strconv.ParseBool(v)
	if err != nil {
		return false, false, errors.New("invalid " + name + " filter")
	}
	return value, true, nil
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func notFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func internalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
