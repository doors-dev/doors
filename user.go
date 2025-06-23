package doors

import (
	"context"
	"errors"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
)

func UserRelocate(ctx context.Context, model any) error {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	return inst.Relocate(ctx, model)
}

func UserSessionExpire(ctx context.Context, d time.Duration) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.SessionExpire(d)
}

func UserSessionEnd(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.SessionEnd()
}

func UserInstanceEnd(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.End()
}

func UserInstanceId(ctx context.Context) string {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	return inst.Id()
}

func UserSessionId(ctx context.Context) string {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	return inst.SessionId()
}

func UserSave(ctx context.Context, key any, value any) bool {
	return ctxstore.Save(ctx, key, value)
}

func UserLoad(ctx context.Context, key any) any {
	return ctxstore.Load(ctx, key)
}

type Location = path.Location

func NewLocation(ctx context.Context, model any) (Location, error) {
	adapters := ctx.Value(common.AdaptersCtxKey).(map[string]path.AnyAdapter)
	name := path.GetAdapterName(model)
	adapter, ok := adapters[name]
	if !ok {
		var l Location
		return l, errors.New("Adapter for " + name + " is not regestered")
	}
	location, err := adapter.EncodeAny(model)
	if err != nil {
		var l Location
		return l, err
	}
	return *location, nil
}
