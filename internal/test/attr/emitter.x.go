// Managed by GoX v0.2.3

//line emitter.gox:1
package attr

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

// pointer

type emitterPointerFragment struct {
	test.NoBeam
	r *test.Reporter
	e doors.Emitter
}

func (f *emitterPointerFragment) attrs() []doors.Attr {
	return []doors.Attr{
		doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 0, "click")
				f.r.Update(ctx, 1, fmt.Sprint(r.Event().Button))
				f.r.Update(ctx, 2, fmt.Sprint(r.Event().Buttons))
				return false
			},
		},
		doors.APointerDown{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 0, "pointerdown")
				return false
			},
		},
		doors.APointerUp{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 0, "pointerup")
				return false
			},
		},
	}
}

//line emitter.gox:45
func (f *emitterPointerFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:47
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")
	f.r.Update(ctx, 2, "")

//line emitter.gox:51
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:52
			__e = __c.Set("id", "target"); if __e != nil { return }
//line emitter.gox:52
			__e = __c.Modify(&f.e); if __e != nil { return }
//line emitter.gox:52
			__e = __c.Modify(doors.A(ctx, f.attrs()...)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line emitter.gox:53
		__e = __c.Any(test.Button("emit-click", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.Click(doors.PointerEmit{Button: 2, Buttons: 2}))
		return false
	})); if __e != nil { return }
//line emitter.gox:57
		__e = __c.Any(test.Button("emit-down", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.PointerDown(doors.PointerEmit{}))
		return false
	})); if __e != nil { return }
//line emitter.gox:61
		__e = __c.Any(test.Button("emit-up", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.PointerUp(doors.PointerEmit{}))
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:65
}

// keyboard

type emitterKeyFragment struct {
	test.NoBeam
	r *test.Reporter
	e doors.Emitter
}

func (f *emitterKeyFragment) attrs() []doors.Attr {
	return []doors.Attr{
		doors.AKeyDown{
			On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
				f.r.Update(ctx, 0, "keydown")
				f.r.Update(ctx, 1, r.Event().Key)
				f.r.Update(ctx, 2, r.Event().Code)
				return false
			},
		},
		doors.AKeyUp{
			On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
				f.r.Update(ctx, 0, "keyup")
				f.r.Update(ctx, 1, r.Event().Key)
				f.r.Update(ctx, 3, fmt.Sprint(r.Event().ShiftKey))
				f.r.Update(ctx, 4, fmt.Sprint(r.Event().CtrlKey))
				return false
			},
		},
	}
}

//line emitter.gox:97
func (f *emitterKeyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:99
		for i := 0; i < 5; i++ {
		f.r.Update(ctx, i, "")
	}

//line emitter.gox:103
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line emitter.gox:104
			__e = __c.Set("type", "text"); if __e != nil { return }
//line emitter.gox:104
			__e = __c.Set("id", "target"); if __e != nil { return }
//line emitter.gox:104
			__e = __c.Modify(&f.e); if __e != nil { return }
//line emitter.gox:104
			__e = __c.Modify(doors.A(ctx, f.attrs()...)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
//line emitter.gox:105
		__e = __c.Any(test.Button("emit-keydown", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.KeyDown(doors.KeyboardEmit{Key: "a", Code: "KeyA"}))
		return false
	})); if __e != nil { return }
//line emitter.gox:109
		__e = __c.Any(test.Button("emit-keyup", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.KeyUp(doors.KeyboardEmit{Key: "A", Code: "KeyA", ShiftKey: true, CtrlKey: true}))
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:113
}

// focus

type emitterFocusFragment struct {
	test.NoBeam
	r *test.Reporter
	e doors.Emitter
}

func (f *emitterFocusFragment) inner() []doors.Attr {
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

func (f *emitterFocusFragment) outer() []doors.Attr {
	return []doors.Attr{
		doors.AFocusIn{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 1, "in")
				return false
			},
		},
		doors.AFocusOut{
			On: func(ctx context.Context, r doors.RequestEvent[doors.FocusEvent]) bool {
				f.r.Update(ctx, 1, "out")
				return false
			},
		},
	}
}

//line emitter.gox:157
func (f *emitterFocusFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:159
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")

//line emitter.gox:162
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:163
			__e = __c.Modify(doors.A(ctx, f.outer()...)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.InitVoid("input"); if __e != nil { return }
			{
//line emitter.gox:164
				__e = __c.Set("type", "text"); if __e != nil { return }
//line emitter.gox:164
				__e = __c.Set("id", "target"); if __e != nil { return }
//line emitter.gox:164
				__e = __c.Modify(&f.e); if __e != nil { return }
//line emitter.gox:164
				__e = __c.Modify(doors.A(ctx, f.inner()...)); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line emitter.gox:166
		__e = __c.Any(test.Button("emit-focus", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.Focus(doors.FocusEmit{}))
		return false
	})); if __e != nil { return }
//line emitter.gox:170
		__e = __c.Any(test.Button("emit-blur", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.Blur(doors.FocusEmit{}))
		return false
	})); if __e != nil { return }
