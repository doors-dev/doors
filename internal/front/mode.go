package front

import "encoding/json"

type HookMode interface {
	MarshalJSON() ([]byte, error)
	hookMode() hookMode
}
type hookMode struct {
	value string
	args  []any
}

func (h hookMode) hookMode() hookMode {
	return h
}

func (h hookMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(append([]any{h.value}, h.args...))
}

var none []any = make([]any, 0)

func Default() HookMode {
	return &hookMode{
		value: "",
		args:  none,
	}
}

func Block() HookMode {
	return &hookMode{
		value: "block",
		args:  none,
	}
}

func Frame() HookMode {
	return &hookMode{
		value: "frame",
		args:  none,
	}
}

func Butter() HookMode {
	return &hookMode{
		value: "butter",
		args:  none,
	}
}

func Debounce(duration int, limit int) HookMode {
	if duration <= 0 {
		return &hookMode{
			value: "",
			args:  none,
		}
	}
	if limit < 0 {
		limit = 0
	}
	return &hookMode{
		value: "debounce",
		args:  []any{duration, limit},
	}
}
