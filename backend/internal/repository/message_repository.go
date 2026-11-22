package repository

import (
	"context"
	"errors"
)

var (
	ErrNotFound       = errors.New("message not found")
	ErrNotImplemented = errors.New("not implemented")
)

type MessageRepository interface {
	GetMessage(ctx context.Context) (string, error)
}
