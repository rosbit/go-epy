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

func (m *userMap) Clear() (err error) {
	if err = m.canModify(); err != nil {
		return
	}
	for _, k := range m.v.MapKeys() {
		m.v.SetMapIndex(k, reflect.Value{})
	}
	return
}

func (m *userMap) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if err = m.canModify(); err != nil {
		return
	}
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
	m.v.SetMapIndex(key, reflect.Value{})
	return
}

func (m *userMap) Items() []starlark.Tuple {
	count := m.v.Len()
	tuples := make([]starlark.Tuple, count)
	i := 0
	it := m.v.MapRange()
	for it.Next() {
		tuple := make(starlark.Tuple, 2)
		tuple[0] = toValue(it.Key().Interface())
		tuple[1] = toValue(it.Value().Interface())
		tuples[i] = tuple
		i += 1
	}
	return tuples
}

func (m *userMap) Keys() []starlark.Value {
	count := m.v.Len()
	keys := make([]starlark.Value, count)
	i := 0
	it := m.v.MapRange()
	for it.Next() {
		keys[i] = toValue(it.Key().Interface())
		i += 1
	}
	return keys
}

func (m *userMap) Len() int {
	return m.v.Len()
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

