// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/instance"
)

type include struct{}

type included struct{}

func (_ include) Render(ctx context.Context, w io.Writer) error {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	door := ctx.Value(common.DoorCtxKey).(door.Core)
	_, already := ctxstore.Swap(ctx, common.InstanceStoreCtxKey, included{}, included{}).(included)
	if already {
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
	attrs := map[string]any{
		"src":          "/" + script.HashString() + ".js",
		"id":           inst.Id(),
		"data-root":    door.Id(),
		"data-ttl":     conf.TTL.Milliseconds(),
		"data-sleep":   conf.SleepTimeout.Milliseconds(),
		"data-request": conf.RequestTimeout.Milliseconds(),
		"data-detached":     inst.Detached(),
	}
	_, err := w.Write([]byte("<script"))
	if err != nil {
		return err
	}
	err = templ.RenderAttributes(context.Background(), w, templ.Attributes(attrs))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("></script>"))
	return err

}

var Include = include{}
