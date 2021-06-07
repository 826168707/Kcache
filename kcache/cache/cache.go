package cache

import (
	"sync"
	"time"

	"kcache/cache/lru"
)

// cache LRU的并发控制
type cache struct {
	mu   sync.Mutex
	lru  *lru.Cache
	size int
}

func (c *cache) addWithExpire(key string, value ByteView, expire time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.size, nil)
	}
	c.lru.AddWithExpire(key, value, expire)
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.size, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) setDefaultExpire(expire time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.NewWithExpire(c.size, expire, nil)
		return
	}
	c.lru.SetDefaultExpire(expire)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return ByteView{}, false
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}
