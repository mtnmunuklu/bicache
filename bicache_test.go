package bicache

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"
)

func TestBiCache_SetGet(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second*20)

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is retrieved correctly
	if !found || result.(string) != "value1" {
		t.Errorf("Set-Get test failed. Expected: 'value1', Got: '%v'", result)
	}
}

func TestBiCache_SetGetExpired(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Set an expired value in the cache
	cache.Set("key1", "value1", -time.Second)

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is not found due to expiration
	if found || result != nil {
		t.Errorf("Set-Get Expired test failed. Expected: not found, Got: '%v'", result)
	}
}

func TestBiCache_SetGetWithGlobalExpiration(t *testing.T) {
	cache := NewBiCache(5, time.Second)
	cache.SetGlobalExpiration(time.Second * 2)

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second)

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is retrieved correctly
	if !found || result.(string) != "value1" {
		t.Errorf("Set-Get with Global Expiration test failed. Expected: 'value1', Got: '%v'", result)
	}

	// Wait for the global expiration to occur
	time.Sleep(time.Second * 3)

	// Get the value again and check if it's expired
	result, found = cache.Get("key1")
	if found || result != nil {
		t.Errorf("Set-Get with Global Expiration (expired) test failed. Expected: not found, Got: '%v'", result)
	}
}

func TestBiCache_Delete(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second)

	// Delete the value from the cache
	cache.Delete("key1")

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is not found after deletion
	if found || result != nil {
		t.Errorf("Delete test failed. Expected: not found, Got: '%v'", result)
	}
}

func TestBiCache_CachePolicy(t *testing.T) {
	cache := NewBiCache(5, time.Second)
	cache.SetCachePolicy(func(key interface{}, entry CacheEntry) bool {
		return key.(int) > 0 // Allow only positive keys
	})

	// Set a value with a positive key
	cache.Set(1, "positive", time.Second)

	// Set a value with a negative key
	cache.Set(-1, "negative", time.Second)

	// Get the value with a positive key
	result, found := cache.Get(1)
	if !found || result.(string) != "positive" {
		t.Errorf("CachePolicy test (positive key) failed. Expected: 'positive', Got: '%v'", result)
	}

	// Get the value with a negative key
	result, found = cache.Get(-1)
	if found || result != nil {
		t.Errorf("CachePolicy test (negative key) failed. Expected: not found, Got: '%v'", result)
	}
}

func TestBiCache_UpdateStrategy(t *testing.T) {
	cache := NewBiCache(5, time.Second)
	cache.SetUpdateStrategy(func(key interface{}, oldValue interface{}) interface{} {
		return "updated_" + oldValue.(string)
	})

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second)

	// Set the value again to trigger the update strategy
	cache.Set("key1", "value1", time.Second)

	// Get the updated value from the cache
	result, found := cache.Get("key1")

	// Check if the value is updated correctly
	if !found || result.(string) != "updated_value1" {
		t.Errorf("UpdateStrategy test failed. Expected: 'updated_value1', Got: '%v'", result)
	}
}

func TestBiCache_Compression(t *testing.T) {
	cache := NewBiCache(5, time.Second)
	cache.SetCompression(func(data []byte) ([]byte, error) {
		return bytes.ToUpper(data), nil
	}, nil)

	// Set a lowercase value in the cache
	cache.Set("key1", "value1", time.Second)

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is compressed correctly
	expected := []byte("VALUE1")
	if !found || !bytes.Equal(result.([]byte), expected) {
		t.Errorf("Compression test failed. Expected: '%v', Got: '%v'", expected, result)
	}
}

func TestBiCache_Serialization(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Initialize a buffer for serialization
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	decoder := gob.NewDecoder(&buf)

	// Set a custom serializer and deserializer
	cache.SetSerializer(encoder)
	cache.SetDeserializer(decoder)

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second*20)

	// Get the value from the cache
	result, found := cache.Get("key1")

	// Check if the value is retrieved correctly with custom serialization
	if !found || result.(string) != "value1" {
		t.Errorf("Serialization test failed. Expected: 'value1', Got: '%v'", result)
	}
}

func TestBiCache_Metrics(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second)

	// Get metrics
	metrics := cache.GetMetrics()

	// Check if metrics are updated correctly
	if metrics.SetSuccess != 1 || metrics.Hits != 0 || metrics.Misses != 0 {
		t.Errorf("Metrics test failed. Expected: SetSuccess=1, Hits=0, Misses=0. Got: SetSuccess=%v, Hits=%v, Misses=%v",
			metrics.SetSuccess, metrics.Hits, metrics.Misses)
	}
}

func TestBiCache_PeriodicCleanup(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	// Set a value in the cache
	cache.Set("key1", "value1", -time.Second)

	// Wait for the periodic cleanup to occur
	time.Sleep(time.Second * 2)

	// Get the value again and check if it's expired
	result, found := cache.Get("key1")
	if found || result != nil {
		t.Errorf("Periodic Cleanup test failed. Expected: not found, Got: '%v'", result)
	}
}

func TestBiCache_CacheEventHandler(t *testing.T) {
	cache := NewBiCache(5, time.Second)

	var eventReceived bool
	var receivedKey interface{}
	var receivedEvent CacheEvent
	var receivedEntry CacheEntry

	// Set a custom event handler
	cache.SetCacheEventHandler(func(event CacheEvent, key interface{}, entry CacheEntry) {
		eventReceived = true
		receivedKey = key
		receivedEvent = event
		receivedEntry = entry
	})

	// Set a value in the cache
	cache.Set("key1", "value1", time.Second)

	// Wait for the event handler to be called
	time.Sleep(time.Millisecond * 100)

	// Check if the event was received correctly
	if !eventReceived || receivedKey != "key1" || receivedEvent != CacheEventSet || receivedEntry.Value != "value1" {
		t.Errorf("CacheEventHandler test failed. Event not received or incorrect values. Received: key=%v, event=%v, entry=%v",
			receivedKey, receivedEvent, receivedEntry)
	}
}
