package bicache_test

import (
	"testing"
	"time"

	"github.com/mtnmunuklu/bicache"
)

func TestBasicCacheOperations(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)

	// Set operation
	cache.Set("key1", "value1", time.Second*10)
	value, exists := cache.Get("key1")
	if !exists || value != "value1" {
		t.Errorf("Set operation failed.")
	}

	// Get operation
	value, exists = cache.Get("key1")
	if !exists || value != "value1" {
		t.Errorf("Get operation failed.")
	}

	// Delete operation
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	if exists {
		t.Errorf("Delete operation failed.")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)

	// Set with expiration
	cache.Set("key1", "value1", time.Millisecond*500)
	time.Sleep(time.Millisecond * 600)
	_, exists := cache.Get("key1")
	if exists {
		t.Errorf("Cache expiration failed.")
	}
}

func TestCachePolicy(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)
	cache.SetCachePolicy(func(key interface{}, entry bicache.CacheEntry) bool {
		// Custom cache policy example
		return entry.Value.(string) != "blocked"
	})

	// Set operation allowed by the custom cache policy
	cache.Set("key1", "allowed", time.Second*10)
	value, exists := cache.Get("key1")
	if !exists || value != "allowed" {
		t.Errorf("Cache policy failed.")
	}

	// Set operation blocked by the custom cache policy
	cache.Set("key2", "blocked", time.Second*10)
	_, exists = cache.Get("key2")
	if exists {
		t.Errorf("Cache policy failed.")
	}
}

func TestGlobalExpiration(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)
	cache.SetGlobalExpiration(time.Millisecond * 500)

	// Set operation with global expiration
	cache.Set("key1", "value1", time.Second*10)
	time.Sleep(time.Millisecond * 600)
	_, exists := cache.Get("key1")
	if exists {
		t.Errorf("Global expiration failed.")
	}
}

func TestCacheEventHandler(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)
	var eventTriggered bool

	cache.SetCacheEventHandler(func(event bicache.CacheEvent, key interface{}, entry bicache.CacheEntry) {
		if event == bicache.CacheEventSet && key == "key1" {
			eventTriggered = true
		}
	})

	// Set operation triggering cache event handler
	cache.Set("key1", "value1", time.Second*10)
	time.Sleep(time.Millisecond * 100) // Wait for the event handler to execute
	if !eventTriggered {
		t.Errorf("Cache event handler failed.")
	}
}

func TestUpdateStrategy(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)
	cache.SetUpdateStrategy(func(key interface{}, oldValue interface{}) interface{} {
		// Custom update strategy example
		return oldValue.(string) + "_updated"
	})

	// Set operation with update strategy
	cache.Set("key1", "value1", time.Second*10)
	cache.Set("key1", "value1", time.Second*10) // Trigger update strategy
	value, exists := cache.Get("key1")
	if !exists || value != "value1_updated" {
		t.Errorf("Update strategy failed.")
	}
}

func TestCompressionDecompression(t *testing.T) {
	cache := bicache.NewBiCache(5, time.Second)
	cache.SetCompression(func(data []byte) ([]byte, error) {
		// Custom compression example
		return []byte("compressed"), nil
	}, func(data []byte) ([]byte, error) {
		// Custom decompression example
		return []byte("decompressed"), nil
	})

	// Set operation with compression
	cache.Set("key1", "value1", time.Second*10)
	value, exists := cache.Get("key1")
	if !exists || value != "decompressed" {
		t.Errorf("Compression and decompression failed.")
	}
}
