package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handler"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetMessage(ctx context.Context, opts service.MessageOptions) (service.MessageResult, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(service.MessageResult), args.Error(1)
}

func setupTestRouter(svc service.MessageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	msgHandler := handler.NewMessageHandler(svc)
	healthHandler := handler.NewHealthHandler()

	r.GET("/message", msgHandler.GetMessage)
	r.GET("/health", healthHandler.GetHealth)
	return r
}

func TestHandler_GetHealth(t *testing.T) {
	r := setupTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "UP", body["status"])
}

func TestHandler_GetMessage_Success(t *testing.T) {
	mockSvc := new(MockService)
	mockSvc.On("GetMessage", mock.Anything, service.MessageOptions{
		FromDB: true,
		Cache:  true,
		SQL:    false,
	}).Return(service.MessageResult{Message: "Hello World", Source: "cache"}, nil)

	r := setupTestRouter(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/message?fromdb=1&cache=1&sql=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "Hello World", body["message"])
	assert.Equal(t, "cache", body["source"])
}

func TestHandler_GetMessage_InvalidFromDB(t *testing.T) {
	r := setupTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/message?fromdb=0&cache=1&sql=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetMessage_InvalidCache(t *testing.T) {
	r := setupTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/message?fromdb=1&cache=invalid&sql=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetMessage_SQLNotImplemented(t *testing.T) {
	r := setupTestRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/message?fromdb=1&cache=1&sql=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
