package main

import (
	"gnext.io/gnext"

	"log"
)

type Response struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type Request struct {
	gnext.MarkBody
	Name string `json:"name"`
}

type Query struct {
	gnext.MarkQuery
	Limit  int
	Offset int
	Order  string
}

type ErrorResult struct {
	Message string `json:"message"`
}

func (e *ErrorResult) StatusCodes() []int {
	return []int{409, 422}
}

type User struct {
}

func someHandler(param1 int, param2 string, body *Request, query *Query) (*Response, *ErrorResult) {
	log.Println(param1, param2, body, query)
	return nil, nil
}

func main() {
	router := gnext.New()

	q := &Query{}
	func(a interface{}) {
		b, ok := a.(gnext.QueryInterface)
		println(b, ok)
	}(q)

	router.GET("/asd/", someHandler)
	//router.POST("/asd/", someHandler)
	//
	//srv := &http.Server{
	//	Addr:    "0.0.0.0:8000",
	//	Handler: router.Engine(),
	//}
	//
	//log.Println("starting server")
	//if err := srv.ListenAndServe(); err != nil {
	//	log.Fatalf("listen: %s\n", err)
	//}
}
