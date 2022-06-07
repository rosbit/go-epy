package epy

import (
	elutils "github.com/rosbit/go-embedding-utils"
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
	key := elutils.MakeValue(mT.Key())
	if err = elutils.SetValue(key, fromValue(k)); err != nil {
		return
	}

	val := elutils.MakeValue(mT.Elem())
	if err = elutils.SetValue(val, fromValue(v)); err != nil {
		return
	}
	m.v.SetMapIndex(key, val)
	return
}

func (m *userMap) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	key := elutils.MakeValue(m.v.Type().Key())
	if err = elutils.SetValue(key, fromValue(k)); err != nil {
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

func (m *userMap) keys() []interface{} {
	l := m.v.Len()
	if l == 0 {
		return nil
	}
	res := make([]interface{}, l)
	i := 0
	it := m.v.MapRange()
	for it.Next() {
		k := it.Key()
		res[i] = k.Interface()
		i += 1
	}
	return res
}

func (m *userMap) items() ([]starlark.Tuple) {
	l := m.v.Len()
	if l == 0 {
		return nil
	}
	res := make([]starlark.Tuple, l)
	i := 0
	it := m.v.MapRange()
	for it.Next() {
		t := make(starlark.Tuple, 2)
		k := it.Key()
		v := it.Value()
		t[0] = toValue(k.Interface())
		t[1] = toValue(v.Interface())
		res[i] = t
		i += 1
	}
	return res
}

func (m *userMap) Attr(name string) (val starlark.Value, err error) {
	switch name {
	default:
		if m.v.Len() == 0 {
			val = starlark.None
			break
		}
		kt := m.v.Type().Key()
		if kt.Kind() != reflect.String {
			val = starlark.None
			break
		}
		key := reflect.ValueOf(name)
		v := m.v.MapIndex(key)
		if v.Kind() == reflect.Invalid {
			val = starlark.None
		} else {
			val = toValue(v.Interface())
		}
	case "get":
		val, err = bindGoFunc(name, m.get)
	case "items":
		val, err = bindGoFunc(name, m.items)
	case "keys":
		val, err = bindGoFunc(name, m.keys)
	}
	return
}

func (m *userMap) AttrNames() []string {
	return []string{"get", "items", "keys"}
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

