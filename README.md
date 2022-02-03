# go-starlark-x

[Starlark in Go](https://github.com/google/starlark-go) is an interpreter for Starlark, a python-like language implemented in Go. 

This package is intended to extend the go-starlark to make `golang functions and struct instances`
as Starlark `built-in functions and modules` very easily, so calling golang functions or modules from
Starlark is very simple.

This package also provides a helper to bind a golang function with a Starlark function, so that calling Starlark function is simple, too.

### Usage

The package is fully go-getable, So, just type

  `go get github.com/rosbit/go-starlark-x`

to install.

```go
package main

import (
  "github.com/rosbit/go-starlark-x"
  "fmt"
)

func main() {
  ctx := slx.NewStarlark()

  res, _ := ctx.Eval("1 + 2", nil)
  fmt.Println("result is:", res)
}
```

### Go calls Starlark function

Suppose there's a Starlark file named `a.star` like this:

```python
def slAdd(a, b):
    return a+b
```

one can call the Starlark function slAdd() in Go code like the following:

```go
package main

import (
  "github.com/rosbit/go-starlark-x"
  "fmt"
)

var slAdd func(int, int)int

func main() {
  ctx := slx.NewStarlark()
  if err := ctx.LoadFile("a.star", nil); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  if err := ctx.BindFunc("slAdd", &slAdd); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  res := slAdd(1, 2)
  fmt.Println("result is:", res)
}
```

### Starlark calls Go function

Starlark calling Go function is also easy. In the Go code, make a golang function
as Starlark built-in func by calling `MakeBuiltinFunc("funcname", function)`. There's the example:

```go
package main

import "github.com/rosbit/go-starlark-x"

// function to be called by Starlark
func adder(a1 float64, a2 float64) float64 {
    return a1 + a2
}

func main() {
  ctx := slx.NewStarlark()

  ctx.MakeBuiltinFunc("adder", adder)
  ctx.EvalFile("b.star", nil)  // b.star containing code calling "adder"
}
```

In Starlark code, one can call the registered name directly. There's the example `b.star`.

```python
r = adder(1, 100)   # the function "adder" is implemented in Go
print(r)
```

### Make Go struct as Starlark module

This package provides a function `SetModule` which will convert a Go struct into
a Starlark module. There's the example `c.star`:

```python
m.IncAge(10)
print(m)

print('m.name', m.name)
print('m.age', m.age)
```

The Go code is like this:

```go
package main

import "github.com/rosbit/go-starlark-x"

type M struct {
   Name string
   Age int
}
func (m *M) IncAge(a int) {
   m.Age += a
}

func main() {
  ctx := js.NewStarlark()
  ctx.SetModule("m", &M{Name:"rosbit", Age: 1})

  ctx.EvalFile("c.star", nil)
}
```

### Set built-in functions and modules at one time

```go
package main

import "github.com/rosbit/go-starlark-x"
import "fmt"

type M struct {
   Name string
   Age int
}
func (m *M) IncAge(a int) {
   m.Age += a
}

func adder(a1 float64, a2 float64) float64 {
    return a1 + a2
}

func main() {
  vars := map[string]interface{}{
     "m": &M{Name:"rosbit", Age:1}, // to Starlark module
     "adder": adder,                // to Starlark built-in function
     "a": []int{1,2,3}              // to Starlark array
  }

  ctx := js.NewStarlark()
  if err := ctx.LoadFile("file.star", vars); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  res, err := ctx.GetGlobals("a") // get the value of var named "a"
  if err != nil {
     fmt.Printf("%v\n", err)
     return
  }
  fmt.Printf("res:", res)
}
```

### Wrap Go functions as Starlark module

This package provides a function `CreateModule` which will create a Starlark module containing
go functions as module methods. There's the example `d.star`:

```python
a = tm.newA("rosbit", 10)
a.IncAge(10)
print(a)

tm.printf('m.name: %s\n', a.name)
tm.printf('m.age: %d\n', a.age)
```

The Go code is like this:

```go
package main

import (
  "github.com/rosbit/go-starlark-x"
  "fmt"
)

type A struct {
   Name string
   Age int
}
func (m *A) IncAge(a int) {
   m.Age += a
}
func newA(name string, age int) *A {
   return &A{Name: name, Age: age}
}

func main() {
  ctx := js.NewStarlark()
  ctx.CreateModule("tm", map[string]interface{}{
     "newA": newA,
     "printf": fmt.Printf,
  })

  ctx.EvalFile("d.star", nil)
}
```

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
