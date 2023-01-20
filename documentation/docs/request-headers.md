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

```console
$ curl -X 'GET' \
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
