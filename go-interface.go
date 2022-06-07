package epy

import (
	elutils "github.com/rosbit/go-embedding-utils"
	"go.starlark.net/starlark"
	"fmt"
	"reflect"
)

type userInterface struct {
	v reflect.Value
}

func (i *userInterface) Attr(name string) (v starlark.Value, err error) {
	if len(name) == 0 {
		v = starlark.None
		err = fmt.Errorf("name expected")
		return
	}
	name = upperFirst(name)
	mV := i.v.MethodByName(name)
	if mV.Kind() != reflect.Invalid {
		mT := mV.Type()
		return starlark.NewBuiltin(name, wrapGoFunc(elutils.NewGolangFuncHelperDiretly(mV, mT))), nil
	}
	return starlark.None, nil
}

func (i *userInterface) AttrNames() []string {
	count := i.v.NumMethod()
	names := make([]string, count)
	t := i.v.Type()
	for j := 0; j < i.v.NumMethod(); j++ {
		names[j] = lowerFirst(t.Method(j).Name)
	}
	return names
}

func (i *userInterface) String() string {
	return fmt.Sprint(i.v.Interface())
}

func (i *userInterface) Type() string {
	return "user_interface"
}

func (i *userInterface) Freeze() {}

func (i *userInterface) Truth() starlark.Bool {
	return true
}

func (i *userInterface) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable %s", i.Type())
}
