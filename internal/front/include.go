// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
	inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
	door := ctx.Value(common.CtxKeyDoor).(door.Core)
	_, already := ctxstore.Swap(ctx, common.CtxKeyInstanceStore, included{}, included{}).(included)
	if already {
		slog.Warn("doors header included multiple times on the page, keeping first", slog.String("instance_id", inst.Id()))
		return nil
	}
	style := inst.ImportRegistry().MainStyle()
	script := inst.ImportRegistry().MainScript()
	_, err := fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"/%s.d00r.css\"/>", style.HashString())
	if err != nil {
		return err
	}
	conf := inst.Conf()
	attrs := map[string]any{
		"src":             "/" + script.HashString() + ".d00r.js",
		"id":              inst.Id(),
		"data-root":       door.Id(),
		"data-ttl":        conf.InstanceTTL.Milliseconds(),
		"data-disconnect": conf.DisconnectHiddenTimer.Milliseconds(),
		"data-request":    conf.RequestTimeout.Milliseconds(),
		"data-ping":       conf.SolitairePing.Milliseconds(),
		"data-detached":   inst.IsDetached(),
	}
	lic := inst.License()
	if lic != nil {
		attrs["data-license"] = fmt.Sprintf("%s:%s:%s", lic.GetId(), lic.GetTier().String(), lic.GetDomain())
	}
	_, err = fmt.Fprint(w, "<script")
	if err != nil {
		return err
	}
	err = templ.RenderAttributes(context.Background(), w, templ.Attributes(attrs))
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, "></script>")
	if err != nil {
		return err
	}
	rm := ctx.Value(common.CtxKeyRenderMap).(*common.RenderMap)
	return rm.WriteImportMap(w)

}

var Include = include{}
