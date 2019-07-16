package memdb

import (
	"strings"
	"sync"
	"time"
)

type Token interface {
	Name() string
	Set(key string, data interface{})
	Get(key string, delay bool) (interface{}, bool)
	Del(key string) bool
	Lst(key string) []interface{}
	Permanent(key string, val bool) bool
}

func NewToken(expMinutes int64, name string) Token {
	return newToken(expMinutes, 5*time.Minute, name)
}

func newToken(expMinutes int64, expCheckInterval time.Duration, name string) Token {
	instance := &innerToken{name: name}
	instance.exp = time.Duration(expMinutes) * time.Minute
	instance.items = make(map[string]*tokenTime)

	if expMinutes > 0 {
		go func(interval time.Duration) {
			instance.checkExpiration(interval)
		}(expCheckInterval)
	}

	return instance
}

type tokenTime struct {
	data      interface{}
	exp       time.Time
	permanent bool
}

type innerToken struct {
	sync.RWMutex

	items map[string]*tokenTime
	exp   time.Duration
	name  string
}

func (s *innerToken) Name() string {
	return s.name
}

func (s *innerToken) Set(key string, data interface{}) {
	s.Lock()
	defer s.Unlock()

	s.items[key] = &tokenTime{
		data:      data,
		exp:       time.Now().Add(s.exp),
		permanent: false,
	}
}

func (s *innerToken) Get(key string, delay bool) (interface{}, bool) {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.items[key]
	if !ok {
		return nil, false
	}

	if delay {
		v.exp = time.Now().Add(s.exp)
	}

	return v.data, true
}

func (s *innerToken) Del(key string) bool {
	s.Lock()
	defer s.Unlock()

	_, ok := s.items[key]
	if ok {
		delete(s.items, key)
	}

	return ok
}

func (s *innerToken) Lst(key string) []interface{} {
	s.RLock()
	defer s.RUnlock()

	items := make([]interface{}, 0)
	for k, v := range s.items {
		if len(key) > 0 {
			if !strings.Contains(k, key) {
				continue
			}
		}
		items = append(items, v.data)
	}

	return items
}

func (s *innerToken) Permanent(key string, val bool) bool {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.items[key]
	if !ok {
		return false
	}
	v.permanent = val

	return true
}

func (s *innerToken) checkExpiration(interval time.Duration) {
	for {
		time.Sleep(interval)
		s.deleteExpiration()
	}
}

func (s *innerToken) deleteExpiration() {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	for k, v := range s.items {
		if !v.permanent {
			if v.exp.Before(now) {
				delete(s.items, k)
			}
		}
	}
}
