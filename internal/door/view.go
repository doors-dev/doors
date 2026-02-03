package door

import (
	"context"
	"fmt"
	"github.com/doors-dev/gox"
)

type view struct {
	tag     string
	attrs   gox.Attrs
	content any
	elem    gox.Elem
}

func (v *view) headFrame(parentCtx context.Context, doorID uint64, headID uint64) (*gox.JobHeadOpen, *gox.JobHeadClose) {
	if v.tag == "" {
		attrs := gox.NewAttrs()
		attrs.Get("id").Set(fmt.Sprintf("d00r/%d", doorID))
		return gox.NewJobHeadOpen(
				headID,
				gox.KindRegular,
				"d0-r",
				parentCtx,
				attrs,
			), gox.NewJobHeadClose(
				headID,
				gox.KindRegular,
				"d0-r",
				parentCtx,
			)
	}
	attrs := v.attrs.Clone()
	attrs.Get("data-d0r").Set(fmt.Sprintf("%d", doorID))
	return gox.NewJobHeadOpen(
			headID,
			gox.KindRegular,
			v.tag,
			parentCtx,
			attrs,
		), gox.NewJobHeadClose(
			headID,
			gox.KindRegular,
			v.tag,
			parentCtx,
		)
}

func (v *view) renderContent(cur gox.Cursor) error {
	if comp, ok := v.content.(gox.Comp); ok {
		return comp.Main()(cur)
	}
	return cur.Any(v.content)
}
