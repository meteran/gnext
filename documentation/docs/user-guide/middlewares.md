# Middlewares

## Authorization
Almost every application needs some authorization and authentication mechanism. 
To be compliant with DRY and to avoid boilerplate code, most frameworks use middleware for such stuff.
gNext also allows you to write middlewares in.

Let's take a simple user model in `users.go` file:

```go
// file: users.go

type User struct {
    Id    int    `json:"id"`
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Token string `json:"-"`
}
```

And a list of user:

```go
// file: users.go

var users = []*User{
    {
        Id:    0,
        Name:  "Krzesimir",
        Age:   34,
        Token: "token-1",
    },
    {
        Id:    1,
        Name:  "Ziemowit",
        Age:   42,
        Token: "token-2",
    },
}
```

!!! warning "Security alert"
    For simplicity, we use a plain text tokens in this example. Normally user authorization should be secured using e.g. [JWT](https://github.com/golang-jwt/jwt).

Let's say, that everyone who is authorized can see information about any other user (file: `example.go`).

```go
// file: example.go

func handler(userId int) (*User, gnext.Status) {
	for _, usr := range users {
		if usr.Id == userId {
			return usr, http.StatusOK
		}
	}
	return nil, http.StatusNotFound
}

func main() {
	r := gnext.Router()

	r.GET("/users/:id/", handler)
	_ = r.Run("", "8080")
}
```

After running it, you can try it out directly from the docs. But now, everyone can see any user without authorization.
Let's define a middleware then.

Middleware is a function which looks very similarly to handler. 
It can be executed before or after handler and have input and output parameters. 
As input parameters you can use the same request data as in handler. Output parameters can be used in next middlewares and handler.

So let's create an authorization middleware, which will check a token provided in `Authorization` header. 
If token doesn't exist nor belong to some user, we want to stop the flow, not running handler.  

```go
// file: middleware.go

import (
	"fmt"
	"github.com/meteran/gnext"
)

type AuthorizationHeaders struct {
	gnext.Headers
	Authorization string `header:"Authorization"`
}

func AuthorizationMiddleware(headers *AuthorizationHeaders) (*User, error) {
	for _, usr := range users {
		if usr.Token == headers.Authorization {
			return usr, nil
		}
	}
	return nil, fmt.Errorf("unauthorized")
}
```

And use this middleware before we set up a handler:

```go
// file: example.go

func main() {
	r := gnext.Router()

	r.Use(gnext.Middleware{
		Before: AuthorizationMiddleware,
	})

	r.GET("/users/:id/", handler)
	_ = r.Run("", "8080")
}
```

Now only authorized users can access the data. The rest will get an unhandled error - 500.
Any values with custom types returned from the middleware can be passed to a next middleware or a handler as an input parameter.
Knowing that, we can modify our handler to take a current user (recognized by a token) as a handler parameter:

```go
// file: example.go

func handler(userId int, actor *User) (*User, gnext.Status) {
    log.Printf("actor: %v", actor)
    for _, usr := range users {
        if usr.Id == userId {
            return usr, http.StatusOK
        }
    }
    return nil, http.StatusNotFound
}
```

gNext has mapped the user object by type `*User`. 
It recognized, that a middleware has output parameter `*User` and handler has an input parameter of the same type, and it called handler with `*User` value.

