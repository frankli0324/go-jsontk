package json

import (
	"reflect"
	"strings"
	"sync"
)

func buildStructIndex(v reflect.Type, now []int, into map[string][]int) {
	hasAnon := false
	for i, nf := 0, v.NumField(); i < nf; i++ {
		f := v.Field(i)
		if f.Anonymous {
			hasAnon = true
			continue
		}
		name := f.Name
		if s := strings.Split(f.Tag.Get("json"), ","); len(s) != 0 && s[0] != "" {
			name = s[0]
		}
		if _, ok := into[name]; ok {
			continue
		}
		into[name] = append(append([]int(nil), now...), i)
	}
	if !hasAnon {
		return
	}
	for i, nf := 0, v.NumField(); i < nf; i++ {
		f := v.Field(i)
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				buildStructIndex(v, append(append([]int(nil), now...), i), into)
			} else {
				name := f.Name
				if s := strings.Split(f.Tag.Get("json"), ","); len(s) != 0 && s[0] != "" {
					name = s[0]
				}
				if _, ok := into[name]; ok {
					continue
				}
				into[name] = append(append([]int(nil), now...), i)
			}
		}
	}

}

var cache sync.Map

func cachedStructIndex(v reflect.Type) map[string][]int {
	if v, ok := cache.Load(v); ok {
		return v.(map[string][]int)
	}
	m := make(map[string][]int)
	buildStructIndex(v, nil, m)
	cache.Store(v, m)
	return m
}
