package lru

import "container/list"

/**
 * @Author Charles
 * @Date 3:53 PM 10/9/2022
 **/

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64 // 0 means no limit
	nbytes    int64
	list      *list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		list:      list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if e, ok := c.cache[key]; ok {
		c.list.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nbytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else {
		e := c.list.PushFront(&entry{key, value})
		c.cache[key] = e
		c.nbytes += int64(len(key) + value.Len())
	}
	if c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	e := c.list.Back()
	if e != nil {
		c.list.Remove(e)
		kv := e.Value.(*entry)
		c.nbytes -= int64(len(kv.key) + kv.value.Len())
		delete(c.cache, kv.key)
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.cache[key]; ok {
		kv := e.Value.(*entry)
		c.list.MoveToFront(e)
		return kv.value, true
	}
	return
}

func (c *Cache) Len() int {
	return c.list.Len()
}
