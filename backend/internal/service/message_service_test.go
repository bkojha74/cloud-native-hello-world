package service_test

import (
	"context"
	"errors"
	"testing"

	"backend/internal/cache"
	"backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) GetMessage(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type MockMessageCache struct {
	mock.Mock
}

func (m *MockMessageCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockMessageCache) Set(ctx context.Context, key string, value string) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func TestMessageService_GetMessage_CacheHit(t *testing.T) {
	mockCache := new(MockMessageCache)
	mockRepo := new(MockMessageRepo)
	svc := service.NewMessageService(mockRepo, mockCache)

	mockCache.On("Get", mock.Anything, service.MessageCacheKey).Return("Hello Cached World", nil)

	res, err := svc.GetMessage(context.Background(), service.MessageOptions{
		FromDB: true,
		Cache:  true,
		SQL:    false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hello Cached World", res.Message)
	assert.Equal(t, "cache", res.Source)
	mockRepo.AssertNotCalled(t, "GetMessage", mock.Anything)
}

func TestMessageService_GetMessage_CacheMiss_Then_DBHit(t *testing.T) {
	mockCache := new(MockMessageCache)
	mockRepo := new(MockMessageRepo)
	svc := service.NewMessageService(mockRepo, mockCache)

	mockCache.On("Get", mock.Anything, service.MessageCacheKey).Return("", cache.ErrCacheMiss)
	mockRepo.On("GetMessage", mock.Anything).Return("Hello World", nil)
	mockCache.On("Set", mock.Anything, service.MessageCacheKey, "Hello World").Return(nil)

	res, err := svc.GetMessage(context.Background(), service.MessageOptions{
		FromDB: true,
		Cache:  true,
		SQL:    false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", res.Message)
	assert.Equal(t, "mongodb", res.Source)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestMessageService_GetMessage_RedisFailureFallback(t *testing.T) {
	mockCache := new(MockMessageCache)
	mockRepo := new(MockMessageRepo)
	svc := service.NewMessageService(mockRepo, mockCache)

	// Simulate Redis connection failure
	mockCache.On("Get", mock.Anything, service.MessageCacheKey).Return("", errors.New("connection refused"))
	mockRepo.On("GetMessage", mock.Anything).Return("Hello World", nil)
	mockCache.On("Set", mock.Anything, service.MessageCacheKey, "Hello World").Return(errors.New("connection refused"))

	res, err := svc.GetMessage(context.Background(), service.MessageOptions{
		FromDB: true,
		Cache:  true,
		SQL:    false,
	})

	// Must NOT fail the request; fallback to MongoDB
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", res.Message)
	assert.Equal(t, "mongodb", res.Source)
}

func TestMessageService_GetMessage_CacheDisabled(t *testing.T) {
	mockCache := new(MockMessageCache)
	mockRepo := new(MockMessageRepo)
	svc := service.NewMessageService(mockRepo, mockCache)

	mockRepo.On("GetMessage", mock.Anything).Return("Hello Direct DB", nil)

	res, err := svc.GetMessage(context.Background(), service.MessageOptions{
		FromDB: true,
		Cache:  false,
		SQL:    false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hello Direct DB", res.Message)
	assert.Equal(t, "mongodb", res.Source)
	mockCache.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestMessageService_GetMessage_SQLNotImplemented(t *testing.T) {
	svc := service.NewMessageService(nil, nil)

	_, err := svc.GetMessage(context.Background(), service.MessageOptions{
		FromDB: true,
		Cache:  true,
		SQL:    true,
	})

	assert.ErrorIs(t, err, service.ErrNotImplemented)
}
