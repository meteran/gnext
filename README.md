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
    - [Advanced documentation](#advanced-documentation)

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

Restart the server and visit the docs page. You can see that request and response of `POST /example` endpoint are documented. That's the real power!

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

## Query parameters

## Request payload

## Response payload

## Status codes

## Request headers

## Response headers

## Request context

## Endpoint groups

## Middleware

## Error handler

## Router options

## Advanced documentation

