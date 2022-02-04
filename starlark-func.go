package epy

import (
	"go.starlark.net/starlark"
	"reflect"
)

func (slw *XStarlark) bindFunc(fn *starlark.Function, funcVarPtr interface{}) {
	dest := reflect.ValueOf(funcVarPtr).Elem()
	fnType := dest.Type()
	dest.Set(reflect.MakeFunc(fnType, slw.wrapFunc(fn, fnType)))
}

func (slw *XStarlark) wrapFunc(fn *starlark.Function, fnType reflect.Type) func(args []reflect.Value) (results []reflect.Value) {
	return func(args []reflect.Value) (results []reflect.Value) {
		// make starlark args
		var slArgs []starlark.Value
		lastNumIn := fnType.NumIn() - 1
		variadic := fnType.IsVariadic()
		for i, arg := range args {
			if i < lastNumIn || !variadic {
				slArgs = append(slArgs, toValue(arg.Interface()))
				continue
			}

			if arg.IsZero() {
				break
			}
			varLen := arg.Len()
			for j:=0; j<varLen; j++ {
				slArgs = append(slArgs, toValue(arg.Index(j).Interface()))
			}
		}

		// call starlark function
		res, err := starlark.Call(slw.thread, fn, starlark.Tuple(slArgs), nil)

		// convert result to golang
		results = make([]reflect.Value, fnType.NumOut())
		if err == nil {
			if fnType.NumOut() > 0 {
				if res.Type() == "tuple" {
					mRes := fromValue(res).([]interface{})
					l := len(mRes)
					n := fnType.NumOut()
					if n < l {
						l = n
					}
					for i:=0; i<l; i++ {
						v := reflect.New(fnType.Out(i)).Elem()
						rv := mRes[i]
						if err = setValue(v, rv); err == nil {
							results[i] = v
						}
					}
				} else {
					v := reflect.New(fnType.Out(0)).Elem()
					rv := fromValue(res)
					if err = setValue(v, rv); err == nil {
						results[0] = v
					}
				}
			}
		}

		if err != nil {
			nOut := fnType.NumOut()
			if nOut > 0 && fnType.Out(nOut-1).Name() == "error" {
				results[nOut-1] = reflect.ValueOf(err).Convert(fnType.Out(nOut-1))
			} else {
				panic(err)
			}
		}

		for i, v := range results {
			if !v.IsValid() {
				results[i] = reflect.Zero(fnType.Out(i))
			}
		}

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
