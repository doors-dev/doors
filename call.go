package doors

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
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
//              The handler receives a CallRequest, which can read data from the frontend.
//   - Cancel: optional. Called if the call is invalidated before it reaches the frontend,
//             or if manually canceled using the returned TryCancel function.
type CallConf struct {
    // Name of the JavaScript call handler handler  (must be registered on the frontend).
	Name string

	// Arg is the value passed to the frontend function. It is serialized as JSON.
	Arg any

	// Trigger is an optional backend handler that is called when the frontend responds.
	// Use this to handle data returned by the frontend, such as result values or side effects.
	Trigger func(context.Context, CallRequest)

	// Cancel is called if the context becomes invalid before the call is delivered,
	// or if the call is canceled explicitly.
	Cancel func(context.Context, error)
}

// TryCancel is a function that attempts to cancel a pending frontend call.
//
// It returns true if the call was still pending and has now been canceled.
// Returns false if the call was already delivered or canceled automatically.
type TryCancel func() bool

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
//   - TryCancel: a function to cancel the pending call (usually ignored).
//   - ok: false if the call couldn't be registered or marshaling failed.
//
// Example:
//
//     // Go (backend):
//     d.Call(ctx, d.CallConf{
//         Name: "alert",
//         Arg:  "Hello",
//     })
//
//     // JavaScript (frontend):
//     <script>
//     $D.on(document.currentScript, "alert", (message) => alert(message))
//     </script>
//
// Notes:
//   - The Name must match the identifier registered via `$D.on(...)` in frontend JavaScript.
//   - The Arg is serialized to JSON and passed as the first argument to the JS handler.
//   - Use `document` instead of `document.currentScript` to register globally.
//   - To handle a frontend response, set the Trigger field in CallConf
func Call(ctx context.Context, conf CallConf) (TryCancel, bool) {
	n := ctx.Value(common.NodeCtxKey).(node.Core)
	call := &node.CallHook{
		Trigger: func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if conf.Trigger != nil {
				conf.Trigger(ctx, &request{w: w, r: r})
			}
		},
		Cancel: conf.Cancel,
	}
	arg, err := json.Marshal(conf.Arg)
	if err != nil {
		slog.Error("Call arg marshaling error", slog.String("call_name", conf.Name), slog.String("json_error", err.Error()))
		return nil, false
	}
	cancel, ok := n.RegisterCallHook(ctx, conf.Name, (common.JsonWritabeRaw)(arg), call)
	return cancel, ok
}
