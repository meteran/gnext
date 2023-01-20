## Url parameters

Okay, in the previous section we saw quick use, let's get to specific things 😎

First, we'll start with the parameters in the url.

Using them the standard way is unpleasant, we have to write a piece of code in the handler method just to be able to use
it knowing the type safely - not cool.

But... Gnext will do it for us 🥳!

Let's see, I'll add a new endpoint with parameter and add handler method to it:

```go
func main() {
    r := gnext.Router()

    r.POST("/example", handler)
    r.GET("/shops/:name/", getShop)
    _ = r.Run()
}

func getShop(paramName string) *MyResponse {
    return &MyResponse{Result: paramName}
}
```

Ok, now restart the server and use a new endpoint:

```console
$ curl -X 'GET' \
  'http://localhost:8080/shops/myownshop/' \
  -H 'accept: application/json'
```

the response will look like this:

```json
{
  "result": "myownshop"
}
```

Cool, yeah? Let's take a look at http://localhost:8080/docs, but ... don't be surprised when

you will see a documented endpoint ```/shops/{name}/```  ready to use straight from the Swagger interface 👏

_Note_:  adding new parameters as arguments to the handler methods, keep the order in accordance with the parameters in
the url.
