package main

import (
	"fmt"
	"gnext.io/gnext"
	"net/http"
	"reflect"

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

type ErrorResult struct {
	Message string `json:"message"`
}

func (e *ErrorResult) StatusCodes() []int {
	return []int{409, 422}
}

func someHandler(param1 int, param2 string, body *Request, query *Query, headers gnext.Headers) (*Response, *ErrorResult) {
	log.Println(param1, param2, body, query, headers)
	return &Response{
		Id:   123,
		Name: "hello world",
	}, nil
}

type Error struct {
	error
}

func f(b string) Error {
	fmt.Printf("\n%v\n", b)
	return Error{}
}

func main() {
	var a *string
	r := reflect.ValueOf(a)
	fmt.Printf("%v", r)
	fmt.Printf("%v", r.Elem())
	fmt.Printf("\n\n%v\n\n", reflect.TypeOf(f).Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()))

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
