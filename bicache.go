package bicache

import (
	"encoding/gob"
	"reflect"
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
		serializer:        nil,
		deserializer:      nil,
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

		if c.decompression != nil {
			byteValue, ok := entry.Value.([]byte)
			if !ok {
				c.metrics.SetError++
				return nil, false
			}

			// Decompress the value
			decompressedValue, err := c.decompression(byteValue)
			if err != nil {
				c.metrics.SetError++
				return nil, false
			}
			entry.Value = decompressedValue
		}

		if c.deserializer != nil {
			// Decode the value
			decodedValue, err := c.decodeValue(entry.Value)
			if err != nil {
				c.metrics.SetError++
				return nil, false
			}
			entry.Value = decodedValue
		}

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

	entry := CacheEntry{Accessed: time.Now()}

	// Encode the value
	if c.serializer != nil {
		encodedValue, err := c.encodeValue(value)
		if err != nil {
			c.metrics.SetError++
			return
		}
		entry.Value = encodedValue
	}

	// Compress the value
	if c.compression != nil {
		compressedValue, err := c.compressValue(entry.Value)
		if err != nil {
			c.metrics.SetError++
			return
		}
		entry.Value = compressedValue
	}

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

func (c *BiCache) compressValue(value interface{}) (interface{}, error) {
	if c.compression == nil {
		// Compression is not enabled, return the original value
		return value, nil
	}

	switch val := value.(type) {
	case []byte:
		// If the value is a byte slice, apply compression
		compressedValue, err := c.compression(val)
		if err != nil {
			return nil, err
		}
		return compressedValue, nil
	default:
		// If the value is not a byte slice, return the original value
		return value, nil
	}
}

func (c *BiCache) decodeValue(encodedValue interface{}) (interface{}, error) {
	value := reflect.New(reflect.TypeOf(encodedValue))
	if err := c.deserializer.DecodeValue(value); err != nil {
		return nil, err
	}
	return value.Elem().Interface(), nil
}

func (c *BiCache) encodeValue(value interface{}) (interface{}, error) {
	valueToEncode := reflect.ValueOf(value)
	if err := c.serializer.EncodeValue(valueToEncode); err != nil {
		return nil, err
	}
	return value, nil
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

// cleanup method cleans up the currently valid items in the cache.
func (c *BiCache) cleanup() {
	// Get the current time
	now := time.Now()

	// Check each item in the cache
	for key, entry := range c.cacheMap {
		// Determine the expiration time to use for comparison
		var expiration time.Time
		if c.globalExpiration > 0 {
			// If globalExpiration is greater than 0, use the item's Accessed time plus globalExpiration
			expiration = entry.Accessed.Add(c.globalExpiration)
		} else {
			// If globalExpiration is 0 or negative, use the item's Expiration directly
			expiration = entry.Expiration
		}

		// If the calculated expiration is in the past, clean up this item.
		if expiration.Before(now) {
			delete(c.cacheMap, key)
			c.metrics.EntriesCount = int64(len(c.cacheMap))

			// If a cache event handler is defined, call it when the item is deleted.
			if c.cacheEventHandler != nil {
				go c.cacheEventHandler(CacheEventDelete, key, CacheEntry{})
			}
		}
	}
}
