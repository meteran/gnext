module gnext

go 1.16

require (
	github.com/alecthomas/jsonschema v0.0.0-20211022214203-8b29eab41725 // indirect
	github.com/gin-gonic/gin v1.7.7 // indirect
	gnext.io/gnext v1.0.0
)

replace (
	gnext.io/gnext => ./lib/gnext
)
