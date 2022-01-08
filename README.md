# go cache

Implemented in two ways

The first one, sync.map is used as storage to ensure the safety of goroutines without using locks. 
This implementation is suitable for scenarios with more reads and fewer writes. 
Both key and val are interface{}.

Second, map[string]interface{} is used as storage, and read-write locks are used to ensure goroutines safety.
This implementation is suitable for the scenario of average reading and writing or more writing and less reading. 
It supports key as string and val as interface{} type.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines.

### Installation

`go get github.com/cnxxy-cn/cache`

### Usage

```go
import (
	"fmt"
	"github.com/cnxxy-cn/cache"
	"time"
)

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	c := cache.NewS(5*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}

}
```

### Reference

`godoc` or [http://godoc.org/github.com/cnxxy-cn/cache](http://godoc.org/github.com/cnxxy-cn/cache)
