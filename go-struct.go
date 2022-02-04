package slx

import (
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

type userModule struct {
	*starlarkstruct.Module
	originStruct reflect.Value
}
func (m *userModule) Type() string {
	return "user_module"
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
		Module: &starlarkstruct.Module{
			Name: name,
			Members: wrapGoStruct(structVar, structE, structT),
		},
		originStruct: structVar,
	}
	return
}

func wrapGoStruct(structVar, structE reflect.Value, structT reflect.Type) starlark.StringDict {
	r := make(starlark.StringDict)
	for i:=0; i<structT.NumField(); i++ {
		strField := structT.Field(i)
		name := strField.Name
		name = strings.ToLower(name[:1]) + name[1:]
		fv := structE.Field(i)
		r[name] = toValue(fv.Interface())
	}

	// receiver is the struct
	bindGoMethod(structE, structT, r)

	// reciver is the pointer of struct
	t := structVar.Type()
	bindGoMethod(structVar, t, r)
	return r
}

func bindGoMethod(structV reflect.Value, structT reflect.Type, r starlark.StringDict) {
	for i := 0; i<structV.NumMethod(); i+=1 {
		m := structT.Method(i)
		name := strings.ToLower(m.Name[:1]) + m.Name[1:]
		mV := structV.Method(i)
		mT := mV.Type()
		r[name] = starlark.NewBuiltin(name, wrapGoFunc(mV, mT))
	}
}
