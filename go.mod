module ginext

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	gnext.io/gnext v1.0.0
)

replace gnext.io/gnext => ./lib/gnext
