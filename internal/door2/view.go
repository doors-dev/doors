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

func (v *view) headFrame(ctx context.Context, doorId uint64, headId uint64) (*gox.JobHeadOpen, *gox.JobHeadClose) {
	if v.tag == "" {
		attrs := gox.NewAttrs(ctx)
		attrs.Get("id").Set(fmt.Sprintf("d00r/%d", doorId))
		return &gox.JobHeadOpen{
			Kind: gox.KindRegular,
			Tag:  "d0-0r",
			Attrs: attrs,
			Id: headId,
		}, &gox.JobHeadClose{
			Kind: gox.KindRegular,
			Tag:  "d0-0r",
			Id: headId,
		}
	} 
	attrs := v.attrs.Clone()
	attrs.Get("data-d00r").Set(fmt.Sprintf("%d", doorId))
	return &gox.JobHeadOpen{
		Kind: gox.KindRegular,
		Tag:  v.tag,
		Attrs: attrs,
		Id: headId,
	}, &gox.JobHeadClose{
		Kind: gox.KindRegular,
		Tag:  v.tag,
		Id: headId,
	}
}


func (v *view) inherit(prev *view) {
	v.tag = prev.tag
	v.attrs = prev.attrs
}


