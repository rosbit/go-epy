package epy

import (
	"go.starlark.net/starlark"
	"fmt"
	"reflect"
)

type userList struct {
	v reflect.Value
	iterCount  int
	frozen bool
}

func (l *userList) String() string {
	return fmt.Sprint(l.v.Interface())
}

func (l *userList) Type() string {
	return "user_list"
}

func (l *userList) Freeze() {
	l.frozen = true
}

func (l *userList) Truth() starlark.Bool {
	return l.v.Len() > 0
}

func (l *userList) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable %s", l.Type())
}

func (l *userList) Clear() (err error) {
	if err = l.canModify(); err != nil {
		return
	}
	l.v = l.v.Slice(0, 0)
	return
}

func (l *userList) Index(i int) starlark.Value {
	return toValue(l.v.Index(i).Interface())
}

func (l *userList) SetIndex(i int, v starlark.Value) (err error) {
	if err = l.canModify(); err != nil {
		return
	}
	return setValue(l.v.Index(i), fromValue(v))
}

func (l *userList) Slice(start, end, step int) starlark.Value {
	if step == 1 {
		newL := reflect.MakeSlice(l.v.Type(), end-start, end-start)
		reflect.Copy(newL, l.v.Slice(start, end))
		return &userList{v: newL}
	}
	newL := reflect.MakeSlice(l.v.Type().Elem(), 0, 0)
	direction := func(step int) int {
		switch {
		case step == 0:
			return 0
		case step < 0:
			return -1
		default:
			return 1
		}
	}
	d := direction(step)
	for i := start; direction(end-i) == d; i += step {
		newL = reflect.Append(newL, l.v.Index(i))
	}
	return &userList{v: newL}
}

func (l *userList) Len() int {
	return l.v.Len()
}

/*
func (l *userList) append(v interface{}) (err error) {
	if err = l.canModify(); err != nil {
		return
	}
	val := makeValue(l.v.Type().Elem())
	if err = setValue(val, v); err != nil {
		return
	}
	l.v = reflect.Append(l.v, val)
	return
}

func (l *userList) Attr(name string) (val starlark.Value, err error) {
	switch name {
	default:
		val = starlark.None
	case "append":
		val, err = bindGoFunc(name, l.append)
	}

	return
}

func (l *userList) AttrNames() []string {
	return []string{"append"}
}*/

func (l *userList) Iterate() starlark.Iterator {
	l.iterCount++
	return &listIter{l: l}
}

type listIter struct {
	l *userList
	i int
}

func (it *listIter) Next(v *starlark.Value) bool {
	if it.i < it.l.v.Len() {
		*v = toValue(it.l.v.Index(it.i).Interface())
		it.i++
		return true
	}
	return false
}

func (l *userList) canModify() error {
	if l.frozen {
		return fmt.Errorf("frozen userList")
	}
	if l.iterCount > 0 {
		return fmt.Errorf("iterated userList")
	}
	return nil
}

func (it *listIter) Done() {
	it.l.iterCount--
}

