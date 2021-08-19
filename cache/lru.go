package cache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	lruNodeList *list.List
	cache       map[string]*list.Element

	mutex    sync.RWMutex
	OnMerged func(key string, value Value)

	Status
}

func NewLRUCache(maxBytes int64, onMerged func(str string, value Value)) *LRUCache {
	return &LRUCache{&list.List{}, make(map[string]*list.Element), sync.RWMutex{}, onMerged, Status{}}
}

func (c *LRUCache) Get(key string) (v Value, err error) {

	if ele, ok := c.cache[key]; ok {
		c.lruNodeList.MoveToFront(ele)
		kv := ele.Value.(*entry)

		return kv.value, nil
	}
	return

}

func (c *LRUCache) Set(key string, v Value) error {

	// change exists keys value
	if ele, ok := c.cache[key]; ok {
		c.lruNodeList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.Status.ValueSize += kv.value.Len()
		kv.value = v
	} else {
		ele := c.lruNodeList.PushFront(&entry{key, v})
		c.cache[key] = ele
		c.Status.add(key, v)
	}
	for c.isOverflow() {
		c.RemoveOldestEntry()
	}
	return nil

}

func (c *LRUCache) Del (key string) error {
	ele := c.lruNodeList.Back()
	if ele == nil {
		return nil
	}
	kv := ele.Value.(*entry)

	delete(c.cache, kv.key)
	c.Status.del(kv.key, kv.value)
	if c.OnMerged != nil {
		c.OnMerged(kv.key, kv.value)
	}
	return nil
}

func (c *LRUCache) GetStatus() Status{
	return c.Status
}


func (c *LRUCache) RemoveOldestEntry() {
	// return the last entry in the node list
	ele := c.lruNodeList.Back()
	if ele == nil {
		return
	}

	c.lruNodeList.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)

	c.Status.del(kv.key, kv.value)
	// execute hook method on GC
	if c.OnMerged != nil {
		c.OnMerged(kv.key, kv.value)
	}

}

