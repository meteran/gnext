# Error handling

## In general
Whenever a middleware or handler returns a value which implements `error` interface, the execution flow is frozen and appropriate error handler is called.
To create an error handler, you need to implement a function with the following  rules:

1. Takes exactly one argument, which type implements `error` interface.
2. Can not return an `error`.
3. Values returned from error handler behave similarly like from a middleware.
4. Should return a response, because the handler might not be called.

The simplest example can look like:

```go
func errorHandler(err error) (string, gnext.Status) {
	return err.Error(), 500
}
```

Register it in the router:

```go
r.OnError(errorHandler)
```

Now every returned error will be handled by our custom error handler.

## Default handler
If you don't register any error handler, the default one will be called, which gracefully handles the following errors:

* `*json.SyntaxError`
* `*json.UnmarshalTypeError`
* `validator.ValidationErrors`
* `*gnext.NotFound`

You probably noticed, that in documentation of your API, next to your response, there are error responses. 
They come from the default error handler definition which returns the following response struct:

```go
type DefaultErrorResponse struct {
	ErrorResponse `default_status:"500" status_codes:"4XX,5XX"`
	Message       string   `json:"message"`
	Details       []string `json:"details"`
	Success       bool     `json:"success"`
}
```

There are 3 status codes defined in response above. 
That's why in your documentation there are `500`, `4XX` and `5XX` status codes with the response scheme above.


## Error response

In order to document error response scheme, you need to define a response structure.
It will be similar to the one in [default response](#default-handler).
Example:

```go
type MyResponse struct {
	Message       string   `json:"message"`
	Success       bool     `json:"success"`
}
```

## HTTP status

You can define HTTP status returned together with the error response just inside a struct.
To do that, you need to mark your struct as an error response and add `default_status` tag.

```go
type MyResponse struct {
	gnext.ErrorResponse `default_status:"422"`
	Message             `json:"message"`
}
```

Now, you can simplify your error handler and don't return a status:

```go
func errorHandler(err error) *MyResponse {
    return &MyResponse{Message: err.Error()}
}
```

gNext will document `MyResponse` scheme under `422` HTTP status code. It will also use it in response as a default status code.
You can override it, by returning `gnext.Status` from handler:

```go
func errorHandler(err error) (*MyResponse, gnext.Status) {
	response := &MyResponse{Message: err.Error()}
	if response.Message == "not found" {
		return response, 404
    }
    return response, 400
}
```

This will return `404` HTTP status code if the error message is `not found`. However, `404` code won't be documented.
To add additional status codes to documentation, you need to write comma-separated statuses to `status_codes` tag of `gnext.ErrorResponse`:

```go
type MyResponse struct {
	gnext.ErrorResponse `default_status:"422" status_codes:"400,404"`
	Message             `json:"message"`
}
```

Now you will see the error response with all three codes.

!!! warning "Warning"
    OpenAPI v3 scheme keeps responses in a map: status code -> response schema. 
    This means, it is not possible to document more than one response scheme for one status code.
    If you register more than one response with the same status code, the random one will be exposed in documentation.

## Unauthorized
Let's extend our example from [authorization middleware](../middlewares/#authorization) and add proper error handling:

```go title="example.go"
package main

import (
	"github.com/meteran/gnext"
	"log"
	"net/http"
)

func handler(userId int, actor *User) (*User, gnext.Status) {
	log.Printf("actor: %v", actor)
	for _, usr := range users {
		if usr.Id == userId {
			return usr, http.StatusOK
		}
	}
	return nil, http.StatusNotFound
}

func main() {
	r := gnext.Router()

	r.Use(gnext.Middleware{
		Before: authorizationMiddleware,
	})

	r.GET("/users/:id/", handler)
	_ = r.Run("", "8080")
}
```

```go title="users.go"
package main

import "github.com/meteran/gnext"

type User struct {
	gnext.Response
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Token string `json:"-"`
}

var users = []*User{
	{
		Id:    0,
		Name:  "Krzesimir",
		Age:   34,
		Token: "token-1",
	},
}
```

This time just one user is enough. We will focus on errors. 
As described in the [authorization middleware](../middlewares/#authorization) page, if we don't provide a valid token, we get a 500 HTTP status code.
This not an intended behavior, since if the user is not authorized we should return 401 status code.

