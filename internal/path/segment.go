package path

import "reflect"

func newLiteralSegment(value string) segment {
	return segment{
		entity: literalSegment(value),
	}
}

func newSingleSegment(field singleField, optional bool) segment {
	return segment{
		entity: singleSegment{
			field:    field,
			optional: optional,
		},
	}
}

func newMultiSegment(field multiField, optional bool) segment {
	return segment{
		entity: multiSegment{
			field:    field,
			optional: optional,
		},
	}
}

type segment struct {
	entity any
}

func (a segment) literal() (string, bool) {
	c, ok := a.entity.(literalSegment)
	return string(c), ok
}

func (a segment) single() (singleSegment, bool) {
	c, ok := a.entity.(singleSegment)
	return c, ok
}

func (a segment) multi() (multiSegment, bool) {
	c, ok := a.entity.(multiSegment)
	return c, ok
}

type literalSegment string

type singleSegment struct {
	field    singleField
	optional bool
}

func (c singleSegment) get(m reflect.Value) (string, bool) {
	return c.field.get(m)
}

func (c singleSegment) set(m reflect.Value, v string) bool {
	return c.field.set(m, v)
}

type multiSegment struct {
	field    multiField
	optional bool
}

func (c multiSegment) get(m reflect.Value) []string {
	return c.field.get(m)
}

func (c multiSegment) set(m reflect.Value, v []string) {
	if len(v) == 0 {
		if !c.optional {
			panic("non-option field can't receive empty value")
		}
	}
	c.field.set(m, v)
}
