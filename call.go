package doors

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/door"
)

// CallConf configures a backend-to-frontend JavaScript call.
//
// This allows server-side code to invoke a JavaScript function in the browser,
// passing arguments and optionally handling a response via a trigger hook.
//
// The call is sent to the frontend immediately once the connection is ready (typically within a few milliseconds).
// If the backend context is no longer valid (e.g., the component was unmounted), the call is automatically canceled.
//
// Fields:
//   - Name: the name of the frontend JavaScript function to call.
//   - Arg: the argument to pass to the function (must be JSON-serializable).
//   - Trigger: optional. Called when the frontend responds to the function call.
//     The handler receives a CallRequest, which can read data from the frontend.
//   - Cancel: optional. Called if the call is invalidated before it reaches the frontend,
//     or if manually canceled using the returned TryCancel function.
type CallConf struct {
	// Name of the JavaScript call handler  (must be registered on the frontend).
	Name string

	// Arg is the value passed to the frontend function. It is serialized to JSON.
	Arg any

	// On is an optional backend handler that is called with the frontend call responce.
	On func(context.Context, RCall)

	// OnCancel is called if the context becomes invalid before the call is delivered,
	// or if the call is canceled explicitly. Optional.
	OnCancel func(context.Context, error)
}

func (conf *CallConf) clientCall() (*door.ClientCall, bool) {
	arg, err := common.MarshalJSON(conf.Arg)
	if err != nil {
		slog.Error("Call arg marshaling error", slog.String("call_name", conf.Name), slog.String("json_error", err.Error()))
		return nil, false
	}
	return &door.ClientCall{
		Name:    conf.Name,
		Arg:     json.RawMessage(arg),
		Trigger: conf.triggerFunc(),
		Cancel:  conf.OnCancel,
	}, true
}

func (c *CallConf) triggerFunc() func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if c.On == nil {
		return nil
	}
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		c.On(ctx, &request{w: w, r: r, ctx: ctx})
	}
}

// Call sends a backend-initiated JavaScript function call to the frontend.
//
// The call is dispatched as soon as the frontend connection is ready.
// It serializes the Arg field of CallConf to JSON and sends it to a frontend function
// registered with the given Name. Optionally, the backend can handle a response
// via the Trigger field, or react to cancellation via the Cancel field.
//
// If the component or context is no longer valid (e.g., unmounted), Cancel is called automatically.
//
// Parameters:
//   - ctx: the current rendering or lifecycle context.
//   - conf: the configuration specifying the function name, argument, and optional handlers.
//
// Returns:
//   - a function to cancel the pending call (usually ignored).
//   - ok: false if the call couldn't be registered or marshaling failed.
//
// Example:
//
//		// Go (backend):
//		d.Call(ctx, d.CallConf{
//		    Name: "alert",
//		    Arg:  "Hello",
//		})
//
//		// JavaScript (frontend):
//	 $doors.Script() {
//			<script>
//				$d.on("alert", (message) => alert(message))
//			</script>
//	}
//
// Notes:
//   - The Name must match the identifier registered via `$D.on(...)` in frontend JavaScript.
//   - The Arg is serialized to JSON and passed as the first argument to the JS handler.
//   - Use `document` instead of `document.currentScript` to register globally.
//   - To handle a frontend response, set the Trigger field in CallConf
func Call(ctx context.Context, conf CallConf) (func(), bool) {
	n := ctx.Value(common.DoorCtxKey).(door.Core)
	call, ok := conf.clientCall()
	if !ok {
		return nil, false
	}
	cancel, ok := n.RegisterClientCall(ctx, call)
	return cancel, ok
}
