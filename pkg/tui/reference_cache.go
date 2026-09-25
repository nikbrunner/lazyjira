package tui

import (
	"time"
)

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

type ttlCache[T any] struct {
	enabled bool
	ttl     time.Duration
	now     func() time.Time
	entries map[string]cacheEntry[T]
}

func newTTLCache[T any](enabled bool, ttl time.Duration) ttlCache[T] {
	return ttlCache[T]{enabled: enabled, ttl: ttl, now: time.Now, entries: make(map[string]cacheEntry[T])}
}

func (c *ttlCache[T]) get(key string) (T, bool) {
	var zero T
	if !c.enabled {
		return zero, false
	}
	entry, ok := c.entries[key]
	if !ok {
		return zero, false
	}
	if !c.now().Before(entry.expiresAt) {
		delete(c.entries, key)
		return zero, false
	}
	return entry.value, true
}

func (c *ttlCache[T]) set(key string, value T) {
	if !c.enabled {
		return
	}
	if c.entries == nil {
		c.entries = make(map[string]cacheEntry[T])
	}
	c.entries[key] = cacheEntry[T]{value: value, expiresAt: c.now().Add(c.ttl)}
}

func (c *ttlCache[T]) clear() {
	clear(c.entries)
}

func (a *App) invalidateReferenceCaches() {
	a.referenceCacheVersion++
	a.boardsCache.clear()
	a.sprintsCache.clear()
	a.usersCache.clear()
	a.createMetaCache.clear()
}

func cacheTTL(value string) time.Duration {
	ttl, err := time.ParseDuration(value)
	if err != nil || ttl <= 0 {
		return 5 * time.Minute
	}
	return ttl
}
