package storage

import (
	"context"
	"sizebot/internal/entities"
)

// MainDBStorage defines interface for maindb storage.
type MainDBStorage interface {
	// Commands returns all commands from the storage. Any error returned is internal.
	Commands(ctx context.Context) (entities.Commands, error)
}
