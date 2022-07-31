package main

import (
	"fmt"
	"github.com/meteran/gnext"
	"time"
)

type MiddlewareOptions struct {
	startValue int
}

type SomeContext struct {
	count int
	start time.Time
}

func NewMiddleware(options MiddlewareOptions) gnext.Middleware {
	return gnext.Middleware{
		Before: func(headers gnext.Headers) *SomeContext {
			context := &SomeContext{
				count: options.startValue,
			}
			context.start = time.Now()
			return context
		},
		After: func(context *SomeContext, resp *Response, status gnext.Status) {
			context.count++
			fmt.Printf("%s\n", time.Now().Sub(context.start))
		},
	}
}

type SomeContext2 struct {
	count int
	start time.Time
}

func NewMiddleware2(options MiddlewareOptions) gnext.Middleware {
	return gnext.Middleware{
		Before: func(headers gnext.Headers) *SomeContext2 {
			context := &SomeContext2{
				count: options.startValue,
			}
			context.start = time.Now()
			return context
		},
		After: func(context *SomeContext2, resp *Response, status gnext.Status) {
			context.count++
			fmt.Printf("%s\n", time.Now().Sub(context.start))
		},
	}
}

type UserCtx struct {
	Id int `json:"id"`
}

type AuthorizationHeaders struct {
	gnext.Headers
	Authorization string `header:"Authorization"`
}

func NewAuthMiddleware() gnext.Middleware {
	return gnext.Middleware{
		Before: func(headers AuthorizationHeaders) (*UserCtx, error) {
			if headers.Authorization == "" {
				return nil, fmt.Errorf("authorization is required")
			}
			return &UserCtx{Id: 1}, nil
		},
	}
}
