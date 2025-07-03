package front

import (
	"encoding/json"
	"net/http"
)

type After struct {
	Name string
	Arg  any
}

func (a *After) MarshalJSON() ([]byte, error) {
	arr := []any{a.Name, a.Arg}
	return json.Marshal(arr)
}

func (a *After) Set(h http.Header) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	h.Set("D00r-After", string(b))
	return nil
}
