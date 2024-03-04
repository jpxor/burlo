package main

import (
	"sync"
)

// RWMap is a generic map wrapper with RWLock.
type RWMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// NewRWMap creates a new RWMap.
func NewRWMap[K comparable, V any]() *RWMap[K, V] {
	return &RWMap[K, V]{
		data: make(map[K]V),
	}
}

// Get retrieves a value from the map.
func (m *RWMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Set adds or updates a value in the map.
func (m *RWMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Delete removes a value from the map.
func (m *RWMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Len returns the number of elements in the map.
func (m *RWMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func RwMapExample() {
	// Example usage:
	myMap := NewRWMap[string, int]()
	myMap.Set("apple", 5)
	myMap.Set("banana", 10)

	val, ok := myMap.Get("apple")
	if ok {
		println("Value for 'apple':", val)
	}

	myMap.Delete("banana")
	println("Map length:", myMap.Len())
}
