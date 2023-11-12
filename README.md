# BiCache

**BiCache** is a simple, flexible, and scalable in-memory caching library. BiCache can store any data type as key-value pairs and can be customized with various features.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Acknowledgement](#acknowledgement)
- [Sponsors](#sponsors)
- [License](#license)

## Features

- **Capacity Control:** BiCache performs automatic cleanup operations when the maximum capacity is reached.
- **Timed Caching:** Support for timed caching where expiration time can be set individually for each item.
- **Cache Policies:** Ability to integrate user-defined custom cache policies.
- **Global Timed Cache:** Setting a global timed cache for all cache items.
- **Event Handler:** Ability to add a custom event handler to track cache events.
- **Update Strategies:** Ability to integrate user-defined strategies for updating items added to the cache.
- **Compression/Decompression:** Ability to integrate user-defined functions for data compression and decompression.

## Installation

```bash
go get -u github.com/mtnmunuklu/bicache
```

## Usage
```go
package main

import (
	"fmt"
	"time"

	"github.com/your-username/your-package/bicache"
)

func main() {
	// Create a BiCache instance
	cache := bicache.NewBiCache(100, time.Minute)

	// Add an item to the cache
	cache.Set("key", "value", time.Second*30)

	// Retrieve an item from the cache
	value, exists := cache.Get("key")
	if exists {
		fmt.Println("Value:", value)
	} else {
		fmt.Println("Item not found.")
	}

	// Cache cleanup operation
	cache.Delete("key")
}
```

## Documentation

For more examples and library usage, please refer to the [Documentation](docs/bicache.md).

## Contributing

Contributions to BiCache are welcome and encouraged! Please read the [contribution guidelines](CONTRIBUTING.md) before making any contributions to the project.

## Acknowledgments

A big shoutout to [Oğuzhan Mevsim](https://github.com/ogzhnmvsm) for discovering the project name for BiCache.

## Sponsors

If you are interested in becoming a sponsor, please visit our [GitHub Sponsors](https://github.com/sponsors/mtnmunuklu) page.

## License

This project is licensed under the [MIT License](LICENSE).
