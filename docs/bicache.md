# bicache
--
    import "."


## Usage

#### type BiCache

```go
type BiCache struct {
}
```


#### func  NewBiCache

```go
func NewBiCache(capacity int, cleanupInterval time.Duration) *BiCache
```

#### func (*BiCache) Delete

```go
func (c *BiCache) Delete(key interface{})
```

#### func (*BiCache) Get

```go
func (c *BiCache) Get(key interface{}) (interface{}, bool)
```

#### func (*BiCache) GetMetrics

```go
func (c *BiCache) GetMetrics() CacheMetrics
```

#### func (*BiCache) Set

```go
func (c *BiCache) Set(key interface{}, value interface{}, expiration time.Duration)
```

#### func (*BiCache) SetCacheEventHandler

```go
func (c *BiCache) SetCacheEventHandler(handler CacheEventHandlerFunc)
```

#### func (*BiCache) SetCachePolicy

```go
func (c *BiCache) SetCachePolicy(policy CachePolicyFunc)
```

#### func (*BiCache) SetCapacity

```go
func (c *BiCache) SetCapacity(capacity int)
```

#### func (*BiCache) SetCompression

```go
func (c *BiCache) SetCompression(compression CompressionFunc, decompression DecompressionFunc)
```

#### func (*BiCache) SetDeserializer

```go
func (c *BiCache) SetDeserializer(deserializer *gob.Decoder)
```

#### func (*BiCache) SetGlobalExpiration

```go
func (c *BiCache) SetGlobalExpiration(expiration time.Duration)
```

#### func (*BiCache) SetSerializer

```go
func (c *BiCache) SetSerializer(serializer *gob.Encoder)
```

#### func (*BiCache) SetUpdateStrategy

```go
func (c *BiCache) SetUpdateStrategy(strategy UpdateStrategyFunc)
```

#### type CacheEntry

```go
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
	Accessed   time.Time
}
```


#### type CacheEvent

```go
type CacheEvent int
```


```go
const (
	CacheEventSet CacheEvent = iota
	CacheEventDelete
)
```

#### type CacheEventHandlerFunc

```go
type CacheEventHandlerFunc func(event CacheEvent, key interface{}, entry CacheEntry)
```


#### type CacheMetrics

```go
type CacheMetrics struct {
	Hits         int64
	Misses       int64
	SetSuccess   int64
	SetError     int64
	EntriesCount int64
}
```


#### type CachePolicyFunc

```go
type CachePolicyFunc func(key interface{}, entry CacheEntry) bool
```


#### type CompressionFunc

```go
type CompressionFunc func(data []byte) ([]byte, error)
```


#### type DecompressionFunc

```go
type DecompressionFunc func(data []byte) ([]byte, error)
```


#### type UpdateStrategyFunc

```go
type UpdateStrategyFunc func(key interface{}, oldValue interface{}) interface{}
```
