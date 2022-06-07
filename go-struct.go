package epy

import (
	elutils "github.com/rosbit/go-embedding-utils"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
	"fmt"
)

type userModule struct {
	modName string
	structVar reflect.Value
	structE   reflect.Value
	structT   reflect.Type
	attrNames []string
}

func (m *userModule) Type() string {
	return "user_module"
}

func (m *userModule) Attr(name string) (starlark.Value, error) {
	if len(name) == 0 {
		return starlark.None, nil
	}
	name = upperFirst(name)
	mV := m.structVar.MethodByName(name)
	if mV.Kind() != reflect.Invalid {
		mT := mV.Type()
		return starlark.NewBuiltin(name, wrapGoFunc(elutils.NewGolangFuncHelperDiretly(mV, mT))), nil
	}
	mV = m.structE.MethodByName(name)
	if mV.Kind() != reflect.Invalid {
		mT := mV.Type()
		return starlark.NewBuiltin(name, wrapGoFunc(elutils.NewGolangFuncHelperDiretly(mV, mT))), nil
	}
	if _, ok := m.structT.FieldByName(name); !ok {
		return starlark.None, nil
	}
	fV := m.structE.FieldByName(name)
	return toValue(fV.Interface()), nil
}

func (m *userModule) AttrNames() []string {
	return m.attrNames
}

func (m *userModule) SetField(name string, val starlark.Value) error {
	if len(name) == 0 {
		return fmt.Errorf("field name expected")
	}
	name = upperFirst(name)
	if _, ok := m.structT.FieldByName(name); !ok {
		return fmt.Errorf("field %s not found", name)
	}
	fV := m.structE.FieldByName(name)
	return elutils.SetValue(fV, fromValue(val))
}

func (m *userModule) Freeze() {}

func (m *userModule) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", m.Type())
}

func (m *userModule) String() string {
	return fmt.Sprintf("<user_module %q>", m.modName)
}

func (m *userModule) Truth() starlark.Bool {
	return true
}

func bindGoStruct(name string, structVar reflect.Value) (goModule *userModule) {
	var structE reflect.Value
	if structVar.Kind() == reflect.Ptr {
		structE = structVar.Elem()
	} else {
		structE = structVar
	}
	structT := structE.Type()

	if structE == structVar {
		// struct is unaddressable, so make a copy of struct to an Elem of struct-pointer.
		// NOTE: changes of the copied struct cannot effect the original one. it is recommended to use the pointer of struct.
		structVar = reflect.New(structT) // make a struct pointer
		structVar.Elem().Set(structE)    // copy the old struct
		structE = structVar.Elem()       // structE is the copied struct
	}

	if len(name) == 0 {
		n := structT.Name()
		if len(n) > 0 {
			if pos := strings.LastIndex(n, "."); pos >= 0 {
				name = n[pos+1:]
			} else {
				name = n
			}
		}
		if len(name) == 0 {
			name = "noname"
		}
	}

	goModule = &userModule{
		modName: name,
		structVar: structVar,
		structE: structE,
		structT: structT,
		attrNames: getAttrNames(structVar, structE, structT),
	}
	return
}

func lowerFirst(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}
func upperFirst(name string) string {
	return strings.ToUpper(name[:1]) + name[1:]
}

func getAttrNames(structVar, structE reflect.Value, structT reflect.Type) []string {
	count := structT.NumField() + structVar.NumMethod() + structE.NumMethod()
	names := make([]string, count)
	i := 0
	for j:=0; j<structT.NumField(); j,i = j+1,i+1 {
		name := lowerFirst(structT.Field(j).Name)
		names[i] = name
	}
	for j:=0; j<structE.NumMethod(); j,i = j+1,i+1 {
		name := lowerFirst(structT.Method(j).Name)
		names[i] = name
	}
	t := structVar.Type()
	for j:=0; j<structVar.NumMethod(); j,i = j+1,i+1 {
		name := lowerFirst(t.Method(j).Name)
		names[i] = name
	}
	return names
}

