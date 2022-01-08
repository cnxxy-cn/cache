package cache

import (
	"time"
)

const (
	// NoExpiration For use with functions that take an expiration time.
	NoExpiration time.Duration = -1

	// DefaultExpiration For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

type CleanCallback func(interface{}, interface{})

type item struct {
	t int64
	v interface{}
}

func (i *item) check() bool {
	return i.t > 0 && time.Now().UnixNano() > i.t
}
