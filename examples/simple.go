package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext"
)

func simpleRouter() {
	r := gnext.Router()
	r.Use(NewAuthMiddleware())

	r.POST("/example", handler)
	r.Group("/shops").
		OnError(shopErrorHandler).
		Use(NewAuthMiddleware()).
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

func getShop(paramName string, q *ShopQuery, userCtx *UserCtx) *MyResponse {
	return &MyResponse{Result: fmt.Sprintf("user_id: %d", userCtx.Id)}
}

func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, error) {
	if q.Search == "any" {
		return nil, &InvalidSearchError{}
	}
	return &MyResponse{Result: q.Search}, nil
}

type ErrorResponse struct {
	gnext.ErrorResponse `status_codes:"400,401,403,409,422"`
	Message             string `json:"message"`
	Success             bool   `json:"success"`
}

type InvalidSearchError struct{ error }

func shopErrorHandler(err error) (gnext.Status, *ErrorResponse) {
	switch e := err.(type) {
	case *gnext.HandlerPanicked:
		return 500, &ErrorResponse{Message: fmt.Sprintf("services panicked with %v", e.Value)}
	case *InvalidSearchError:
		return 422, &ErrorResponse{Message: "invalid search value"}
	}
	return 200, &ErrorResponse{Message: err.Error(), Success: true}
}
