package main

import (
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext"
	"net/http"
)

func simpleRouter() {
	r := gnext.Router()

	r.POST("/example", handler)
	r.Group("/shops").
		GET("/", getShopsList).
		GET("/:name/", getShop)
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

func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status) {
	return &MyResponse{Result: c.Request.Method}, http.StatusOK
}
