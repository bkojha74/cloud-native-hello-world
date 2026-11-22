package service

import (
	"context"
	"errors"
	"log"

	"backend/internal/cache"
	"backend/internal/metrics"
	"backend/internal/repository"
)

const MessageCacheKey = "hello-world:message:1"

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrNotImplemented = errors.New("not implemented")
)

type MessageOptions struct {
	FromDB bool
	Cache  bool
	SQL    bool
}

type MessageResult struct {
	Message string `json:"message"`
	Source  string `json:"source"`
}

type MessageService interface {
	GetMessage(ctx context.Context, options MessageOptions) (MessageResult, error)
}

type messageService struct {
	repo  repository.MessageRepository
	cache cache.MessageCache
}

func NewMessageService(repo repository.MessageRepository, cache cache.MessageCache) MessageService {
	return &messageService{
		repo:  repo,
		cache: cache,
	}
}

func (s *messageService) GetMessage(ctx context.Context, opts MessageOptions) (MessageResult, error) {
	if opts.SQL {
		return MessageResult{}, ErrNotImplemented
	}

	if !opts.FromDB {
		return MessageResult{}, ErrInvalidRequest
	}

	// 1. Check cache if enabled
	if opts.Cache && s.cache != nil {
		cachedMsg, err := s.cache.Get(ctx, MessageCacheKey)
		if err == nil && cachedMsg != "" {
			metrics.CacheHitsTotal.Inc()
			return MessageResult{
				Message: cachedMsg,
				Source:  "cache",
			}, nil
		}

		if errors.Is(err, cache.ErrCacheMiss) {
			metrics.CacheMissesTotal.Inc()
		} else if err != nil {
			// Redis failures must not fail the API request. Log and count the Redis error, then continue to MongoDB.
			metrics.CacheErrorsTotal.Inc()
			log.Printf("[WARN] Redis cache error: %v", err)
		}
	}

	// 2. Fetch from repository
	msg, err := s.repo.GetMessage(ctx)
	if err != nil {
		return MessageResult{}, err
	}

	// 3. Best-effort write to cache
	if opts.Cache && s.cache != nil {
		if err := s.cache.Set(ctx, MessageCacheKey, msg); err != nil {
			metrics.CacheErrorsTotal.Inc()
			log.Printf("[WARN] Failed to write to Redis cache: %v", err)
		}
	}

	return MessageResult{
		Message: msg,
		Source:  "mongodb",
	}, nil
}
