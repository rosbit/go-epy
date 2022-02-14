package epy

import (
	"go.starlark.net/starlark"
	"reflect"
	"fmt"
)

type userMap struct {
	v reflect.Value
	iterCount  int
	frozen bool
}

func (m *userMap) canModify() (err error) {
	if m.frozen {
		err = fmt.Errorf("map is frozen")
		return
	}
	if m.iterCount > 0 {
		err = fmt.Errorf("map is iterated")
		return
	}
	return
}

func (m *userMap) SetKey(k, v starlark.Value) (err error) {
	if err = m.canModify(); err != nil {
		return
	}

	mT := m.v.Type()
	key := makeValue(mT.Key())
	if err = setValue(key, fromValue(k)); err != nil {
		return
	}

	val := makeValue(mT.Elem())
	if err = setValue(val, fromValue(v)); err != nil {
		return
	}
	m.v.SetMapIndex(key, val)
	return
}

func (m *userMap) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	key := makeValue(m.v.Type().Key())
	if err = setValue(key, fromValue(k)); err != nil {
		return
	}
	val := m.v.MapIndex(key)
	if val.Kind() == reflect.Invalid {
		v = starlark.None
		return
	}
	v, found = toValue(val.Interface()), true
	return
}

func (m *userMap) String() string {
	return fmt.Sprint(m.v.Interface())
}

func (m *userMap) Type() string {
	return "user_map"
}

func (m *userMap) Freeze() {
	m.frozen = true
}

func (m *userMap) Truth() starlark.Bool {
	return m.v.Len() > 0
}

func (m *userMap) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable %s", m.Type())
}

func (m *userMap) Len() int {
	return m.v.Len()
}

func (m *userMap) get(k interface{}, def ...interface{}) interface{} {
	key := reflect.ValueOf(k)
	val := m.v.MapIndex(key)
	if val.Kind() == reflect.Invalid {
		if len(def) > 0 {
			return def[0]
		}
		return nil
	}
	return val.Interface()
}

func (m *userMap) Attr(name string) (val starlark.Value, err error) {
	if name != "get" {
		return starlark.None, nil
	}
	val, err = bindGoFunc(name, m.get)
	return
}

func (m *userMap) AttrNames() []string {
	return []string{"get"}
}

func (m *userMap) Iterate() starlark.Iterator {
	m.iterCount++
	return &mapIter{
		m: m,
		i: m.v.MapRange(),
	}
}

type mapIter struct {
	m *userMap
	i *reflect.MapIter
}

func (it *mapIter) Next(k *starlark.Value) bool {
	if !it.i.Next() {
		return false
	}
	key := it.i.Key()
	*k = toValue(key.Interface())
	return true
}

func (it *mapIter) Done() {
	it.m.iterCount--
}

