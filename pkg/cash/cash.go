package cash

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Expireser interface {
	ExpiresAt() time.Time
}

type Cash[K comparable, V Expireser] struct {
	data map[K]V
	mu   sync.RWMutex
}

var ErrNotFound = errors.New("element not found")

func New[K comparable, V Expireser](ctx context.Context, gcCallPeriod time.Duration) *Cash[K, V] {
	cash := &Cash[K, V]{
		data: make(map[K]V),
		mu:   sync.RWMutex{},
	}

	go cash.deleteOldObjects(ctx, gcCallPeriod)

	return cash
}

func (c *Cash[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
}

func (c *Cash[K, V]) Get(key K) (V, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.data[key]; ok {
		return v, nil
	}

	var none V
	return none, ErrNotFound
}

func (c *Cash[K, V]) GetWithDelete(key K) (V, error) {
	v, err := c.Get(key)
	if err != nil {
		var none V
		return none, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)

	return v, nil
}

func (c *Cash[K, V]) deleteOldObjects(ctx context.Context, gcCallPeriod time.Duration) {
	t := time.NewTicker(gcCallPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		for k, v := range c.data {
			if time.Now().Compare(v.ExpiresAt()) == 1 {
				c.mu.Lock()
				delete(c.data, k)
				c.mu.Unlock()
			}
		}
	}
}
