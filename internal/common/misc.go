package common

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"time"
	"unsafe"

	"github.com/a-h/templ"
	"github.com/mr-tron/base58"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

var bytesNull = []byte("null")

func MarshalJSON(value any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(value)
	if err != nil {
		return bytesNull, err
	}
	b := StripN(buf.Bytes())
	return b, nil
}

func Ts() {
	fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
}

func AsString(buf *[]byte) string {
	return *(*string)(unsafe.Pointer(buf))
}

func AsBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func StripN(buf []byte) []byte {
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	return buf
}

func IsNill(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func Debug(a any) string {
	return fmt.Sprintf("%+v", a)
}

func RandId() string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("failed to generate random bytes: %v", err)
	}
	return base58.Encode(randomBytes)
}

func Zip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(input)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MinifyCSS(input []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	return m.Bytes("text/css", input)
}

func Hash(input []byte) string {
	hash := crypto.SHA3_224.New()
	hash.Write(input)
	return base58.Encode(hash.Sum(nil)[0:12])
}

func Catch(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	f()
	return
}

func CatchValue[V any](f func() V) (value V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	value = f()
	return
}

var nopPointer = uintptr(reflect.ValueOf(templ.NopComponent).UnsafePointer())

func GetChildren(ctx context.Context) (context.Context, templ.Component, bool) {
	c := templ.GetChildren(ctx)
	if uintptr(reflect.ValueOf(c).UnsafePointer()) == nopPointer {
		return ctx, nil, false
	}
	return templ.ClearChildren(ctx), c, true
}
