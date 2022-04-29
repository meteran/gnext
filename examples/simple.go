package main

import (
	"github.com/meteran/gnext"
	"net/http"
)

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

type MyHeaders struct {
	gnext.Headers
	ContentType string `header:"Content-Type,default=application/json"`
}

type ShopQuery struct {
	gnext.Query
	Search string `form:"search"`
}

func handler(req *MyRequest) *MyResponse {
	return &MyResponse{Result: req.Name}
}

func getShop(paramName string, q *ShopQuery) *MyResponse {
	return &MyResponse{Result: paramName}
}

func getShopsList(q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status) {
	return &MyResponse{Result: h.ContentType}, http.StatusOK
}
