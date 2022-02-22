package main

import (
	"fmt"
	"github.com/meteran/gnext"
	gdocs "github.com/meteran/gnext/docs"
	"log"
)

type Response struct {
	gnext.Response `default_status:"200" status_codes:"201,202"`
	Id             int           `json:"id"`
	Name           string        `json:"name"`
	Test           interface{}   `json:"test"`
	TestAny        []interface{} `json:"test_any"`
}

type Request struct {
	gnext.Body                // optional if the handler has just one unknown type and method is POST/PUT/PATCH
	Name       string         `json:"name"`
	LastName   string         `json:"last_name" binding:"required"`
	Age        int            `json:"age"`
	Things     map[string]int `json:"things" binding:"required"`
}

type Query struct {
	gnext.Query        // optional if the handler has just one unknown type and method is GET/HEAD/DELETE/OPTIONS
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
	Order       string `form:"order"`
}

type Headers struct {
	gnext.Headers
	Test    string `header:"test"`
	ReqTest string `header:"req_test" binding:"required,max=255"`
}

func someHandler(param1 int, param2 string, query *Query, context *SomeContext, headers *Headers) (gnext.Status, *Response) { // NOTE: context comes from middleware
	fmt.Printf("%d, %s, %v, %v, %v", param1, param2, query, context, headers)
	return 200, &Response{
		Id:   123,
		Name: "hello world",
	}
}

func innerHandler(request *Request, context *SomeContext, context2 *SomeContext2) *Response { // NOTE: both contexts come from middlewares
	fmt.Printf("%v, %v, %v", request, context, context2)
	return &Response{
		Id:   0,
		Name: "123",
	}
}

type ErrorResponse struct {
	gnext.ErrorResponse `status_codes:"400,401,403,409,422"`
	Message             string `json:"message"`
	Success             bool   `json:"success"`
}

func dummyErrorHandler(err error) (gnext.Status, *ErrorResponse) {
	log.Printf("[222432755] err: %v", err)
	switch e := err.(type) {
	case *gnext.HandlerPanicked:
		return 500, &ErrorResponse{Message: fmt.Sprintf("services panicked with %v", e.Value)}
	}
	return 200, &ErrorResponse{Message: err.Error(), Success: true}
}

func main() {
	router := gnext.Router(&gdocs.Options{
		Title:       "My Project",
		Description: "An Awesome project",
		Servers:     []string{"http://example.com"},
	})

	router.Use(NewMiddleware(MiddlewareOptions{
		startValue: 10,
	}))

	router.GET("/asd/:id/:id2/asd", someHandler, &gdocs.Endpoint{Summary: "test"})
	router.POST("/header/:id/:id2/", someHandler)
	group := router.Group("/prefix")
	group.Use(NewMiddleware2(MiddlewareOptions{
		startValue: 0,
	}))
	group.OnError(dummyErrorHandler)
	group.POST("/path", innerHandler)

	//Example swagger servers
	//router.Docs.AddServer("http://localhost:8080/")
	//router.Docs.AddServer("https://api.test.com/v1")

	err := router.Run("0.0.0.0", "8080")
	if err != nil {
		panic(err)
	}
}
