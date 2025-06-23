package common

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"

	"github.com/mr-tron/base58"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

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

func BadPanic(v any) {
	//defer log.Fatal("Critical error, execution stopped")
	panic(v)
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