//line emitter.gox:174
		__e = __c.Any(test.Button("emit-focusin", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.FocusIn(doors.FocusEmit{}))
		return false
	})); if __e != nil { return }
//line emitter.gox:178
		__e = __c.Any(test.Button("emit-focusout", func(ctx context.Context) bool {
		doors.Call(ctx, f.e.FocusOut(doors.FocusEmit{}))
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:182
}

// input, change, submit

type emitterFormFragment struct {
	test.NoBeam
	r *test.Reporter
	ei doors.Emitter
	es doors.Emitter
}

func (f *emitterFormFragment) fieldAttrs() []doors.Attr {
	return []doors.Attr{
		doors.AInput{
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				f.r.Update(ctx, 0, "input")
				f.r.Update(ctx, 1, r.Event().Data)
				return false
			},
		},
		doors.AChange{
			On: func(ctx context.Context, r doors.RequestEvent[doors.ChangeEvent]) bool {
				f.r.Update(ctx, 2, "change")
				f.r.Update(ctx, 3, r.Event().Name)
				return false
			},
		},
	}
}

func (f *emitterFormFragment) submit() doors.Attr {
	return doors.ASubmit[struct{}]{
		On: func(ctx context.Context, r doors.RequestForm[struct{}]) bool {
			f.r.Update(ctx, 4, "submit")
			return false
		},
	}
}

//line emitter.gox:221
func (f *emitterFormFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:223
		for i := 0; i < 5; i++ {
		f.r.Update(ctx, i, "")
	}

//line emitter.gox:227
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line emitter.gox:228
			__e = __c.Set("type", "text"); if __e != nil { return }
//line emitter.gox:228
			__e = __c.Set("id", "field"); if __e != nil { return }
//line emitter.gox:228
			__e = __c.Set("name", "field"); if __e != nil { return }
//line emitter.gox:228
			__e = __c.Modify(&f.ei); if __e != nil { return }
//line emitter.gox:228
			__e = __c.Modify(doors.A(ctx, f.fieldAttrs()...)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
//line emitter.gox:229
		__e = (f.submit()).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("form"); if __e != nil { return }
			{
//line emitter.gox:229
				__e = __c.Set("id", "form"); if __e != nil { return }
//line emitter.gox:229
				__e = __c.Modify(&f.es); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("input"); if __e != nil { return }
				{
//line emitter.gox:230
					__e = __c.Set("type", "text"); if __e != nil { return }
//line emitter.gox:230
					__e = __c.Set("name", "field"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line emitter.gox:232
		__e = __c.Any(test.Button("emit-input", func(ctx context.Context) bool {
		doors.Call(ctx, f.ei.Input(doors.InputEmit{Data: "hey"}))
		return false
	})); if __e != nil { return }
//line emitter.gox:236
		__e = __c.Any(test.Button("emit-change", func(ctx context.Context) bool {
		doors.Call(ctx, f.ei.Change(doors.ChangeEmit{}))
		return false
	})); if __e != nil { return }
//line emitter.gox:240
		__e = __c.Any(test.Button("emit-submit", func(ctx context.Context) bool {
		doors.Call(ctx, f.es.Submit(doors.SubmitEmit{}))
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:244
}

// multiple elements + capture count

type emitterMultiFragment struct {
	test.NoBeam
	r *test.Reporter
	e doors.Emitter
}

func (f *emitterMultiFragment) hit(slot int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, slot, "hit")
			return false
		},
	}
}

//line emitter.gox:263
func (f *emitterMultiFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:265
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")
	f.r.Update(ctx, 2, "")

//line emitter.gox:269
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:270
			__e = __c.Set("id", "m1"); if __e != nil { return }
//line emitter.gox:270
			__e = __c.Modify(&f.e); if __e != nil { return }
//line emitter.gox:270
			__e = __c.Modify(doors.A(ctx, f.hit(1))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("m1"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:271
			__e = __c.Set("id", "m2"); if __e != nil { return }
//line emitter.gox:271
			__e = __c.Modify(&f.e); if __e != nil { return }
//line emitter.gox:271
			__e = __c.Modify(doors.A(ctx, f.hit(2))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("m2"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line emitter.gox:272
		__e = __c.Any(test.Button("emit-count", func(ctx context.Context) bool {
		var count int
		ch := doors.Call(ctx, f.e.Click(doors.PointerEmit{}).Into(&count))
		select {
		case err := <-ch:
			if err != nil {
				f.r.Update(ctx, 0, "err")
			} else {
				f.r.Update(ctx, 0, fmt.Sprint(count))
			}
		case <-ctx.Done():
		}
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:286
}

// several emitters per element

func emitterCount(r *test.Reporter, slot int, action doors.ActionInto[int]) func(context.Context) bool {
	return func(ctx context.Context) bool {
		var count int
		ch := doors.Call(ctx, action.Into(&count))
		select {
		case err := <-ch:
			if err != nil {
				r.Update(ctx, slot, "err")
			} else {
				r.Update(ctx, slot, fmt.Sprint(count))
			}
		case <-ctx.Done():
		}
		return false
	}
}

type emitterSharedFragment struct {
	test.NoBeam
	r *test.Reporter
	a doors.Emitter
	b doors.Emitter
	c doors.Emitter
	s doors.Setter
}

func (f *emitterSharedFragment) hit(slot int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, slot, fmt.Sprint(r.Event().Button))
			return false
		},
	}
}

//line emitter.gox:325
func (f *emitterSharedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:327
		for i := 0; i < 7; i++ {
		f.r.Update(ctx, i, "")
	}

//line emitter.gox:331
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:332
			__e = __c.Set("id", "s1"); if __e != nil { return }
//line emitter.gox:332
			__e = __c.Modify(&f.a); if __e != nil { return }
//line emitter.gox:332
			__e = __c.Modify(&f.b); if __e != nil { return }
//line emitter.gox:332
			__e = __c.Modify(&f.s); if __e != nil { return }
//line emitter.gox:332
			__e = __c.Modify(doors.A(ctx, f.hit(3))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("s1"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:333
			__e = __c.Set("id", "s2"); if __e != nil { return }
//line emitter.gox:333
			__e = __c.Modify(&f.b); if __e != nil { return }
//line emitter.gox:333
			__e = __c.Modify(doors.A(ctx, f.hit(4))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("s2"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:334
			__e = __c.Set("id", "s3"); if __e != nil { return }
//line emitter.gox:334
			__e = __c.Modify(&f.c); if __e != nil { return }
//line emitter.gox:334
			__e = __c.Modify(&f.c); if __e != nil { return }
//line emitter.gox:334
			__e = __c.Modify(doors.A(ctx, f.hit(5))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("s3"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line emitter.gox:335
		__e = __c.Any(test.Button("emit-a", emitterCount(f.r, 0, f.a.Click(doors.PointerEmit{Button: 1})))); if __e != nil { return }
//line emitter.gox:336
		__e = __c.Any(test.Button("emit-b", emitterCount(f.r, 1, f.b.Click(doors.PointerEmit{Button: 2})))); if __e != nil { return }
//line emitter.gox:337
		__e = __c.Any(test.Button("emit-c", emitterCount(f.r, 2, f.c.Click(doors.PointerEmit{Button: 3})))); if __e != nil { return }
//line emitter.gox:338
		__e = __c.Any(test.Button("set", emitterCount(f.r, 6, f.s.Set("data-test", "x")))); if __e != nil { return }
	return })
//line emitter.gox:339
}

// emitter on a door container element

type emitterContainerFragment struct {
	test.NoBeam
	r *test.Reporter
	d doors.Door
	e doors.Emitter
}

func (f *emitterContainerFragment) hit(slot int) doors.Attr {
	return doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, slot, fmt.Sprint(r.Event().Button))
			return false
		},
	}
}

//line emitter.gox:359
func (f *emitterContainerFragment) body(text string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line emitter.gox:360
			__e = __c.Set("id", "c2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line emitter.gox:360
			__e = __c.Any(text); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line emitter.gox:361
}

//line emitter.gox:363
func (f *emitterContainerFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line emitter.gox:365
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")

//line emitter.gox:368
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line emitter.gox:369
			__e = __c.Set("id", "c0"); if __e != nil { return }
//line emitter.gox:369
			__e = __c.Modify(doors.A(ctx, f.hit(1))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line emitter.gox:370
			__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line emitter.gox:370
					__e = __c.Set("id", "c1"); if __e != nil { return }
//line emitter.gox:370
					__e = __c.Modify(doors.A(ctx, &f.e)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line emitter.gox:371
					__e = __c.Any(f.body("first")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line emitter.gox:374
		__e = __c.Any(test.Button("emit", emitterCount(f.r, 0, f.e.Click(doors.PointerEmit{Button: 1})))); if __e != nil { return }
//line emitter.gox:375
		__e = __c.Any(test.Button("emit2", emitterCount(f.r, 0, f.e.Click(doors.PointerEmit{Button: 2})))); if __e != nil { return }
//line emitter.gox:376
		__e = __c.Any(test.Button("inner", func(ctx context.Context) bool {
		f.d.Inner(ctx, f.body("second"))
		return false
	})); if __e != nil { return }
	return })
//line emitter.gox:380
}
