package main

import "sync"

type Users struct {
	mu sync.RWMutex
	m  map[string]int
}

var users = &Users{
	m: make(map[string]int),
}

func (s *Users) Set(key string, value int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *Users) Get(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[key]
}

func (s *Users) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}
