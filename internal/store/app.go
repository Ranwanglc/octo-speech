package store

import (
	"database/sql"
	"sync"
	"time"
)

type AppInfo struct {
	AppID  string
	Status int
}

type cacheEntry struct {
	info      *AppInfo
	expiresAt time.Time
}

type AppStore struct {
	db       *sql.DB
	mu       sync.RWMutex
	cache    map[string]cacheEntry
	ttl      time.Duration
	maxSize  int
}

func NewAppStore(db *sql.DB, ttlSeconds int) *AppStore {
	return &AppStore{
		db:      db,
		cache:   make(map[string]cacheEntry),
		ttl:     time.Duration(ttlSeconds) * time.Second,
		maxSize: 10000,
	}
}

func (s *AppStore) Authenticate(apiKey string) (*AppInfo, error) {
	s.mu.RLock()
	if entry, ok := s.cache[apiKey]; ok && time.Now().Before(entry.expiresAt) {
		s.mu.RUnlock()
		return entry.info, nil
	}
	s.mu.RUnlock()

	var appID string
	var status int
	err := s.db.QueryRow(
		"SELECT app_id, status FROM app_registry WHERE api_key = ?",
		apiKey,
	).Scan(&appID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	info := &AppInfo{AppID: appID, Status: status}

	s.mu.Lock()
	if len(s.cache) >= s.maxSize {
		s.evictOldest()
	}
	s.cache[apiKey] = cacheEntry{
		info:      info,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return info, nil
}

func (s *AppStore) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range s.cache {
		if first || v.expiresAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.expiresAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(s.cache, oldestKey)
	}
}
