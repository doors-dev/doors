// Managed by GoX v0.1.28

//line input.gox:1
package attr

import (
	"context"
	"strings"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type inputFragment struct {
	test.NoBeam
	r *test.Reporter
}

//line input.gox:17
func (f *inputFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line input.gox:18
			__e = __c.Set("id", "unfocus"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("dd"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line input.gox:19
		__e = __c.Any(f.r); if __e != nil { return }
//line input.gox:20
		__e = __c.Any(f.focusFields()); if __e != nil { return }
//line input.gox:21
		__e = __c.Any(f.inputFields()); if __e != nil { return }
		__e = __c.InitVoid("hr"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
//line input.gox:23
		__e = __c.Any(f.changeFields()); if __e != nil { return }
	return })
//line input.gox:24
}

func (f *inputFragment) focusioouter() []doors.Attr {
	return []doors.Attr{
		doors.AFocusIn{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 2, "in")
				return false
			},
		},
		doors.AFocusOut{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 2, "out")
				return false
			},
		},
	}
}
func (f *inputFragment) focusio() []doors.Attr {
	return []doors.Attr{
		doors.AFocusIn{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 1, "in")
				return false
			},
		},
		doors.AFocusOut{
			StopPropagation: true,
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 1, "out")
				return false
			},
		},
	}
}
func (f *inputFragment) focus() []doors.Attr {
	return []doors.Attr{
		doors.AFocus{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 0, "focus")
				return false
			},
		},
		doors.ABlur{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 0, "blur")
				return false
			},
		},
	}
}

