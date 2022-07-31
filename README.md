# gNext Web Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/meteran/gnext)](https://goreportcard.com/report/github.com/meteran/gnext)
[![GoDoc](https://pkg.go.dev/badge/github.com/meteran/gnext?status.svg)](https://pkg.go.dev/github.com/meteran/gnext?tab=doc)
[![Release](https://img.shields.io/github/release/meteran/gnext.svg?style=flat-square)](https://github.com/meteran/gnext/releases)

gNext is a brilliant Golang API-focused framework extending [Gin](https://github.com/gin-gonic/gin). Offers the API
structuring, automates validation and generates documentation. It's compatible with the existing Gin handlers and Gin
middlewares. Designed to simplify and boost development of JSON APIs. You can leave generic and boring stuff to gNext
and purely focus on the business logic of your code.

## Contents

- [gNext Web Framework](#gnext-web-framework)
    - [Contents](#contents)
    - [Installation](#installation)
    - [Quick start](#quick-start)
    - [Url parameters](#url-parameters)
    - [Query parameters](#query-parameters)
    - [Request payload](#request-payload)
    - [Response payload](#response-payload)
    - [Status codes](#status-codes)
    - [Request headers](#request-headers)
    - [Response headers](#response-headers)
    - [Request context](#request-context)
    - [Endpoint groups](#endpoint-groups)
    - [Middleware](#middleware)
    - [Error handler](#error-handler)
    - [Router options](#router-options)
    - [Benchmarks](#benchmarks)
    - [Advanced documentation](#advanced-documentation)
    - [Authors](#authors)

## Installation

You can download gNext and install it in your project by running:

```shell
$ go get -u github.com/meteran/gnext
```

## Quick start

This tutorial assumes, that you already have Golang installation and basic knowledge about how to build and run Go
programs. If this is your first hit with Go, and you feel you have no idea what is happening here, please read how
to [get started with Go](https://go.dev/doc/tutorial/getting-started).

Ok, so let's create a project:

```shell
mkdir gnext-example
cd gnext-example
go mod init example.com/gnext
go get github.com/meteran/gnext
```

Create a file `example.go` and fill it up with the following code:

```go
package main

import "github.com/meteran/gnext"

func main() {
	r := gnext.Router()

	r.GET("/example", func() string {
		return "Hello World!"
	})

	_ = r.Run()
}
```

Run it:

```shell
go run example
```

Now you can visit this link in your browser: http://localhost:8080/example

Yes, yes... of course it works, but that's boring... Let's open this page: http://localhost:8080/docs

Whoa, that was amazing, ...but not very useful.

Let's try some real example. With request and response. We can modify our handler to use structures:

```go
package main

import "github.com/meteran/gnext"

func main() {
	r := gnext.Router()

	r.POST("/example", handler)
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
```

Restart the server and visit the docs page. You can see that request and response of `POST /example` endpoint are
documented. That's the real power!

The POST request without required `id` now fails with the validation error:

```shell
curl --request POST http://localhost:8080/example --data '{"name": "some name"}'
```

gives output:

```json
{
  "message": "validation error",
  "details": [
    "field validation for 'id' failed on the 'required' tag with value ''"
  ],
  "success": false
}
```

the valid request:

```shell
curl --request POST http://localhost:8080/example --data '{"name": "some name", "id": 4}'
```

gives us the expected response:

```json
{
  "result": "some name"
}
```

Congratulations! Now you are prepared for the fast forwarding development of your great API.

_Note:_ all following sections will base on the program from this overview.

## Url parameters

Okay, in the previous section we saw quick use, let's get to specific things 😎

First, we'll start with the parameters in the url.

Using them the standard way is unpleasant, we have to write a piece of code in the handler method just to be able to use
it knowing the type safely - not cool.

But... Gnext will do it for us 🥳!

Let's see, I'll add a new endpoint with parameter and add handler method to it:

```go
func main() {
    r := gnext.Router()

    r.POST("/example", handler)
    r.GET("/shops/:name/", getShop)
    _ = r.Run()
}

func getShop(paramName string) *MyResponse {
    return &MyResponse{Result: paramName}
}
```

Ok, now restart the server and use a new endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/myownshop/' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "myownshop"
}
```

Cool, yeah? Let's take a look at http://localhost:8080/docs, but ... don't be surprised when

you will see a documented endpoint ```/shops/{name}/```  ready to use straight from the Swagger interface 👏

_Note_:  adding new parameters as arguments to the handler methods, keep the order in accordance with the parameters in
the url.

## Query parameters

Okay, let's move on to a topic with a similar problem as in the previous section - Query parameters.

Exactly the same problem as with url parameters, to use them we have to add a piece of code, but why not use the magic
of gNext 🎩?

Let's add some query parameter to our new shop list endpoint```/shops/```:

```go
func main() {
    r := gnext.Router()

    r.POST("/example", handler)
    r.GET("/shops/", getShopsList)
    r.GET("/shops/:name/", getShop)
    _ = r.Run()
}

type ShopQuery struct {
  gnext.Query
  Search       string    `form:"search"`
}

func getShopsList(q *ShopQuery) *MyResponse {
    return &MyResponse{Result: q.Search}
}
```

Ok, now restart the server and use a new endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/?search=wantedshop' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "wantedshop"
}
```

As before, in the documentation we find a new endpoint ready to be used by the interface 👷‍♀️

## Request payload

## Response payload

## Status codes

In this section we will show you returning statuses with gNext 🙌.

It's simple, just add ```gnext.Status``` to the returned handler parameters

Example:

```go
func getShopsList(q *ShopQuery)(*MyResponse, gnext.Status) {
    return nil, http.StatusNotFound
}
```

Ok, now restart the server and use endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/?search=wantedshop' \
  -H 'accept: application/json'
```

And the response status we will be ```404``` 🦾

## Request headers

It happens that we want to do something with the request `Headers` - nothing difficult.

Look, I will add the headers structure and use it in the handler:

```go
type MyHeaders struct {
  gnext.Headers
  ContentType string `header:"Content-Type,default=application/json"`
}

func getShopsList(q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status){
    return &MyResponse{Result: h.ContentType}, http.StatusOK
}
```

Ok, now restart the server and use endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "application/json"
}
```

It's all simple isn't it? Of course you can enter headers in the Swagger interface 🫡

## Response headers

## Request context

It may happen that we need to get directly to the request context, just add ```*gin.Context``` to the arguments of the
handler method.

Let's look at an example:

```go
func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status){
    return &MyResponse{Result: c.Request.Method}, http.StatusOK
}
```

Ok, now restart the server and use endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "GET"
}
```

## Endpoint groups

As in the standard GinGonic and many frameworks, we support Endpoint Groups.

We use them not only for structuring, we can also add gnext middleware or gnext error handler to them.

Let's look at a simple example to create a group:

```go
func main() {
  r := gnext.Router()
  
  r.POST("/example", handler)
  r.Group("/shops").
    GET("/", getShopsList).
    GET("/:name/", getShop)
  _ = r.Run()
}
```

Okay, now we can restart the server using the previously created endpoints in exactly the same way.

_Note_: using middleware and error handler for the group will be presented in their individual documentation sections.

## Middleware

We decided to make the use of middleware a little easier, but also to develop it.

In gNext we offer the possibility to use the `Before` and `After` middleware.

We will prepare a sample middleware using `Before` to authorize the user and pass the user context to the handler method.

Let's start by creating the user context that we want to use in the handler method:

```go
type UserCtx struct {
	Id int `json:"id"`
}
```
Then let's take care of where we will get the data for the middleware, in this case it will be the header and the `Authorization` field:
```go
type AuthorizationHeaders struct {
	gnext.Headers
	Authorization string `header:"Authorization"`
}
```

Okay, it's time to create a middleware method, in this case it will be this:

```go
func NewAuthMiddleware() gnext.Middleware {
	return gnext.Middleware{
		Before: func(headers AuthorizationHeaders) (*UserCtx, error) {
			if headers.Authorization == "" {
				return nil, fmt.Errorf("authorization is required")
			}
			return &UserCtx{Id: 1}, nil
		},
	}
}
```

Now let's add to our group to use this middleware:

```go
r.Group("/shops").
  OnError(shopErrorHandler).
  Use(NewAuthMiddleware()).
  GET("/", getShopsList).
  GET("/:name/", getShop)
```

It's time to use our middleware in the handler method 😎

```go
func getShop(paramName string, q *ShopQuery, userCtx *UserCtx) *MyResponse {
	return &MyResponse{Result: fmt.Sprintf("user_id: %d", userCtx.Id)}
}
```

Ok now restart the server and use the new endpoint, let's start with the situation without authorization header given:
```shell
curl -X 'GET' \
  'http://localhost:8080/shops/someshop/' \
  -H 'accept: application/json'
```

the response will look like this with `200` http code:

```json
{
  "message": "authorization is required",
  "success": true
}
```

_Note_: to make it `401`, just add a new error and add it to the error handler 👏

Okay, now let's add the value in the Authorization header:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/asd/' \
  -H 'accept: application/json' \
  -H 'authorization: some_key'
```

the response will look like this with `200` http code:

```json
{"result":"user_id: 1"}
```

_Note_: detailed middleware description will be added in `Advanced documentation`.

## Error handler

Okay, since we already know that gNext can do a lot, maybe we'll take care of it and structure error handling?

Our gNext allows you to create your own error handlers, you can bind them to groups.

Error handler allows you to separate error handling from the handler method, after all, we don't need to repeat exactly the same lines of code in all methods.

Well, let's see how it looks in practice, at the beginning let's add the error structure we want to return:

```go
type ErrorResponse struct {
  gnext.ErrorResponse `status_codes:"400,401,403,409,422"`
  Message             string `json:"message"`
  Success             bool   `json:"success"`
}
```
As you probably noticed, we added a struct tag for `gnext.ErrorResponse`,

`status_codes` - codes from this tag will be added to the documentation.

Okay, now let's add our own error type:

```go
type InvalidSearchError struct{ error }
```

It's time to edit our handler method a bit, we'll add out the error parameter:

```go
func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, error) {
	if q.Search == "any"{
		return nil, &InvalidSearchError{}
	}
	return &MyResponse{Result: q.Search}, nil
}
```

Now, we will create out error handler:
```go
func shopErrorHandler(err error) (gnext.Status, *ErrorResponse) {
	switch e := err.(type) {
	case *gnext.HandlerPanicked:
		return 500, &ErrorResponse{Message: fmt.Sprintf("services panicked with %v", e.Value)}
	case *InvalidSearchError:
		return 422, &ErrorResponse{Message: fmt.Sprintf("invalid search value")}
	}
	return 200, &ErrorResponse{Message: err.Error(), Success: true}
}
```

_Note_: Error handler will not be used if error is `nil`.

Okay, there is the last straight, so let's add our error handler to the group:

```go
func simpleRouter() {
  r := gnext.Router()
  
  r.POST("/example", handler)
  r.Group("/shops").
    OnError(shopErrorHandler).
    GET("/", getShopsList).
    GET("/:name/", getShop)
    _ = r.Run()
}
```

Ok, now restart the server and use endpoint:

```shell
curl -X 'GET' \
  'http://localhost:8080/shops/?search=any' \
  -H 'accept: application/json'
```

the response will look like this with `422` http code:

```json
{
  "message": "invalid search value",
  "success": false
}
```

## Router options

## Benchmarks

## Advanced documentation

## Authors

