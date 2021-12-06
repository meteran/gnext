package main

import (
	"gnext.io/gnext"
	"time"
)

type MiddlewareOptions struct {
	startValue int
}

type SomeMiddleware struct {
	count int
	start time.Time
}

func NewMiddleware(options MiddlewareOptions) gnext.Middleware {
	return gnext.Middleware{
		Before: func(headers gnext.Headers) *SomeMiddleware {
			context := &SomeMiddleware{
				count: options.startValue,
			}
			context.start = time.Now()
			return context
		},
		After: func(context *SomeMiddleware) {
			context.count++
			println(time.Now().Sub(context.start))
		},
	}
}
