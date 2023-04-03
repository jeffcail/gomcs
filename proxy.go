package gomcs

import "time"

type cacheImpl struct {
	memCache Cache
}

// NewCache
func NewCache() *cacheImpl {
	return &cacheImpl{memCache: NewMemCache()}
}

// SetMaxMemory
func (c *cacheImpl) SetMaxMemory(size string) bool {
	return c.memCache.SetMaxMemory(size)
}

// Set
func (c *cacheImpl) Set(key string, val interface{}, expire ...time.Duration) bool {
	expireT := time.Second * 0
	if len(expire) > 0 {
		expireT = expire[0]
	}
	return c.memCache.Set(key, val, expireT)
}

// Get
func (c *cacheImpl) Get(key string) (interface{}, bool) {
	return c.memCache.Get(key)
}

// Del
func (c *cacheImpl) Del(key string) bool {
	return c.memCache.Del(key)
}

// Exists
func (c *cacheImpl) Exists(key string) bool {
	return c.memCache.Exists(key)
}

// Flush
func (c *cacheImpl) Flush() bool {
	return c.memCache.Flush()
}

// Keys
func (c *cacheImpl) Keys() int64 {
	return c.memCache.Keys()
}
