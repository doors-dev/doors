package doors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/doors-dev/gox"
)

// Class returns a [Classes] value holding the given class names.
//
// Each argument is split on whitespace, so Class("a", "b c") and
// Class("a b c") produce the same classes.
func Class(classes ...string) Classes {
	c := Classes{}
	for _, class := range classes {
		for class := range strings.FieldsSeq(class) {
			c.add = append(c.add, class)
		}
	}
	return c
}

// Classes is an immutable list of class names, plus names to omit from output.
//
// Use it as a class attribute value, as an attribute modifier, or as a proxy
// before an element or component. Every operation returns a new value and
// leaves the receiver unchanged.
type Classes struct {
	add    []string
	filter []string
}

// Add returns a new value with classes appended, split like [Class].
func (c Classes) Add(classes ...string) Classes {
	c = c.Clone()
	for _, class := range classes {
		for class := range strings.FieldsSeq(class) {
			c.add = append(c.add, class)
		}
	}
	return c
}

// Remove returns a new value without the listed classes.
//
// Unlike [Classes.Filter], the names are not remembered, so the same class can
// be added again afterwards.
func (c Classes) Remove(classes ...string) Classes {
	add := make([]string, 0, len(c.add))
main:
	for _, class := range c.add {
		for _, removeClasses := range classes {
			for removeClass := range strings.FieldsSeq(removeClasses) {
				if class == removeClass {
					continue main
				}
			}
		}
		add = append(add, class)
	}
	filter := slices.Clone(c.filter)
	return Classes{
		add:    add,
		filter: filter,
	}
}

// Filter returns a new value that omits matching classes from output.
//
// Unlike [Classes.Remove], the names stay omitted even when added later or
// joined in from another value.
func (c Classes) Filter(classes ...string) Classes {
	c = c.Clone()
	for _, removeClass := range classes {
		for removeClass := range strings.FieldsSeq(removeClass) {
			c.filter = append(c.filter, removeClass)
		}
	}
	return c
}

// Join returns a new value that merges classes into the current one.
//
// Names to omit are merged too, so a filter from any joined value applies to
// the whole result.
func (c Classes) Join(classes ...Classes) Classes {
	c = c.Clone()
	for _, classes := range classes {
		c.add = append(c.add, classes.add...)
		c.filter = append(c.filter, classes.filter...)
	}
	return c
}

func (c Classes) Mutate(name string, prev any) any {
	if name != "class" {
		slog.Warn(
			"doors.Class used on a non-class attribute",
			"attribute", name,
			"expected", "class",
		)
	}
	if classes, ok := prev.(Classes); ok {
		return classes.Join(c)
	}
	if s, ok := prev.(string); ok {
		classes := Class(s)
		return classes.Join(c)
	}
	if s, ok := prev.(fmt.Stringer); ok {
		classes := Class(s.String())
		return classes.Join(c)
	}
	return c
}

// Clone returns an independent copy.
func (c Classes) Clone() Classes {
	c.add = slices.Clone(c.add)
	c.filter = slices.Clone(c.filter)
	return c
}

func (c Classes) Modify(ctx context.Context, tag string, atts gox.Attrs) error {
	atts.Get("class").Set(c)
	return nil
}

func (c Classes) Proxy(cur gox.Cursor, el gox.Elem) error {
	return ProxyMod(c).Proxy(cur, el)
}

// String returns the class list as it is rendered in a class attribute.
func (c Classes) String() string {
	buf := bytes.Buffer{}
	if err := c.Output(&buf); err != nil {
		panic(errors.Join(err, errors.New("class buffer output can't error")))
	}
	return buf.String()
}

func (c Classes) Output(w io.Writer) error {
	first := true
main:
	for _, class := range c.add {
		for _, remove := range c.filter {
			if remove == class {
				continue main
			}
		}
		if !first {
			if _, err := io.WriteString(w, " "); err != nil {
				return err
			}
		}
		first = false
		if _, err := io.WriteString(w, class); err != nil {
			return err
		}
	}
	return nil
}
