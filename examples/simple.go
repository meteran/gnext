package main

import "github.com/meteran/gnext"

func simpleRouter() {
	r := gnext.Router()

	r.GET("/example", func() gnext.Status {
		return 204
	})

	_ = r.Run()
}
