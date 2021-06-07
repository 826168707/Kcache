package lru

import (
	"container/list"
	"time"
)

type Value interface {
	Len() int
}

type entry struct {
	key      string
	value    Value
	deadline *time.Time
}

//Cache 带TTL的LRU缓存淘汰策略
type Cache struct {
	size      int // LRU所能容纳的最大数据量
	nodeList  *list.List
	items     map[string]*list.Element
	expire    time.Duration
	onEvicted func(key string, value Value)
}

func New(size int, onEvicted func(string, Value)) *Cache {
	return NewWithExpire(size, 0, onEvicted)
}

func NewWithExpire(size int, expire time.Duration, onEvicted func(string, Value)) *Cache {
	return &Cache{
		size:      size,
		nodeList:  list.New(),
		items:     make(map[string]*list.Element),
		expire:    expire,
		onEvicted: onEvicted,
	}
}

func (e *entry) isExpired() bool {
	if e.deadline == nil {
		return false
	}
	return time.Now().After(*e.deadline)
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.items[key]; ok {
		// 检查是否过期
		kv := ele.Value.(*entry)
		if kv.isExpired() {
			c.Remove(key)
			return nil, false
		}
		c.nodeList.MoveToFront(ele)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) removeElement(ele *list.Element) {
	c.nodeList.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.items, kv.key)
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}

func (c *Cache) removeOldest() {
	ele := c.nodeList.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) Remove(key string) bool {
	if ele, ok := c.items[key]; ok {
		c.removeElement(ele)
		return true
	}
	return false
}

// 添加带有生命周期的数据，返回值表示是否触发removeOldest
func (c *Cache) AddWithExpire(key string, value Value, expire time.Duration) bool {
	var deadline *time.Time = nil
	if expire > 0 {
		dl := time.Now().Add(expire)
		deadline = &dl
	} else if c.expire > 0 {
		dl := time.Now().Add(c.expire)
		deadline = &dl
	}

	if ele, ok := c.items[key]; ok {
		c.nodeList.MoveToFront(ele)
		ele.Value.(*entry).value = value
		ele.Value.(*entry).deadline = deadline
		return true
	}

	ent := &entry{
		key:      key,
		value:    value,
		deadline: deadline,
	}
	ele := c.nodeList.PushFront(ent)
	c.items[key] = ele

	evict := c.nodeList.Len() > c.size
	if evict {
		c.removeOldest()
	}
	return evict
}

func (c *Cache) Add(key string, value Value) bool {
	return c.AddWithExpire(key, value, 0)
}

func (c *Cache) SetDefaultExpire(expire time.Duration) {
	c.expire = expire
}
