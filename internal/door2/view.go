package door2

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

func (v *view) headFrame(ctx context.Context, doorID uint64, headID uint64) (*gox.JobHeadOpen, *gox.JobHeadClose) {
	if v.tag == "" {
		attrs := gox.NewAttrs(ctx)
		attrs.Get("id").Set(fmt.Sprintf("d00r/%d", doorID))
		return &gox.JobHeadOpen{
				Kind:  gox.KindRegular,
				Tag:   "d0-0r",
				Attrs: attrs,
				ID:    headID,
			}, &gox.JobHeadClose{
				Kind: gox.KindRegular,
				Tag:  "d0-0r",
				ID:   headID,
			}
	}
	attrs := v.attrs.Clone()
	attrs.Get("data-d00r").Set(fmt.Sprintf("%d", doorID))
	return &gox.JobHeadOpen{
			Kind:  gox.KindRegular,
			Tag:   v.tag,
			Attrs: attrs,
			ID:    headID,
		}, &gox.JobHeadClose{
			Kind: gox.KindRegular,
			Tag:  v.tag,
			ID:   headID,
		}
}

func (v *view) renderContent(cur gox.Cursor) error {
	if comp, ok := v.content.(gox.Comp); ok {
		return comp.Main()(cur)
	}
	return cur.Any(v.content)
}

func (v *view) inherit(prev *view) {
	v.tag = prev.tag
	v.attrs = prev.attrs
}
