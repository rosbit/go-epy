package slx

import (
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlark"
	"fmt"
	"reflect"
)

func wrapModule(modName string, name2FuncVarPtr map[string]interface{}) (mod *starlarkstruct.Module, err error) {
	methods := make(starlark.StringDict)
	for n, fn := range name2FuncVarPtr {
		if len(n) == 0 {
			err = fmt.Errorf("blank method name found")
			return
		}
		fnV := reflect.ValueOf(fn)
		if fnV.Kind() != reflect.Func {
			err = fmt.Errorf("func expected for method %s", n)
			return
		}
		fnT := fnV.Type()
		methods[n] = starlark.NewBuiltin(n, wrapGoFunc(fnV, fnT))
	}

	mod = &starlarkstruct.Module{
		Name: modName,
		Members: methods,
	}
	return
}
