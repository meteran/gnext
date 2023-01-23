## Gin context

It may happen that we need to get directly to the gin request context. If so, just add ```*gin.Context``` to handler's arguments.

Let's have a look at an example:

```go
func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status){
    return &MyResponse{Result: c.Request.Method}, http.StatusOK
}
```

Ok, now restart the server and use endpoint:

```console
$ curl -X 'GET' \
  'http://localhost:8080/shops/' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "GET"
}
```
