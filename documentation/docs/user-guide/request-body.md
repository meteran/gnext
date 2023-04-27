# Request body

As you could see in the [First steps](../../first-steps/) section, you can define the request payload as your own structure.
It will be parsed then to fill the documentation. When request is made, its body is bound to the given structure using gin-binding module.
This implies body validation using the [validator package](https://github.com/go-playground/validator).
If there is a validation error, the error from gin's `Context.ShouldBindWith` method will be raised. Be sure, you have an [error handler](../error-handling/) listening on such errors.

Because the handler arguments can be mixed, gNext has to recognize which parameter should be bound to the body.
To mark structure as a request body, use a base struct inside:

```go
type MyRequest struct {
	gnext.Body
    Id   int    `json:"id" binding:"required"`
    Name string `json:"name"`
}
```
