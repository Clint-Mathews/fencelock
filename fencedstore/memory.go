package fencedstore

import (
	"context"
	"sync"

	"github.com/Clint-Mathews/fencelock/lock"
)

// Memory is an in-memory FencedResource: it tracks the highest fencing
// token seen per key and rejects any Write with a lower token.
type Memory struct {
	mu        sync.Mutex
	lastToken map[string]int64
	data      map[string][]byte
}

func MewMemory() *Memory {
	return &Memory{
		lastToken: make(map[string]int64),
		data:      make(map[string][]byte),
	}
}

func (m *Memory) Write(ctx context.Context, key string, token int64, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token < m.lastToken[key] {
		return lock.ErrStaleToken
	}
	m.lastToken[key] = token
	m.data[key] = payload

	return nil
}

// Helper method
func (m *Memory) Get(key string) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key]
}