//line input.gox:76
func (f *inputFragment) focusFields() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("focus"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line input.gox:78
			__e = __c.Modify(doors.A(ctx, f.focusioouter()...)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("div"); if __e != nil { return }
			{
//line input.gox:79
				__e = __c.Modify(doors.A(ctx, f.focusio()...)); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line input.gox:80
					__e = __c.Set("id", "focus"); if __e != nil { return }
//line input.gox:80
					__e = __c.Modify(doors.A(ctx, f.focus()...)); if __e != nil { return }
//line input.gox:80
					__e = __c.Set("type", "text"); if __e != nil { return }
//line input.gox:80
					__e = __c.Set("name", "text"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line input.gox:83
			__e = __c.Set("id", "blur"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line input.gox:84
}

func (f *inputFragment) inputAttr(excudeValue bool) doors.Attr {
	return doors.AInput{
		ExcludeValue: excudeValue,
		On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
			f.r.Update(ctx, 0, r.Event().Data)
			f.r.Update(ctx, 1, r.Event().Value)
			//		fmt.Printf("%+v\n", r.Event())
			return false
		},
	}
}

//line input.gox:98
func (f *inputFragment) inputFields() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("input"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:100
			__e = __c.Set("id", "input"); if __e != nil { return }
//line input.gox:100
			__e = __c.Modify(doors.A(ctx, f.inputAttr(false))); if __e != nil { return }
//line input.gox:100
			__e = __c.Set("type", "text"); if __e != nil { return }
//line input.gox:100
			__e = __c.Set("name", "text"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("input ex"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:102
			__e = __c.Set("id", "input_ex"); if __e != nil { return }
//line input.gox:102
			__e = __c.Modify(doors.A(ctx, f.inputAttr(true))); if __e != nil { return }
//line input.gox:102
			__e = __c.Set("type", "text"); if __e != nil { return }
//line input.gox:102
			__e = __c.Set("name", "text"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line input.gox:103
}

func (f *inputFragment) attr(index string) doors.Attr {
	return doors.AChange{
		On: func(ctx context.Context, r doors.RequestEvent[doors.ChangeEvent]) bool {
			//		fmt.Printf("%+v\n", r.Event())
			if r.Event().Name != index {
				return false
			}
			f.r.Update(ctx, 0, index)
			f.r.Update(ctx, 1, r.Event().Value)
			if r.Event().Number != nil {
				f.r.Update(ctx, 2, test.Float(*r.Event().Number))
			} else {
				f.r.Update(ctx, 2, "")
			}
			if r.Event().Date != nil {
				f.r.Update(ctx, 3, r.Event().Date.String())
			} else {
				f.r.Update(ctx, 3, "")
			}
			s := strings.Join(r.Event().Selected, ",")
			f.r.Update(ctx, 4, s)
			if r.Event().Checked {
				f.r.Update(ctx, 5, "true")
			} else {
				f.r.Update(ctx, 5, "false")
			}
			return false
		},
	}
}

//line input.gox:136
func (f *inputFragment) changeFields() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("text"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:138
			__e = __c.Set("id", "text"); if __e != nil { return }
//line input.gox:138
			__e = __c.Modify(doors.A(ctx, f.attr("text"))); if __e != nil { return }
//line input.gox:138
			__e = __c.Set("type", "text"); if __e != nil { return }
//line input.gox:138
			__e = __c.Set("name", "text"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("password"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:140
			__e = __c.Set("id", "password"); if __e != nil { return }
//line input.gox:140
			__e = __c.Modify(doors.A(ctx, f.attr("password"))); if __e != nil { return }
//line input.gox:140
			__e = __c.Set("type", "password"); if __e != nil { return }
//line input.gox:140
			__e = __c.Set("name", "password"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("email"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:142
			__e = __c.Set("id", "email"); if __e != nil { return }
//line input.gox:142
			__e = __c.Modify(doors.A(ctx, f.attr("email"))); if __e != nil { return }
//line input.gox:142
			__e = __c.Set("type", "email"); if __e != nil { return }
//line input.gox:142
			__e = __c.Set("name", "email"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("tel"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:144
			__e = __c.Set("id", "tel"); if __e != nil { return }
//line input.gox:144
			__e = __c.Modify(doors.A(ctx, f.attr("tel"))); if __e != nil { return }
//line input.gox:144
			__e = __c.Set("type", "tel"); if __e != nil { return }
//line input.gox:144
			__e = __c.Set("name", "tel"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("url"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:146
			__e = __c.Set("id", "url"); if __e != nil { return }
//line input.gox:146
			__e = __c.Modify(doors.A(ctx, f.attr("url"))); if __e != nil { return }
//line input.gox:146
			__e = __c.Set("type", "url"); if __e != nil { return }
//line input.gox:146
			__e = __c.Set("name", "url"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("search"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:148
			__e = __c.Set("id", "search"); if __e != nil { return }
//line input.gox:148
			__e = __c.Modify(doors.A(ctx, f.attr("search"))); if __e != nil { return }
//line input.gox:148
			__e = __c.Set("type", "search"); if __e != nil { return }
//line input.gox:148
			__e = __c.Set("name", "search"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("number"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:150
			__e = __c.Set("id", "number"); if __e != nil { return }
//line input.gox:150
			__e = __c.Modify(doors.A(ctx, f.attr("number"))); if __e != nil { return }
//line input.gox:150
			__e = __c.Set("type", "number"); if __e != nil { return }
//line input.gox:150
			__e = __c.Set("name", "number"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("date"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:152
			__e = __c.Set("id", "date"); if __e != nil { return }
//line input.gox:152
			__e = __c.Modify(doors.A(ctx, f.attr("date"))); if __e != nil { return }
//line input.gox:152
			__e = __c.Set("type", "date"); if __e != nil { return }
//line input.gox:152
			__e = __c.Set("name", "date"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("datetime-local"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:154
			__e = __c.Set("id", "datetime-local"); if __e != nil { return }
//line input.gox:154
			__e = __c.Modify(doors.A(ctx, f.attr("datetime-local"))); if __e != nil { return }
//line input.gox:154
			__e = __c.Set("type", "datetime-local"); if __e != nil { return }
//line input.gox:154
			__e = __c.Set("name", "datetime-local"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("month"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:156
			__e = __c.Set("id", "month"); if __e != nil { return }
//line input.gox:156
			__e = __c.Modify(doors.A(ctx, f.attr("month"))); if __e != nil { return }
//line input.gox:156
			__e = __c.Set("type", "month"); if __e != nil { return }
//line input.gox:156
			__e = __c.Set("name", "month"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("time"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:158
			__e = __c.Set("id", "time"); if __e != nil { return }
//line input.gox:158
			__e = __c.Modify(doors.A(ctx, f.attr("time"))); if __e != nil { return }
//line input.gox:158
			__e = __c.Set("type", "time"); if __e != nil { return }
//line input.gox:158
			__e = __c.Set("name", "time"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("color"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:160
			__e = __c.Set("id", "color"); if __e != nil { return }
//line input.gox:160
			__e = __c.Modify(doors.A(ctx, f.attr("color"))); if __e != nil { return }
//line input.gox:160
			__e = __c.Set("type", "color"); if __e != nil { return }
//line input.gox:160
			__e = __c.Set("name", "color"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("checkbox"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line input.gox:162
			__e = __c.Set("id", "checkbox"); if __e != nil { return }
//line input.gox:162
			__e = __c.Modify(doors.A(ctx, f.attr("checkbox"))); if __e != nil { return }
//line input.gox:162
			__e = __c.Set("type", "checkbox"); if __e != nil { return }
//line input.gox:162
			__e = __c.Set("name", "checkbox"); if __e != nil { return }
//line input.gox:162
			__e = __c.Set("value", "on"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("radio"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line input.gox:165
		radio := doors.A(ctx, f.attr("radio"))

//line input.gox:167
		__e = (radio).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line input.gox:167
				__e = __c.Set("id", "radio-1"); if __e != nil { return }
//line input.gox:167
				__e = __c.Set("type", "radio"); if __e != nil { return }
//line input.gox:167
				__e = __c.Set("name", "radio"); if __e != nil { return }
//line input.gox:167
				__e = __c.Set("value", "option1"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
//line input.gox:168
		__e = (radio).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line input.gox:168
				__e = __c.Set("id", "radio-2"); if __e != nil { return }
//line input.gox:168
				__e = __c.Set("type", "radio"); if __e != nil { return }
//line input.gox:168
				__e = __c.Set("name", "radio"); if __e != nil { return }
//line input.gox:168
				__e = __c.Set("value", "option2"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("textarea"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("textarea"); if __e != nil { return }
		{
//line input.gox:170
			__e = __c.Set("id", "textarea"); if __e != nil { return }
//line input.gox:170
			__e = __c.Modify(doors.A(ctx, f.attr("textarea"))); if __e != nil { return }
//line input.gox:170
			__e = __c.Set("name", "textarea"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("select"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("select"); if __e != nil { return }
		{
//line input.gox:172
			__e = __c.Set("id", "select"); if __e != nil { return }
//line input.gox:172
			__e = __c.Modify(doors.A(ctx, f.attr("select"))); if __e != nil { return }
//line input.gox:172
			__e = __c.Set("name", "select"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("option"); if __e != nil { return }
			{
//line input.gox:173
				__e = __c.Set("value", "option1"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Option 1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("option"); if __e != nil { return }
			{
//line input.gox:174
				__e = __c.Set("value", "option2"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Option 2"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("h3"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("multiselect"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("select"); if __e != nil { return }
		{
//line input.gox:177
			__e = __c.Set("id", "multiselect"); if __e != nil { return }
//line input.gox:177
			__e = __c.Modify(doors.A(ctx, f.attr("multiselect"))); if __e != nil { return }
//line input.gox:177
			__e = __c.Set("name", "multiselect"); if __e != nil { return }
			__e = __c.Set("multiple", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("option"); if __e != nil { return }
			{
//line input.gox:178
				__e = __c.Set("value", "option1"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Option 1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("option"); if __e != nil { return }
			{
//line input.gox:179
				__e = __c.Set("value", "option2"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("Option 2"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line input.gox:181
}
