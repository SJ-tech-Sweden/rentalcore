package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a single cached value with metadata for monitoring.
type CacheEntry struct {
	Value      interface{}
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Hits       int64
	LastAccess time.Time
}

// CacheManager provides a lightweight in-memory cache with basic TTL support.
type CacheManager struct {
	mu     sync.RWMutex
	items  map[string]*CacheEntry
	stats  CacheStats
	ticker *time.Ticker
	quit   chan struct{}
}

// CacheStats captures aggregate statistics for monitoring.
type CacheStats struct {
	Items        int64
	Hits         int64
	Misses       int64
	Evictions    int64
	LastEviction time.Time
}

// NewCacheManager creates a cache manager with automatic cleanup of expired entries.
func NewCacheManager() *CacheManager {
	manager := &CacheManager{
		items:  make(map[string]*CacheEntry),
		ticker: time.NewTicker(1 * time.Minute),
		quit:   make(chan struct{}),
	}

	go manager.startEvictionLoop()
	return manager
}

// startEvictionLoop periodically removes expired entries to keep the cache tidy.
func (m *CacheManager) startEvictionLoop() {
	for {
		select {
		case <-m.ticker.C:
			m.evictExpired()
		case <-m.quit:
			return
		}
	}
}

// Stop gracefully stops background eviction processing.
func (m *CacheManager) Stop() {
	close(m.quit)
	m.ticker.Stop()
}

// Set inserts or replaces a cache entry with the provided TTL.
func (m *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &CacheEntry{
		Value:      value,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
	if ttl > 0 {
		entry.ExpiresAt = entry.CreatedAt.Add(ttl)
	}

	if _, exists := m.items[key]; exists {
		// Replacing an existing entry; no need to modify eviction stats.
		m.items[key] = entry
		return
	}

	m.items[key] = entry
	m.stats.Items = int64(len(m.items))
}

// Get retrieves a cached value and reports whether it was found and not expired.
func (m *CacheManager) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.items[key]
	if !ok {
		m.stats.Misses++
		return nil, false
	}

	if entry.ExpiresAt.IsZero() || entry.ExpiresAt.After(time.Now()) {
		entry.Hits++
		entry.LastAccess = time.Now()
		m.stats.Hits++
		return entry.Value, true
	}

	// Entry expired
	delete(m.items, key)
	m.stats.Evictions++
	m.stats.LastEviction = time.Now()
	m.stats.Items = int64(len(m.items))
	m.stats.Misses++
	return nil, false
}

// Delete removes an entry if it exists.
func (m *CacheManager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[key]; exists {
		delete(m.items, key)
		m.stats.Items = int64(len(m.items))
		m.stats.Evictions++
		m.stats.LastEviction = time.Now()
	}
}

// Clear removes all entries from the cache.
func (m *CacheManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*CacheEntry)
	m.stats.Items = 0
}

// evictExpired removes expired entries and updates stats.
func (m *CacheManager) evictExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	evicted := 0
	for key, entry := range m.items {
		if !entry.ExpiresAt.IsZero() && entry.ExpiresAt.Before(now) {
			delete(m.items, key)
			evicted++
		}
	}

	if evicted > 0 {
		m.stats.Items = int64(len(m.items))
		m.stats.Evictions += int64(evicted)
		m.stats.LastEviction = now
	}
}

// GetAllStats returns a snapshot of cache statistics for monitoring dashboards.
func (m *CacheManager) GetAllStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var newest time.Time
	var oldest time.Time
	for _, entry := range m.items {
		if newest.IsZero() || entry.CreatedAt.After(newest) {
			newest = entry.CreatedAt
		}
		if oldest.IsZero() || entry.CreatedAt.Before(oldest) {
			oldest = entry.CreatedAt
		}
	}

	return map[string]interface{}{
		"items":          m.stats.Items,
		"hits":           m.stats.Hits,
		"misses":         m.stats.Misses,
		"evictions":      m.stats.Evictions,
		"last_eviction":  m.stats.LastEviction,
		"oldest_insert":  oldest,
		"newest_insert":  newest,
		"entries_active": len(m.items),
	}
}
