package slx

import (
	"go.starlark.net/starlark"
	"reflect"
	"fmt"
	"runtime"
	"strings"
)

func bindGoFunc(name string, funcVar interface{}) (goFunc *starlark.Builtin, err error) {
	if funcVar == nil {
		err = fmt.Errorf("funcVar must be a non-nil value")
		return
	}
	t := reflect.TypeOf(funcVar)
	if t.Kind() != reflect.Func {
		err = fmt.Errorf("funcVar expected to be a func")
		return
	}

	if len(name) == 0 {
		n := runtime.FuncForPC(reflect.ValueOf(funcVar).Pointer()).Name()
		if pos := strings.LastIndex(n, "."); pos >= 0 {
			name = n[pos+1:]
		} else {
			name = n
		}

		if len(name) == 0 {
			name = "noname"
		}
	}

	goFunc = starlark.NewBuiltin(name, wrapGoFunc(reflect.ValueOf(funcVar), t))
	return
}

func wrapGoFunc(fnVal reflect.Value, fnType reflect.Type) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (val starlark.Value, err error) {
		// check args number
		argsNum := args.Len()
		variadic := fnType.IsVariadic()
		lastNumIn := fnType.NumIn() - 1
		if variadic {
			if argsNum < lastNumIn {
				err = fmt.Errorf("at least %d args to call %s", lastNumIn, b.Name())
				return
			}
		} else {
			if argsNum != fnType.NumIn() {
				err = fmt.Errorf("%d args expected to call %s", argsNum, b.Name())
				return
			}
		}

		// make golang func args
		goArgs := make([]reflect.Value, argsNum)
		var fnArgType reflect.Type
		for i:=0; i<argsNum; i++ {
			if i<lastNumIn || !variadic {
				fnArgType = fnType.In(i)
			} else {
				fnArgType = fnType.In(lastNumIn).Elem()
			}

			goArgs[i] = makeValue(fnArgType)
			setValue(goArgs[i], fromValue(args.Index(i)))
		}

		// call golang func
		res := fnVal.Call(goArgs)

		// convert result to starlark
		retc := len(res)
		if retc == 0 {
			val = starlark.None
			return
		}
		lastRetType := fnType.Out(retc-1)
		if lastRetType.Name() == "error" {
			e := res[retc-1].Interface()
			if e != nil {
				err = e.(error)
				return
			}
			retc -= 1
			if retc == 0 {
				val = starlark.None
				return
			}
		}

		if retc == 1 {
			val = toValue(res[0].Interface())
			return
		}
		retV := make([]starlark.Value, retc)
		for i:=0; i<retc; i++ {
			retV[i] = toValue(res[i].Interface())
		}
		val = starlark.Tuple(retV)
		return
	}
}
