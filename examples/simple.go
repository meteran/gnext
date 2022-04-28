package main

import "github.com/meteran/gnext"

func simpleRouter() {
	r := gnext.Router()

	r.POST("/example", handler)
	r.GET("/shop/:name/", getShop)
	_ = r.Run()
}

type MyRequest struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name"`
}

type MyResponse struct {
	Result string `json:"result"`
}

func handler(req *MyRequest) *MyResponse {
	return &MyResponse{Result: req.Name}
}

func getShop(paramName string) *MyResponse {
	return &MyResponse{Result: paramName}
}
