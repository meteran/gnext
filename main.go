package main

import (
	"github.com/gin-gonic/gin"
	"gnext.io/gnext"
	gdocs "gnext.io/gnext/docs"
	"log"
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

func someHandler(param1 int, param2 string, body *Request, query *Query, headers *gnext.Headers, ctx *gin.Context, context *SomeMiddleware) (gnext.Status, *Response) {
	log.Println(param1, param2, body, query, headers, ctx.Request.Method, context)
	return 201, &Response{
		Id:   123,
		Name: "hello world",
	}
}

func main() {
	router := gnext.New(
		&gdocs.Docs{
			OpenAPIPath:    "/docs",
			OpenAPIUrl:     "http://localhost:8000/docs/openapi.json",
			Title:          "gNext",
			Description:    "",
			TermsOfService: "http://localhost/terms",
			License:        nil,
			Contact:        nil,
			Version:        "1.0.0",
		},
	)

	router.Use(NewMiddleware(MiddlewareOptions{
		startValue: 10,
	}))

	router.GET("/asd/:id/:id2/asd", someHandler, &gdocs.PathDoc{
		Summary: "test",
	})
	router.POST("/asd/:id/:id2/asd", someHandler, nil)

	err := router.Run("0.0.0.0:8000")
	if err != nil {
		panic(err)
	}
}
