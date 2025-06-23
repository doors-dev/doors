package path

import (
	"errors"
	"log"
)

type Capture int

const (
	NoCapture Capture = iota
	CapturePart
	CaptureToEnd
)

type atom struct {
	runes       []rune
	captureType Capture
	tail        bool
	capture     *capture
}

func newAtom() *atom {
	return &atom{
		runes:       make([]rune, 0),
		captureType: NoCapture,
		tail:        false,
	}
}

func (a *atom) encode(m any) ([]string, error) {
	if a.captureType == NoCapture {
		parts := []string{string(a.runes)}
		if a.tail {
			parts = append(parts, "/")
		}
		return parts, nil
	}
	part := a.capture.get(m)
	if len(part) == 0 {
		return nil, errors.New("Part value cannot be empty")
	}
	parts := []string{part}
	if a.tail {
		parts = append(parts, "/")
	}
	return parts, nil
}

func (a *atom) decode(p []rune, next []*atom) (mutations, bool) {
	if a.captureType == CaptureToEnd {
		return a.matchCaptureToEnd(p)
	}
	if a.captureType == NoCapture {
		return a.matchNoCapture(p, next)
	}
	return a.matchCapturePart(p, next)
}

func (a *atom) matchCaptureToEnd(p []rune) (mutations, bool) {
	muts := []mutation{func(m any) error {
		return a.capture.set(m, string(p))
	}}
	return muts, true
}

func (a *atom) matchNoCapture(p []rune, next []*atom) (mutations, bool) {
	if len(p) < len(a.runes) {
		return nil, false
	}
	for i, r := range a.runes {
		if r != p[i] {
			return nil, false
		}
	}
	start := len(a.runes)
	if len(p) == start {
		if len(next) != 0 {
			return nil, false
		}
		return []mutation{}, true
	}
	if a.tail {
		if p[start] != '/' {
			return nil, false
		}
		start += 1
	}
	if len(next) == 0 {
		if len(p) == start {
			return []mutation{}, true
		}
		return nil, false
	}
	return next[0].decode(p[start:], next[1:])
}

func (a *atom) matchCapturePart(p []rune, next []*atom) (mutations, bool) {
	if !a.tail {
		log.Fatalf("Capture part is not tail")
	}
	value := make([]rune, 0)
	start := 0
	for start < len(p) {
		if p[start] == '/' {
			start += 1
			break
		}
		value = append(value, p[start])
		start += 1
	}
	if len(value) == 0 {
		return nil, false
	}
	muts := []mutation{func(m any) error {
		return a.capture.set(m, string(value))
	}}
	if len(p) == start {
		if len(next) == 0 {
			return muts, true
		}
		return nil, false
	}
    if len(p) > start && len(next) == 0 {
        return nil, false
    }
	m, ok := next[0].decode(p[start:], next[1:])
	if !ok {
		return nil, false
	}
	return append(muts, m...), true
}

func (a *atom) collectParams(s map[string][]*atom) {
	if a.captureType == CapturePart || a.captureType == CaptureToEnd {
		name := string(a.runes)
		arr, has := s[name]
		if !has {
			arr = []*atom{a}
		} else {
			arr = append(arr, a)
		}
		s[name] = arr
	}
}

func (a *atom) setCapture(c *capture) {
	a.capture = c
}

func (a *atom) addTo(branch *branch) error {
	if len(a.runes) == 0 {
		if a.captureType == CapturePart || a.captureType == CaptureToEnd {
			return errors.New("Capture syntax error: capture name not provided")
		}
		if a.tail {
			branch.setLastTail()
		}
		return nil
	}
	if a.captureType == CapturePart && !a.tail {
		return errors.New("Capture synax error: capture part must be tail")
	}
	return branch.append(a)
}

func (a *atom) append(r rune) error {
	if a.captureType == CaptureToEnd {
		return errors.New("To end capture syntax error")
	}
	a.runes = append(a.runes, r)
	return nil
}

func (a *atom) capturePart() error {
	if len(a.runes) != 0 {
		return errors.New("Capture syntax error, you must start part with :")
	}
	a.captureType = CapturePart
	return nil
}

func (a *atom) setTail() {
	a.tail = true
}

func (a *atom) isEnd() bool {
	return a.captureType == CaptureToEnd
}

func (a *atom) captureToEnd() error {
	if a.captureType != CapturePart {
		return errors.New("Capture to end syntax error: provide field name")
	}
	if len(a.runes) == 0 {
		return errors.New("Capture syntax error, name not provided")
	}
	a.captureType = CaptureToEnd
	return nil
}
