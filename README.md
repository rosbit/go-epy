# go-epy, an embedding Python

[Starlark in Go](https://github.com/google/starlark-go), a Python-like script language, is an interpreter for Starlark implemented in pure Go. 

`go-epy` is a package extending the starlark-go and making it a **pragmatic embedding** language.
With some helper functions provided by `go-epy`, calling Golang functions or modules from Starlark, 
or calling Starlark functions from Golang are both very simple. So, with the help of `go-epy`, starlark-go
can be looked as **an embedding Python**.

### Usage

The package is fully go-getable, so, just type

  `go get github.com/rosbit/go-epy`

to install.

#### 1. Evaluate expressions

```go
package main

import (
  "github.com/rosbit/go-epy"
  "fmt"
)

func main() {
  ctx := epy.New()

  res, _ := ctx.Eval("1 + 2", nil)
  fmt.Println("result is:", res)
}
```

#### 2. Go calls Starlark function

Suppose there's a Starlark file named `a.py` like this:

```python
def add(a, b):
    return a+b
```

one can call the Starlark function `add()` in Go code like the following:

```go
package main

import (
  "github.com/rosbit/go-epy"
  "fmt"
)

var add func(int, int)int

func main() {
  ctx := epy.New()
  if err := ctx.LoadFile("a.py", nil); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  if err := ctx.BindFunc("add", &add); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  res := add(1, 2)
  fmt.Println("result is:", res)
}
```

#### 3. Starlark calls Go function

Starlark calling Go function is also easy. In the Go code, make a Golang function
as Starlark built-in function by calling `MakeBuiltinFunc("funcname", function)`. There's the example:

```go
package main

import "github.com/rosbit/go-epy"

// function to be called by Starlark
func adder(a1 float64, a2 float64) float64 {
    return a1 + a2
}

func main() {
  ctx := epy.New()

  ctx.MakeBuiltinFunc("adder", adder)
  ctx.EvalFile("b.py", nil)  // b.py containing code calling "adder"
}
```

In Starlark code, one can call the registered function directly. There's the example `b.py`.

```python
r = adder(1, 100)   # the function "adder" is implemented in Go
print(r)
```

#### 4. Make Go struct instance as a Starlark module

This package provides a function `SetModule` which will convert a Go struct instance into
a Starlark module. There's the example `c.py`, `m` is the module provided by Go code:

```python
m.incAge(10)
print(m)

print('m.name', m.name)
print('m.age', m.age)
```

The Go code is like this:

```go
package main

import "github.com/rosbit/go-epy"

type M struct {
   Name string
   Age int
}
func (m *M) IncAge(a int) {
   m.Age += a
}

func main() {
  ctx := epy.New()
  ctx.SetModule("m", &M{Name:"rosbit", Age: 1}) // "m" is the module name

  ctx.EvalFile("c.py", nil)
}
```

#### 5. Set many built-in functions and modules at one time

If there're a lot of functions and modules to be registered, a map could be constructed and put as an
argument for functions `LoadFile`, `LoadScript`, `EvalFile` or `Eval`.

```go
package main

import "github.com/rosbit/go-epy"
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

  ctx := epy.New()
  if err := ctx.LoadFile("file.py", vars); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  res, err := ctx.GetGlobal("a") // get the value of var named "a". Any variables in script could be get by GetGlobal
  if err != nil {
     fmt.Printf("%v\n", err)
     return
  }
  fmt.Printf("res:", res)
}
```

#### 6. Wrap Go functions as Starlark module

This package also provides a function `CreateModule` which will create a Starlark module integrating any
Go functions as module methods. There's the example `d.py` which will use module `tm` provided by Go code:

```python
a = tm.newA("rosbit", 10)
a.incAge(10)
print(a)

tm.printf('a.name: %s\n', a.name)
tm.printf('a.age: %d\n', a.age)
```

The Go code is like this:

```go
package main

import (
  "github.com/rosbit/go-epy"
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
  ctx := epy.New()
  ctx.CreateModule("tm", map[string]interface{}{ // module name is "tm"
     "newA": newA,            // make user defined function as module method named "tm.newA"
     "printf": fmt.Printf,    // make function in a standard package named "tm.printf"
  })

  ctx.EvalFile("d.py", nil)
}
```

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
