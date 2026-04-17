package shredder

import "context"

type FreeFrame struct{}

func (f FreeFrame) Release() {

}

func (f FreeFrame) Run(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(run{runtime: r, ctx: ctx, fun: fun})
}

func (f FreeFrame) Submit(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: r, ctx: ctx, fun: fun})
}

func (f FreeFrame) schedule(e executable) {
	e.execute(func(error) {})
}

var _ Frame = FreeFrame{}
