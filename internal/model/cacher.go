package model

import (
	"context"
	"time"
)

// NilKey is used to indicate that this value / cache is intended to be treated as nil
// and different from not found from cache
const NilKey = "NIL"

// Cacher :nodoc:
type Cacher interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, exp time.Duration) error
}
