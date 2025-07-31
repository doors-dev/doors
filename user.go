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

func UserLocationReload(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocatinReload{})
}

func UserLocationAssignRaw(ctx context.Context, url string) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationAssign{
		Href:   url,
		Origin: false,
	})

}
func UserLocationReplaceRaw(ctx context.Context, url string) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationReplace{
		Href:   url,
		Origin: false,
	})
}

func UserLocationReplace(ctx context.Context, model any) error {
	l, err := NewLocation(ctx, model)
	if err != nil {
		return err
	}

	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationReplace{
		Href:   l.String(),
		Origin: true,
	})
	return nil
}

func UserLocationAssign(ctx context.Context, model any) error {
	l, err := NewLocation(ctx, model)
	if err != nil {
		return err
	}

	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationAssign{
		Href:   l.String(),
		Origin: true,
	})
	return nil
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

func RandId() string {
	return common.RandId()
}
