package slx

import (
	"go.starlark.net/starlark"
)

type XStarlark struct {
	globals starlark.StringDict
	thread *starlark.Thread
}

