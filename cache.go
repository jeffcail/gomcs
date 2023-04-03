package gomcs

import (
	"fmt"
	"sync"
	"time"
)

type MemoryCache struct {
	maxMemorySize           int64
	maxMemorySizeS          string
	currentMemorySize       int64
	cmap                    map[string]*CMapValue
	mlock                   sync.RWMutex  // 读写锁
	clearExpireKeysInterval time.Duration // 清除过期key时间周期
}

type CMapValue struct {
	value   interface{}   // value值
	expireT time.Time     // 过期时间
	expire  time.Duration // 有效时长
	size    int64         // value大小
}

func NewMemCache() Cache {
	m := &MemoryCache{
		cmap:                    make(map[string]*CMapValue),
		clearExpireKeysInterval: time.Second * 10,
	}
	go m.clearExpireKeys()
	return m
}

var defaultMemorySize int64 = 100 // 100MB

// SetMaxMemory size: 1KB 100KB 1MB 100MB 1GB default 100MB
func (m *MemoryCache) SetMaxMemory(size string) bool {
	m.maxMemorySize, m.maxMemorySizeS = CovertSize(size, defaultMemorySize)
	fmt.Println(m.maxMemorySize, m.maxMemorySizeS)
	return false
}

// Set
func (m *MemoryCache) Set(key string, val interface{}, expire time.Duration) bool {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	v := &CMapValue{
		value:   val,
		expireT: time.Now().Add(expire),
		expire:  expire,
		size:    CalculateSize(val),
	}
	m.del(key)
	m.add(key, v)
	if m.currentMemorySize > m.maxMemorySize {
		m.del(key)
		panic("max memory size is not enough")
	}
	return true
}

func (m *MemoryCache) get(key string) (*CMapValue, bool) {
	val, ok := m.cmap[key]
	return val, ok
}

func (m *MemoryCache) del(key string) {
	tmp, ok := m.get(key)
	if ok && tmp != nil {
		m.currentMemorySize -= tmp.size
		delete(m.cmap, key)
	}
}

func (m *MemoryCache) add(key string, val *CMapValue) {
	m.cmap[key] = val
	m.currentMemorySize += val.size
}

// Get
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.mlock.RLock()
	defer m.mlock.RUnlock()
	v, ok := m.get(key)
	if ok {
		if v.expire != 0 && v.expireT.Before(time.Now()) {
			m.del(key)
			return nil, false
		}
	}
	return v.value, ok
}

// Del
func (m *MemoryCache) Del(key string) bool {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	m.del(key)
	return true
}

// Exists
func (m *MemoryCache) Exists(key string) bool {
	m.mlock.RLock()
	defer m.mlock.RUnlock()
	_, ok := m.cmap[key]
	return ok
}

// Flush 清空所有的key
func (m *MemoryCache) Flush() bool {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	m.cmap = make(map[string]*CMapValue, 0)
	m.currentMemorySize = 0
	return true
}

// Keys
func (m *MemoryCache) Keys() int64 {
	m.mlock.RLock()
	defer m.mlock.RUnlock()
	return int64(len(m.cmap))
}

// clearExpireKeys
func (m *MemoryCache) clearExpireKeys() {
	tk := time.NewTicker(m.clearExpireKeysInterval)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			for k, item := range m.cmap {
				if item.expire != 0 && time.Now().After(item.expireT) {
					m.mlock.Lock()
					m.del(k)
					m.mlock.Unlock()
				}
			}
		}
	}
}
