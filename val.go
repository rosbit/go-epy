package slx

import (
	"go.starlark.net/starlark"
	sltime "go.starlark.net/lib/time"
	"reflect"
	"fmt"
	"time"
)

func toValue(v interface{}) starlark.Value {
	if v == nil {
		return starlark.None
	}

	switch vv := v.(type) {
	case int:
		return starlark.MakeInt(vv)
	case uint:
		return starlark.MakeUint(vv)
	case int8,int16,int32,int64:
		return starlark.MakeInt64(reflect.ValueOf(v).Int())
	case uint8,uint16,uint32,uint64:
		return starlark.MakeUint64(reflect.ValueOf(v).Uint())
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
			vt := v2
			r := make([]starlark.Value, vt.Len())
			for i:=0; i<vt.Len(); i++ {
				r[i] = toValue(vt.Index(i).Interface())
			}
			return starlark.NewList(r)
		case reflect.Map:
			vm := v2
			r := starlark.NewDict(vm.Len())
			iter := vm.MapRange()
			for iter.Next() {
				k, v1 := iter.Key(), iter.Value()
				if k.Kind() == reflect.String {
					switch {
					case v1.IsNil():
						r.SetKey(toValue(k.Interface()), starlark.None)
						continue
					case v1.Kind() == reflect.Func:
						if f, err := bindGoFunc(k.String(), v1.Interface()); err == nil {
							r.SetKey(toValue(k.Interface()), f)
						}
						continue
					case v1.Kind() == reflect.Struct:
						r.SetKey(toValue(k.Interface()), bindGoStruct(k.String(), v1))
						continue
					case v1.Kind() == reflect.Ptr && v1.Elem().Kind() == reflect.Struct:
						r.SetKey(toValue(k.Interface()), bindGoStruct(k.String(), v1))
						continue
					default:
						if sv, ok := v1.Interface().(starlark.Value); ok {
							r.SetKey(toValue(k.Interface()), sv)
							continue
						}
					}
				}
				r.SetKey(toValue(k.Interface()), toValue(v1.Interface()))
			}
			return r
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
		return 0
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
		res := make(map[string]interface{})
		ks := d.Keys()
		for _, k := range ks {
			vv, _, _ := d.Get(k)
			res[k.String()] = fromValue(vv)
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
	default:
		return nil
	}
}

func setValue(dest reflect.Value, val interface{}) error {
	v := reflect.ValueOf(val)
	vt := reflect.TypeOf(val)
	dt := dest.Type()
	if vt.AssignableTo(dt) {
		dest.Set(v)
		return nil
	}

	if vt.ConvertibleTo(dt) {
		dest.Set(v.Convert(dt))
		return nil
	}

	return fmt.Errorf("cannot convert %s to %s", vt, dt)
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
			reflect.Array, reflect.Map, reflect.Struct:
		return reflect.Indirect(reflect.New(t))
	case reflect.Ptr:
		e := makeValue(t.Elem())
		return e.Addr()
	default:
		panic("unsupport type")
	}
}

func makeSlice(el reflect.Type) reflect.Value {
	t := reflect.SliceOf(el)
	return reflect.Indirect(reflect.New(t))
}
