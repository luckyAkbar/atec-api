// Package db contains all database related code
package db

import (
	"context"
	"time"

	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type cacher struct {
	client *redis.Client
}

// NewCacher return a new model.Chacher instance
func NewCacher(client *redis.Client) model.Cacher {
	return &cacher{
		client: client,
	}
}

// Get get cache value by given key. Return json string if found. Otherwise return a non nil error
func (c *cacher) Get(ctx context.Context, key string) (string, error) {
	res, err := c.client.Get(ctx, key).Result()
	switch err {
	case nil:
		return res, nil
	case redis.Nil:
		return res, err
	default:
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"key": key,
		}).Error(err)
		return res, err
	}
}

// Set set a cache value by key with the given expiry time. The val should be a json string
func (c *cacher) Set(ctx context.Context, key string, val string, exp time.Duration) error {
	err := c.client.Set(ctx, key, val, exp).Err()
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"key": key,
			"val": val,
		}).Error(err)

		return err
	}

	return nil
}

func (c *cacher) Del(ctx context.Context, key []string) error {
	err := c.client.Del(ctx, key...).Err()
	if err != nil {
		return err
	}

	return nil
}
