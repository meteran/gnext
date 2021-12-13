package main

import (
	"fmt"
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
		After: func(context *SomeMiddleware, resp *Response) {
			context.count++
			fmt.Println(resp)
			fmt.Printf("%s\n", time.Now().Sub(context.start))
		},
	}
}
