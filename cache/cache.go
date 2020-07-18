package cache

import "sync"

type BaseCache interface {
	Set(string, []byte) error
	Get(string) ([]byte, error)
	Del(string) error
	GetStat() Status
}

//缓存使用情况
type Status struct {
	Count     int64
	KeySize   int64
	ValueSize int64
}

func (s *Status) add(k string, v []byte) {
	s.Count += 1
	s.KeySize += int64(len(k))
	s.ValueSize += int64(len(v))
}

func (s *Status) del(k string, v []byte) {
	s.Count -= 1
	s.KeySize -= int64(len(k))
	s.ValueSize -= int64(len(v))
}

type InMemoryCache struct {
	c map[string] []byte
	mutex sync.RWMutex
	Status
}

func NewInMemoryCache () *InMemoryCache {
	return &InMemoryCache{make(map[string][]byte), sync.RWMutex{}, Status{}}
}

func (c *InMemoryCache) Get(k string) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.c[k], nil
}

func (c *InMemoryCache) Set(k string, v []byte) error{
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

func (c *InMemoryCache) GetStatus() Status {
	return c.Status
}