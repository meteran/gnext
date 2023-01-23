## Endpoint groups

As in the standard GinGonic and many frameworks, we support Endpoint Groups.

We use them not only for structuring, we can also add gnext middleware or gnext error handler to them.

Let's have a look at a simple example to create a group:

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
