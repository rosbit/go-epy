package epy

import (
	elutils "github.com/rosbit/go-embedding-utils"
	"go.starlark.net/starlark"
)

func bindGoFunc(name string, funcVar interface{}) (goFunc *starlark.Builtin, err error) {
	helper, e := elutils.NewGolangFuncHelper(name, funcVar)
	if e != nil {
		err = e
		return
	}

	goFunc = starlark.NewBuiltin(helper.GetRealName(), wrapGoFunc(helper))
	return
}

func wrapGoFunc(helper *elutils.GolangFuncHelper) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (val starlark.Value, err error) {
		getArgs := func(i int) interface{} {
			return fromValue(args.Index(i))
		}

		v, e := helper.CallGolangFunc(args.Len(), b.Name(), getArgs)
		if e != nil {
			err = e
			return
		}
		if v == nil {
			val = starlark.None
			return
		}

		if vv, ok := v.([]interface{}); ok {
			retV := make([]starlark.Value, len(vv))
			for i, rv := range vv {
				retV[i] = toValue(rv)
			}
			val = starlark.Tuple(retV)
		} else {
			val = toValue(v)
		}
		return
	}
}
