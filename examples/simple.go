package main

import "github.com/meteran/gnext"

func simpleRouter() {
	r := gnext.Router()

	r.POST("/example", handler)
	r.GET("/shops/", getShopsList)
	r.GET("/shops/:name/", getShop)
	_ = r.Run()
}

type MyRequest struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name"`
}

type MyResponse struct {
	Result string `json:"result"`
}

type ShopQuery struct {
	gnext.Query
	Search       string    `form:"search"`
}

func handler(req *MyRequest) *MyResponse {
	return &MyResponse{Result: req.Name}
}

func getShop(paramName string, q *ShopQuery) *MyResponse {
	return &MyResponse{Result: q.Search}
}

func getShopsList(q *ShopQuery) *MyResponse {
	return &MyResponse{Result: q.Search}
}
