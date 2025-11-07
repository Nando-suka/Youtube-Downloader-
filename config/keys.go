package config

import (
	"sync"
	"time"
)

type APIKeyManager struct {
	keys         []string
	currentIndex int
	usageCount   map[string]int
	lastReset    time.Time
	mutex        sync.RWMutex
}

func NewAPIKeyManager(keys []string) *APIKeyManager {
	usageCount := make(map[string]int)
	for _, key := range keys {
		usageCount[key] = 0
	}

	return &APIKeyManager{
		keys:       keys,
		usageCount: usageCount,
		lastReset:  time.Now(),
	}
}

func (m *APIKeyManager) GetCurrentKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.keys) == 0 {
		return ""
	}

	return m.keys[m.currentIndex]
}

func (m *APIKeyManager) RotateKey() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.keys) == 0 {
		return ""
	}

	m.currentIndex = (m.currentIndex + 1) % len(m.keys)
	return m.keys[m.currentIndex]
}

func (m *APIKeyManager) RecordUsage() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	currentKey := m.keys[m.currentIndex]
	m.usageCount[currentKey]++

	// Reset counters every hour
	if time.Since(m.lastReset) > time.Hour {
		for key := range m.usageCount {
			m.usageCount[key] = 0
		}
		m.lastReset = time.Now()
	}
}

func (m *APIKeyManager) GetUsageStats() map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]int)
	for key, count := range m.usageCount {
		stats[key] = count
	}
	return stats
}

func (m *APIKeyManager) GetAvailableKeysCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.keys)
}
