# gNext Web Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/meteran/gnext)](https://goreportcard.com/report/github.com/meteran/gnext)
[![GoDoc](https://pkg.go.dev/badge/github.com/meteran/gnext?status.svg)](https://pkg.go.dev/github.com/meteran/gnext?tab=doc)
[![Release](https://img.shields.io/github/release/meteran/gnext.svg?style=flat-square)](https://github.com/meteran/gnext/releases)


gNext is a brilliant Golang API-focused framework extending [Gin](https://github.com/gin-gonic/gin). 
Offers the API structuring, automates validation and generates documentation. 
It's compatible with the existing Gin handlers and Gin middlewares.
Designed to simplify and boost development of JSON APIs. 
You can leave generic and boring stuff to gNext and purely focus on the business logic of your code.

## Contents

- [gNext Web Framework](#gnext-web-framework)
    - [Contents](#contents)
    - [Installation](#installation)
    - [Quick start](#quick-start)


## Installation

You can download gNext and install it in your project by running:

```sh
$ go get -u github.com/meteran/gnext
```


## Quick start

This tutorial assumes, that you already have Golang installation and basic knowledge about how to build and run Go programs.
If this is your first hit with Go, and you feel you have no idea what is happening here, please read how to [get started with Go](https://go.dev/doc/tutorial/getting-started).

Ok, so let's create a project:

```sh
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

```sh
go run example
```

Now you can visit this link in your browser: http://localhost:8080/example

Yes, yes... of course it works, but that's boring... Let's open this page: http://localhost:8080/docs

Whoa, that was amazing, and it's just the beginning.



