package cache

import (
	"scheduler_rpc/internal"
	"sync"
)

// ThreadSafeCache 是一个线程安全的本地缓存
type Cache struct {
	Cache map[string]*internal.NodeResource
	lock  sync.Mutex
}

func NewCache() *Cache {
	nodeView := &Cache{}
	nodeView.Cache = make(map[string]*internal.NodeResource, 0)
	return nodeView
}

// Set 将键值对放入缓存
func (c *Cache) Set(key string, value *internal.NodeResource) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Cache[key] = value
}

// Get 根据键获取缓存中的值
func (c *Cache) Get(key string) *internal.NodeResource {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.Cache[key]
}

// Delete 根据键删除缓存中的值
func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.Cache, key)
}
