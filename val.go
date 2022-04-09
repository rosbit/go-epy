package epy

import (
	sltime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"reflect"
	"fmt"
	"time"
	"strings"
	"math/big"
)

func toValue(v interface{}) starlark.Value {
	if v == nil {
		return starlark.None
	}

	switch vv := v.(type) {
	case int,int8,int16,int32,int64:
		return starlark.MakeInt64(reflect.ValueOf(v).Int())
	case uint,uint8,uint16,uint32,uint64:
		return starlark.MakeUint64(reflect.ValueOf(v).Uint())
	case *big.Int:
		return starlark.MakeBigInt(vv)
	case float32,float64:
		return starlark.Float(reflect.ValueOf(v).Float())
	case string:
		return starlark.String(vv)
	case []byte:
		return starlark.Bytes(vv)
	case bool:
		return starlark.Bool(vv)
	case time.Time:
		return sltime.Time(vv)
	case time.Duration:
		return sltime.Duration(vv)
	case starlark.Value:
		return vv
	default:
		v2 := reflect.ValueOf(v)
		switch v2.Kind() {
		case reflect.Slice, reflect.Array:
			return &userList{v: v2}
		case reflect.Map:
			return &userMap{v: v2}
		case reflect.Struct:
			return bindGoStruct("", v2)
		case reflect.Ptr:
			e := v2.Elem()
			if e.Kind() == reflect.Struct {
				return bindGoStruct("", v2)
			}
			return toValue(e.Interface())
		case reflect.Func:
			if f, err := bindGoFunc("", v); err == nil {
				return f
			}
			return starlark.None
		case reflect.Interface:
			return &userInterface{v: v2}
		default:
			return starlark.None
		}
	}
}

func fromValue(v starlark.Value) (interface{}) {
	if v == nil {
		return nil
	}
	switch v.Type() {
	case "NoneType":
		return nil
	case "bool":
		return bool(v.(starlark.Bool))
	case "bytes":
		return []byte(string(v.(starlark.Bytes)))
	case "int":
		i := v.(starlark.Int)
		if i64, ok := i.Int64(); ok {
			return i64
		}
		if u64, ok := i.Uint64(); ok {
			return u64
		}
		return i.BigInt()
	case "float":
		return float64(v.(starlark.Float))
	case "string":
		return string(v.(starlark.String))
	case "list":
		list := v.(*starlark.List)
		l := list.Len()
		res := make([]interface{}, l)
		for i:=0; i<l; i++ {
			res[i] = fromValue(list.Index(i))
		}
		return res
	case "tuple":
		t := v.(starlark.Tuple)
		l := t.Len()
		res := make([]interface{}, l)
		for i:=0; i<l; i++ {
			res[i] = fromValue(t.Index(i))
		}
		return res
	case "dict":
		d := v.(*starlark.Dict)
		res := make(map[interface{}]interface{})
		ks := d.Keys()
		for _, k := range ks {
			vv, _, _ := d.Get(k)
			res[fromValue(k)] = fromValue(vv)
		}
		return res
	case "set":
		s := v.(*starlark.Set)
		l := s.Len()
		res := make([]interface{}, l)
		iter := s.Iterate()
		defer iter.Done()
		var vv starlark.Value
		i := 0
		for iter.Next(&vv) {
			res[i] = fromValue(vv)
		}
		return res
	case "function":
		f := v.(*starlark.Function)
		return f
	case "builtin_function_or_method":
		f := v.(*starlark.Builtin)
		return f
	case "time.time":
		return time.Time(v.(sltime.Time))
	case "time.duration":
		return time.Duration(v.(sltime.Duration))
	case "user_module":
		return v.(*userModule).structVar.Interface()
	case "user_map":
		return v.(*userMap).v.Interface()
	case "user_list":
		return v.(*userList).v.Interface()
	case "user_interface":
		return v.(*userInterface).v.Interface()
	default:
		return nil
	}
}

func setValue(dest reflect.Value, val interface{}) error {
	dt := dest.Type()
	if val == nil {
		if dest.CanAddr() {
			dest.Set(reflect.Zero(dt))
		}
		return nil
	}
	v := reflect.ValueOf(val)
	vt := reflect.TypeOf(val)
	if vt.AssignableTo(dt) {
		dest.Set(v)
		return nil
	}

	if vt.ConvertibleTo(dt) {
		dest.Set(v.Convert(dt))
		return nil
	}

	switch v.Kind() {
	case reflect.Map:
		switch dest.Kind() {
		case reflect.Struct:
			return map2Struct(dest, v)
		case reflect.Ptr:
			if dest.Elem().Kind() == reflect.Struct {
				return map2Struct(dest.Elem(), v)
			}
		case reflect.Map:
			return map2Map(dest, v)
		default:
		}
	case reflect.Slice:
		if dest.Kind() == reflect.Slice {
			return copySlice(dest, v)
		}
	}

	return fmt.Errorf("cannot convert %s to %s", vt, dt)
}

func map2Struct(dest reflect.Value, v reflect.Value) error {
	dt := dest.Type()
	for i:=0; i<dt.NumField(); i++ {
		ft := dt.Field(i)
		fv := dest.Field(i)
		fn := ft.Name
		tag := ft.Tag
		if tv := tag.Get("json"); len(tv) > 0 {
			fn = tv
		} else {
			fn = strings.ToLower(fn[:1]) + fn[1:]
		}
		mv := v.MapIndex(reflect.ValueOf(fn))
		if mv.IsValid() {
			if err := setValue(fv, mv.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}

func map2Map(dest reflect.Value, v reflect.Value) error {
	dt := dest.Type()
	kt := dt.Key()
	et := dt.Elem()

	it := v.MapRange()
	for it.Next() {
		vk := it.Key()
		dk := makeValue(kt)
		if err := setValue(dk, vk.Interface()); err != nil {
			return err
		}

		vv := it.Value()
		dv := makeValue(et)
		if err := setValue(dv, vv.Interface()); err != nil {
			return err
		}

		dest.SetMapIndex(dk, dv)
	}

	return nil
}

func copySlice(dest reflect.Value, v reflect.Value) error {
	l := v.Len()
	if l == 0 {
		dest.SetLen(0)
		return nil
	}
	newDest := reflect.MakeSlice(dest.Type(), l, l)
	for i:=0; i<l; i++ {
		if err := setValue(newDest.Index(i), v.Index(i).Interface()); err != nil {
			return err
		}
	}
	dest.Set(newDest)
	return nil
}

func makeValue(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return reflect.Indirect(reflect.New(reflect.TypeOf("")))
		}
		fallthrough
	case reflect.Bool,reflect.Int,reflect.Uint,
			reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
			reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,
			reflect.Float32,reflect.Float64,reflect.String,
			reflect.Array/*,reflect.Map*/,reflect.Struct,
			reflect.Interface/*,reflect.Ptr*/,reflect.Func:
		return reflect.Indirect(reflect.New(t))
	case reflect.Map:
		return reflect.MakeMap(t)
	case reflect.Ptr:
		el := makeValue(t.Elem())
		ptr := reflect.Indirect(reflect.New(t))
		ptr.Set(el.Addr())
		return ptr
	default:
		panic("unsupport type")
	}
}

func makeSlice(el reflect.Type) reflect.Value {
	t := reflect.SliceOf(el)
	return reflect.Indirect(reflect.New(t))
}
