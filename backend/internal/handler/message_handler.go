package handler

import (
	"errors"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	service service.MessageService
}

func NewMessageHandler(service service.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

func (h *MessageHandler) GetMessage(c *gin.Context) {
	// Parse query parameters
	// fromdb: must be 1 in initial implementation
	fromdbStr := c.DefaultQuery("fromdb", "1")
	if fromdbStr != "1" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request: 'fromdb' must be 1"})
		return
	}

	// cache: 1 or 0, default 1
	cacheStr := c.DefaultQuery("cache", "1")
	var useCache bool
	switch cacheStr {
	case "1":
		useCache = true
	case "0":
		useCache = false
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request: 'cache' must be 0 or 1"})
		return
	}

	// sql: 0 or 1, default 0
	sqlStr := c.DefaultQuery("sql", "0")
	var useSQL bool
	switch sqlStr {
	case "0":
		useSQL = false
	case "1":
		useSQL = true
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request: 'sql' must be 0 or 1"})
		return
	}

	if useSQL {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not Implemented: PostgreSQL is not supported"})
		return
	}

	opts := service.MessageOptions{
		FromDB: true,
		Cache:  useCache,
		SQL:    useSQL,
	}

	result, err := h.service.GetMessage(c.Request.Context(), opts)
	if err != nil {
		if errors.Is(err, service.ErrNotImplemented) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Not Implemented: SQL database is not supported"})
			return
		}
		if errors.Is(err, service.ErrInvalidRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error: failed to retrieve message"})
		return
	}

	// Store source in context for structured logger
	c.Set(middleware.ContextSourceKey, result.Source)

	c.JSON(http.StatusOK, result)
}
