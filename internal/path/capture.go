package path

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type marker struct {
	set func(any) error
	get func(any) bool
}

func newMarker(f field) *marker {
	set := func(a any) error {
		v := reflect.ValueOf(a).Elem()
		v.Field(f.i).SetBool(true)
		return nil
	}
	get := func(a any) bool {
		v := reflect.ValueOf(a).Elem()
		return v.Field(f.i).Bool()
	}
	return &marker{
		set: set,
		get: get,
	}
}

type capture struct {
	get func(any) string
	set func(any, string) error
}

func newCapture(f field, a *atom) (*capture, error) {
	if a.isEnd() {
		return newToEndCapture(f)
	}
	kind := f.f.Type.Kind()
	if kind == reflect.String {
		return newStringCapture(f)
	}
	if kind == reflect.Int {
		return newIntCapture(f)
	}
	if kind == reflect.Float64 {
		return newFloat64Capture(f)
	}
	return nil, errors.New(fmt.Sprint("Field type for capture is not supported ", kind))
}

func newFloat64Capture(f field) (*capture, error) {
	return &capture{
		set: func(m any, value string) error {
			v := reflect.ValueOf(m).Elem()
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			v.Field(f.i).SetFloat(num)
			return nil
		},
		get: func(m any) string {
			v := reflect.ValueOf(m)
			num := v.Field(f.i).Float()
			return strconv.FormatFloat(num, 'f', -1, 64)
		},
	}, nil
}

func newIntCapture(f field) (*capture, error) {
	return &capture{
		set: func(m any, value string) error {
			v := reflect.ValueOf(m).Elem()
			num, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			v.Field(f.i).SetInt(int64(num))
			return nil
		},
		get: func(m any) string {
			v := reflect.ValueOf(m).Elem()
			num := v.Field(f.i).Int()
			return strconv.Itoa(int(num))
		},
	}, nil
}

func newStringCapture(f field) (*capture, error) {
	return &capture{
		set: func(m any, value string) error {
			v := reflect.ValueOf(m).Elem()
			var err error
			value, err = url.QueryUnescape(value)
			if err != nil {
				return err
			}
			v.Field(f.i).SetString(value)
			return nil
		},
		get: func(m any) string {
			v := reflect.ValueOf(m).Elem()
			return v.Field(f.i).String()
		},
	}, nil
}

func newToEndCapture(f field) (*capture, error) {
	var set func(any, string) error
	var get func(any) string
	if f.f.Type.Kind() == reflect.String {
		set = func(m any, value string) error {
			v := reflect.ValueOf(m).Elem()
			v.Field(f.i).SetString(value)
			return nil
		}
		get = func(m any) string {
			v := reflect.ValueOf(m).Elem()
			return v.Field(f.i).String()
		}
	} else if f.f.Type.Kind() == reflect.Slice && f.f.Type.Elem().Kind() == reflect.String {
		set = func(m any, value string) error {
			vals := make([]string, 0)
			for part := range strings.SplitSeq(string(value), "/") {
				value, error := url.QueryUnescape(part)
				if error != nil {
					return error
				}
				vals = append(vals, value)
			}
			v := reflect.ValueOf(m).Elem()
			v.Field(f.i).Set(reflect.ValueOf(vals))
			return nil
		}
		get = func(m any) string {
			v := reflect.ValueOf(m).Elem()
			values := v.Field(f.i).Interface().([]string)
			parts := make([]string, len(values))
			for _, part := range values {
				parts = append(parts, url.QueryEscape(part))
			}
			return strings.Join(parts, "/")
		}
	} else {
		return nil, errors.New("Capture to end must me string or []string")
	}
	return &capture{
		set: set,
		get: get,
	}, nil
}
