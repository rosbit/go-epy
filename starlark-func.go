package epy

import (
	elutils "github.com/rosbit/go-embedding-utils"
	"go.starlark.net/starlark"
	"reflect"
)

func (slw *XStarlark) bindFunc(fn *starlark.Function, funcVarPtr interface{}) (err error) {
	helper, e := elutils.NewEmbeddingFuncHelper(funcVarPtr)
	if e != nil {
		err = e
		return
	}
	helper.BindEmbeddingFunc(slw.wrapFunc(fn, helper))
	return
}

func (slw *XStarlark) wrapFunc(fn *starlark.Function, helper *elutils.EmbeddingFuncHelper) elutils.FnGoFunc {
	return func(args []reflect.Value) (results []reflect.Value) {
		var slArgs []starlark.Value

		// make starlark args
		itArgs := helper.MakeGoFuncArgs(args)
		for arg := range itArgs {
			slArgs = append(slArgs, toValue(arg))
		}

		// call starlark function
		res, err := starlark.Call(slw.thread, fn, starlark.Tuple(slArgs), nil)
		// convert result to golang
		results = helper.ToGolangResults(fromValue(res), res.Type() == "tuple", err)
		return
	}
}

func (slw *XStarlark) callFunc(fn *starlark.Function, args ...interface{}) (res starlark.Value, err error) {
	slArgs := make([]starlark.Value, len(args))
	for i, arg := range args {
		slArgs[i] = toValue(arg)
	}

	return starlark.Call(slw.thread, fn, starlark.Tuple(slArgs), nil)
}
