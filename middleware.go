package main

import (
	"gnext.io/gnext"
	"os/user"
	"time"
)

type MiddlewareOptions struct {
	startValue int
}

func NewMiddleware(options MiddlewareOptions) gnext.MiddlewareFactory {
	return func() gnext.Middleware {
		middleware := SomeMiddleware{
			count: options.startValue,
		}
		return gnext.Middleware{
			Before: middleware.Begin,
			After:  middleware.End,
		}
	}
}

type SomeMiddleware struct {
	count int
	start time.Time
}

func (m *SomeMiddleware) Begin(headers gnext.Headers) *user.User {
	m.count++
	m.start = time.Now()
	return nil
}

func (m *SomeMiddleware) End() {
	println(time.Now().Sub(m.start))
}
