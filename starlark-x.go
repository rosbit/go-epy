package slx

import (
	"go.starlark.net/starlark"
	"go.starlark.net/lib/json"
	"go.starlark.net/lib/math"
	"go.starlark.net/lib/time"
	"fmt"
	"reflect"
)

func init() {
	starlark.Universe["json"] = json.Module
	starlark.Universe["time"] = time.Module
	starlark.Universe["math"] = math.Module
}

func NewStarlark() *XStarlark {
	return &XStarlark{
		thread: &starlark.Thread{Name:"starlark-x"},
	}
}

func (slw *XStarlark) LoadFile(path string, vars map[string]interface{}) (err error) {
	globals, err := starlark.ExecFile(slw.thread, path, nil, convertEnv(vars))
	if err != nil {
			return err
	}
	slw.globals = globals
	return nil
}

func (slw *XStarlark) LoadScript(script string, vars map[string]interface{}) (err error) {
	globals, err := starlark.ExecFile(slw.thread, "load-script.star", script, convertEnv(vars))
	if err != nil {
			return err
	}
	slw.globals = globals
	return nil
}

func (slw *XStarlark) GetGlobal(name string) (res interface{}, err error) {
	r, e := slw.getVar(name)
	if e != nil {
		err = e
		return
	}
	res = fromValue(r)
	return
}

func (slw *XStarlark) EvalFile(path string, env map[string]interface{}) (res interface{}, err error) {
	v, e := starlark.Eval(slw.thread, path, nil, convertEnv(env))
	if e != nil  {
		err = e
		return
	}
	res = fromValue(v)
	return
}

func (slw *XStarlark) Eval(script string, env map[string]interface{}) (res interface{}, err error) {
	v, e := starlark.Eval(slw.thread, "eval-script", script, convertEnv(env))
	if e != nil  {
		err = e
		return
	}
	res = fromValue(v)
	return
}

func (slw *XStarlark) CallFunc(funcName string, args ...interface{}) (res interface{}, err error) {
	v, e := slw.getVar(funcName)
	if e != nil {
		err = e
		return
	}
	fn, ok := v.(*starlark.Function)
	if !ok {
		err = fmt.Errorf("var %s is not with type function", funcName)
		return
	}

	r, e := slw.callFunc(fn, args...)
	if e != nil {
		err = e
		return
	}
	res = fromValue(r)
	return
}

// bind a var of golang func with a Starlark function name, so calling Starlark function
// is just calling the related golang func.
// @param funcVarPtr  in format `var funcVar func(....) ...; funcVarPtr = &funcVar`
func (slw *XStarlark) BindFunc(funcName string, funcVarPtr interface{}) (err error) {
	if funcVarPtr == nil {
		err = fmt.Errorf("funcVarPtr must be a non-nil poiter of func")
		return
	}
	t := reflect.TypeOf(funcVarPtr)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Func {
		err = fmt.Errorf("funcVarPtr expected to be a pointer of func")
		return
	}

	v, e := slw.getVar(funcName)
	if e != nil {
		err = e
		return
	}
	fn, ok := v.(*starlark.Function)
	if !ok {
		err = fmt.Errorf("var %s is not with type function", funcName)
		return
	}
	slw.bindFunc(fn, funcVarPtr)
	return
}

// make a golang func as a built-in Starlark function, so the function can be called in Starlark script.
func (slw *XStarlark) MakeBuiltinFunc(funcName string, funcVar interface{}) (err error) {
	goFunc, e := bindGoFunc(funcName, funcVar)
	if e != nil {
		err = e
		return
	}
	starlark.Universe[funcName] = goFunc
	return
}

// make a golang pointer of sturct instance as a Starlark module.
// @param structVarPtr  pointer of struct instance is recommended.
func (slw *XStarlark) SetModule(modName string, structVarPtr interface{}) (err error) {
	if structVarPtr == nil {
		err = fmt.Errorf("structVarPtr must ba non-nil pointer of struct")
		return
	}
	v := reflect.ValueOf(structVarPtr)
	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		starlark.Universe[modName] = bindGoStruct(modName, v)
		return
	}
	err = fmt.Errorf("structVarPtr must be struct or pointer of strcut")
	return
}

// wrapper some `name2FuncVarPtr` to a module named `modName`
// @param name2FuncVarPtr must be string => func
func (slw *XStarlark) CreateModule(modName string, name2FuncVarPtr map[string]interface{}) (err error) {
	if len(modName) == 0 {
		err = fmt.Errorf("modName expected")
		return
	}
	if len(name2FuncVarPtr) == 0 {
		err = fmt.Errorf("name2FuncVarPtr expected")
		return
	}
	mod, e := wrapModule(modName, name2FuncVarPtr)
	if e != nil {
		err = e
		return
	}
	starlark.Universe[modName] = mod
	return
}

func convertEnv(vars map[string]interface{}) (starlark.StringDict) {
	if len(vars) == 0 {
		return nil
	}
	res := make(starlark.StringDict)
	for k, v := range vars {
		res[k] = toValue(v)
	}
	return res
}

func (slw *XStarlark) getVar(name string) (v starlark.Value, err error) {
	if len(slw.globals) == 0 {
		err = fmt.Errorf("no var named %s found", name)
		return
	}
	r, ok := slw.globals[name]
	if !ok {
		err = fmt.Errorf("no var named %s found", name)
		return
	}
	v = r
	return
}
