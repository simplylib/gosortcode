# gosortcode
gosortcode is a Go (golang) program to sort Go source code in a opinionated way

## known limitations

### undefined behavior
behavior is undefined if source input is not valid and compiling go code

### var declarations
var declarations are not currently sorted lexicographically unlike const declarations

```go
type Name int

const (
	Jill Name = iota
	John
	Caddy
)
``` 

is sorted

this is not:

```go
type Name int

var (
	Jill  Name = 0
	John  Name = 1
	Caddy Name = 2
)
```
