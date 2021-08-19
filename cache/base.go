package cache

import (
	"sync"
)

// return entry len
type Value interface {
	Len() int64
}

type Cache interface {
	Set(string, Value) error
	Get(string) (Value, error)
	Del(string) error
	GetStat() Status
}

// single entry
type entry struct {
	key   string
	value Value
}

type Status struct {
	Count     int64
	KeySize   int64
	ValueSize int64
	LimitSize int64
}

func (s *Status) add(k string, v Value) {
	s.Count += 1
	s.KeySize += int64(len([]byte(k)))
	s.ValueSize += v.Len()
}

func (s *Status) del(k string, v Value) {
	s.Count -= 1
	s.KeySize -= v.Len()
	s.ValueSize -= v.Len()
}

func (s *Status) isOverflow() bool {
	if s.LimitSize > 0 && s.ValueSize > s.LimitSize {
		return true
	}
	return false
}

type SCache struct {
	c     map[string]Value
	mutex sync.RWMutex
	Status
}

func NewSimpleCache() *SCache {
	return &SCache{make(map[string]Value), sync.RWMutex{}, Status{}}
}

func (c *SCache) Get(k string) (Value, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.c[k], nil
}

func (c *SCache) Set(k string, v Value) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, exist := c.c[k]
	if exist {
		delete(c.c, k)
	}
	c.c[k] = v
	c.Status.add(k, v)
	return nil
}

func (c *SCache) GetStatus() Status {
	return c.Status
}

