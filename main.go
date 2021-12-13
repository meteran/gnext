package main

import (
	"github.com/gin-gonic/gin"
	"gnext.io/gnext"
	"log"
	"net/http"
)

type Response struct {
	gnext.Response
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type Request struct {
	gnext.Body
	Name string `json:"name"`
}

type Query struct {
	gnext.Query
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
	Order  string `form:"order"`
}

type ErrorResult struct {
	Message string `json:"message"`
}

func (e *ErrorResult) StatusCodes() []int {
	return []int{409, 422}
}

func someHandler(param1 int, param2 string, body *Request, query *Query, headers *gnext.Headers, ctx *gin.Context) (*Response, error) {
	log.Println(param1, param2, body, query, headers, ctx.Request.Method)
	return &Response{
		Id:   123,
		Name: "hello world",
	}, nil
}

func main() {
	router := gnext.New()

	router.Use(NewMiddleware(MiddlewareOptions{
		startValue: 10,
	}))

	router.GET("/asd/:id/:id2/asd", someHandler)
	router.POST("/asd/:id/:id2/asd", someHandler)
	//router.POST("/asd/", someHandler)
	//
	srv := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router.Engine(),
	}

	log.Println("starting server")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("listen: %s\n", err)
	}
}
