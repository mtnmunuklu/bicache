package bicache

import (
	"encoding/gob"
	"sync"
	"time"
)

type CacheEvent int

const (
	CacheEventSet CacheEvent = iota
	CacheEventDelete
)

type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
	Accessed   time.Time
}

type CacheMetrics struct {
	Hits         int64
	Misses       int64
	SetSuccess   int64
	SetError     int64
	EntriesCount int64
}

type CachePolicyFunc func(key interface{}, entry CacheEntry) bool

type CacheEventHandlerFunc func(event CacheEvent, key interface{}, entry CacheEntry)

type UpdateStrategyFunc func(key interface{}, oldValue interface{}) interface{}

type CompressionFunc func(data []byte) ([]byte, error)
type DecompressionFunc func(data []byte) ([]byte, error)

type BiCache struct {
	mu                sync.RWMutex
	capacity          int
	cacheMap          map[interface{}]CacheEntry
	metrics           CacheMetrics
	cleanupTicker     *time.Ticker
	serializer        *gob.Encoder
	deserializer      *gob.Decoder
	cachePolicy       CachePolicyFunc
	globalExpiration  time.Duration
	cacheEventHandler CacheEventHandlerFunc
	updateStrategy    UpdateStrategyFunc
	compression       CompressionFunc
	decompression     DecompressionFunc
}

func NewBiCache(capacity int, cleanupInterval time.Duration) *BiCache {
	cache := &BiCache{
		capacity:          capacity,
		cacheMap:          make(map[interface{}]CacheEntry),
		cleanupTicker:     time.NewTicker(cleanupInterval),
		serializer:        gob.NewEncoder(nil),
		deserializer:      gob.NewDecoder(nil),
		cachePolicy:       nil, // Cache policy can be set using SetCachePolicy method
		globalExpiration:  0,   // Global expiration can be set using SetGlobalExpiration method
		cacheEventHandler: nil, // Cache event handler can be set using SetCacheEventHandler method
		updateStrategy:    nil, // Update strategy can be set using SetUpdateStrategy method
		compression:       nil, // Compression can be set using SetCompression method
		decompression:     nil, // Decompression can be set using SetDecompression method
	}

	go cache.periodicCleanup()

	return cache
}

func (c *BiCache) Get(key interface{}) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cacheMap[key]
	if exists {
		entry.Accessed = time.Now()

		if entry.Expiration.IsZero() || time.Now().Before(entry.Expiration) {
			c.metrics.Hits++
			return entry.Value, true
		}

		delete(c.cacheMap, key)
		c.metrics.EntriesCount = int64(len(c.cacheMap))
	}

	c.metrics.Misses++
	return nil, false
}

func (c *BiCache) Set(key interface{}, value interface{}, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{Value: value, Accessed: time.Now()}

	if expiration > 0 {
		entry.Expiration = time.Now().Add(expiration)
	}

	if c.cachePolicy != nil && !c.cachePolicy(key, entry) {
		return
	}

	if c.updateStrategy != nil {
		oldValue, exists := c.cacheMap[key]
		if exists {
			entry.Value = c.updateStrategy(key, oldValue.Value)
		}
	}

	c.cacheMap[key] = entry
	c.metrics.SetSuccess++
	c.metrics.EntriesCount = int64(len(c.cacheMap))

	if len(c.cacheMap) > c.capacity {
		c.cleanup()
	}

	if c.cacheEventHandler != nil {
		go c.cacheEventHandler(CacheEventSet, key, entry)
	}
}

func (c *BiCache) Delete(key interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cacheMap, key)
	c.metrics.EntriesCount = int64(len(c.cacheMap))

	if c.cacheEventHandler != nil {
		go c.cacheEventHandler(CacheEventDelete, key, CacheEntry{})
	}
}

func (c *BiCache) GetMetrics() CacheMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metrics
}

func (c *BiCache) SetSerializer(serializer *gob.Encoder) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.serializer = serializer
}

func (c *BiCache) SetDeserializer(deserializer *gob.Decoder) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deserializer = deserializer
}

func (c *BiCache) SetCapacity(capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.capacity = capacity

	if len(c.cacheMap) > c.capacity {
		c.cleanup()
	}
}

func (c *BiCache) SetCachePolicy(policy CachePolicyFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cachePolicy = policy
}

func (c *BiCache) SetGlobalExpiration(expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.globalExpiration = expiration
}

func (c *BiCache) SetCacheEventHandler(handler CacheEventHandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cacheEventHandler = handler
}

func (c *BiCache) SetUpdateStrategy(strategy UpdateStrategyFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updateStrategy = strategy
}

func (c *BiCache) SetCompression(compression CompressionFunc, decompression DecompressionFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.compression = compression
	c.decompression = decompression
}

func (c *BiCache) periodicCleanup() {
	for range c.cleanupTicker.C {
		c.cleanup()
	}
}

func (c *BiCache) cleanup() {
	now := time.Now()

	for key, entry := range c.cacheMap {
		if entry.Expiration.IsZero() || now.Before(entry.Expiration) {
			continue
		}

		delete(c.cacheMap, key)
		c.metrics.EntriesCount = int64(len(c.cacheMap))

		if c.cacheEventHandler != nil {
			go c.cacheEventHandler(CacheEventDelete, key, CacheEntry{})
		}
	}
}
