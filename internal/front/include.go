package front

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
)

type include struct{}

func (_ include) Render(ctx context.Context, w io.Writer) error {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	node := ctx.Value(common.NodeCtxKey).(node.Core)
	if !inst.Include() {
		slog.Warn("doors header included multiple times on the page, keeping first", slog.String("instance_id", inst.Id()))
		return nil
	}
	style := inst.ImportRegistry().MainStyle()
	script := inst.ImportRegistry().MainScript()
	_, inline := inst.InlineNonce()
	if !inline {
		_, err := w.Write(fmt.Appendf(nil, "<link rel=\"stylesheet\" href=\"/%s.css\"/>", style.HashString()))
		if err != nil {
			return err
		}
	} else {
		_, err := w.Write([]byte("<style>"))
		if err != nil {
			return err
		}
		_, err = w.Write(style.Content())
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("</style>"))
		if err != nil {
			return err
		}
	}
	conf := inst.ClientConf()
	_, err := w.Write(fmt.Appendf(nil,
		"<script src=\"/%s.js\" id=\"%s\" data-root=\"%d\" data-ttl=\"%d\" data-sleep=\"%d\" data-request=\"%d\"></script>",
		script.HashString(),
		inst.Id(),
		node.Id(),
		conf.TTL.Milliseconds(),
		conf.SleepTimeout.Milliseconds(),
		conf.RequestTimeout.Milliseconds(),
	))
	return err

}

var Include = include{}
