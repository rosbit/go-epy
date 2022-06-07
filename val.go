package epy

import (
	sltime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"reflect"
	"time"
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
		var res map[interface{}]interface{}
		res2 := make(map[string]interface{})
		allKeyString := true

		ks := d.Keys()
		for _, k := range ks {
			vv, _, _ := d.Get(k)
			key := fromValue(k)
			if allKeyString {
				if strKey, ok := key.(string); ok {
					res2[strKey] = fromValue(vv)
					continue
				}

				allKeyString = false
				res = make(map[interface{}]interface{})
				for sk, sv := range res2 {
					res[sk] = sv
				}
			}
			res[key] = fromValue(vv)
		}

		if allKeyString {
			return res2
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

