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
As you could see, the query struct is marked as query using the base struct `gnext.Query`.
Some structs have to be marked in order to inform gNext what part of request it represents.
If there is just one data structure of unknown request part, it is considered depending on the HTTP method:

* `GET`, `DELETE`, `HEAD`, `OPTIONS` - unknown argument is considered as query parameters
* `POST`, `PUT`, `PATCH` - unknown argument is considered as request body

So, if you want to catch the query params in `GET` request, and all other parameters don't exist, or are marked as headers, url params, you don't need to mark the structure as `gnext.Query`.

If there is more than one unknown structure, the gNext will panic during handler creation.

Ok, now restart the server and use a new endpoint:

```console
$ curl -X 'GET' \
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

It is important, that in one request, you can have only one query parameters structure. Using the following handler will panic:
```go
type ShopQuery struct {
	gnext.Query
	Search       string    `form:"search"`
}

type ShopSecondQuery struct {
	gnext.Query
	Order       string    `form:"order"`
}

func getShopsList(q1 *ShopQuery, q2 *ShopSecondQuery) *MyResponse {
	...
}
```
