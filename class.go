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

// Class returns a Classes value from one or more class strings.
//
// Each argument is split with strings.Fields, so Class("a", "b c") and
// Class("a b c") produce the same classes. The returned value can be used as a
// class attribute value, as an attribute modifier, or as a proxy before an
// element/component.
func Class(classes ...string) Classes {
	c := Classes{}
	for _, class := range classes {
		for class := range strings.FieldsSeq(class) {
			c.add = append(c.add, class)
		}
	}
	return c
}

// Classes describes class names to add and filter out.
//
// Classes values are immutable: Add, Remove, Filter, Join, and Clone return a
// new value and leave the receiver unchanged.
type Classes struct {
	add    []string
	filter []string
}

// Add returns a new Classes value with classes appended.
//
// Arguments are split like Class, so Add("a", "b c") adds three classes.
func (c Classes) Add(classes ...string) Classes {
	c = c.Clone()
	for _, class := range classes {
		for class := range strings.FieldsSeq(class) {
			c.add = append(c.add, class)
		}
	}
	return c
}

// Remove returns a new Classes value with matching currently-added classes
// removed.
//
// Removed classes are not remembered: the same class can be added again later.
// Use Filter when matching classes should be omitted from final output even if
// they are added later or come from a joined Classes value.
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

// Filter returns a new Classes value that omits matching classes from output.
//
// Removed classes are filtered regardless of whether they were added before or
// after Filter was called.
func (c Classes) Filter(classes ...string) Classes {
	c = c.Clone()
	for _, removeClass := range classes {
		for removeClass := range strings.FieldsSeq(removeClass) {
			c.filter = append(c.filter, removeClass)
		}
	}
	return c
}

// Join returns a new Classes value that combines several class modifiers.
//
// Both added and filtered class names are preserved, so filters from joined
// values still affect the final rendered class list.
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

// Clone returns an independent copy of c.
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

// String returns the class list as it would be rendered in a class attribute.
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
