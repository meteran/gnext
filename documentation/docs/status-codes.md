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

```console
$ curl -X 'GET' \
  'http://localhost:8080/shops/?search=wantedshop' \
  -H 'accept: application/json'
```

And the response status we will be ```404``` 🦾
