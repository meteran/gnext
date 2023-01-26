# Gin context

It is possible to use the raw gin context. Just add to your middleware/handler an argument of type `*gin.Context`:

```go
func getShopsList(c *gin.Context, q *ShopQuery, h *MyHeaders) (*MyResponse, gnext.Status){
    return &MyResponse{Result: c.Request.Method}, http.StatusOK
}
```

