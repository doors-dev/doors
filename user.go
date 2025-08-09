package doors

import (
	"context"
	"errors"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/path"
	"time"
)

// UserLocationReload triggers a browser location reload for the current instance.
// This executes location.reload() in JavaScript through the framework's call mechanism.
//
// The reload is performed asynchronously, causing the browser to reload the current
// page and create a new instance.
//
// Example:
//
//	func (h *handler) logout() doors.Attr {
//	    return doors.AClick{
//	        On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
//	            // Clear session data...
//	            doors.UserLocationReload(ctx)
//	            return true
//	        },
//	    }
//	}
func UserLocationReload(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocatinReload{})
}

// UserLocationAssignRaw navigates the browser to the specified URL by calling
// location.assign(url) in JavaScript. This creates a new entry in the browser's
// history stack, allowing the user to navigate back.
//
// The url parameter should be a complete URL string. The Origin field is set to
// false, indicating this is a raw URL navigation rather than model-based routing.
// Use this for external URLs or paths that don't correspond to path models.
//
// Example:
//
//	// Navigate to external site
//	doors.UserLocationAssignRaw(ctx, "https://example.com")
//
//	// Navigate to a specific path
//	doors.UserLocationAssignRaw(ctx, "/static/docs.pdf")
func UserLocationAssignRaw(ctx context.Context, url string) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationAssign{
		Href:   url,
		Origin: false,
	})
}

// UserLocationReplaceRaw replaces the current browser location with the specified
// URL by calling location.replace(url) in JavaScript. This does not create a new
// history entry, preventing the user from navigating back to the current page.
//
// The url parameter should be a complete URL string. Use this for redirects
// where you don't want the current page in the browser's history.
//
// Example:
//
//	// Redirect after form submission
//	doors.UserLocationReplaceRaw(ctx, "/success")
func UserLocationReplaceRaw(ctx context.Context, url string) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.Call(&instance.LocationReplace{
		Href:   url,
		Origin: false,
	})
}

// UserLocationReplace replaces the current browser location with a URL generated
// from the provided model. This calls location.replace(url) in JavaScript and
// does not create a new history entry.
//
// The model should be a struct with a registered path adapter. The adapter
// encodes the model into a URL path according to struct tags defined in the
// path model. The Origin field is set to true for model-based routing.
//
// Example:
//
//	type CatalogPath struct {
//	    IsCat  bool   `path:"/catalog/:CatId"`
//	    CatId  string
//	}
//
//	// Replace current location with category page
//	err := doors.UserLocationReplace(ctx, CatalogPath{
//	    IsCat: true,
//	    CatId: "electronics",
//	})
//
// Returns an error if the model cannot be encoded into a location.
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

// UserLocationAssign navigates the browser to a URL generated from the provided
// model by calling location.assign(url) in JavaScript. This creates a new history
// entry, allowing the user to navigate back.
//
// The model should be a struct with a registered path adapter. Navigation between
// different path model types triggers a full page reload and new instance creation,
// while navigation within the same model type updates the current instance reactively.
//
// Example:
//
//	// Navigate to item page
//	err := doors.UserLocationAssign(ctx, CatalogPath{
//	    IsItem: true,
//	    CatId:  "electronics",
//	    ItemId: 123,
//	})
//
// Returns an error if the model cannot be encoded into a location.
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

// UserSessionExpire sets the expiration duration for the current session.
// After the specified duration without activity, the session will be
// automatically terminated along with all its instances.
//
// Setting d to 0 disables automatic expiration. If the session has no active
// instances when expiration is disabled, it will be immediately terminated.
//
// This is commonly used to align the framework's session lifetime with your
// application's authentication session to ensure authorized pages don't
// outlive the authentication.
//
// Example:
//
//	// In login handler
//	const sessionDuration = 24 * time.Hour
//	session := createAuthSession(user, sessionDuration)
//	doors.UserSessionExpire(ctx, sessionDuration)
func UserSessionExpire(ctx context.Context, d time.Duration) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.SessionExpire(d)
}

// UserSessionEnd immediately terminates the current session and all its
// associated instances. This disconnects all active connections and
// cleans up server-side session resources.
//
// This must be called during logout to ensure no authorized pages remain
// active after the user logs out. Each session can contain multiple instances
// (browser tabs/windows), and this ensures they are all terminated.
//
// Example:
//
//	func (h *handler) logout() doors.Attr {
//	    return doors.AClick{
//	        On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
//	            // Clear auth cookie
//	            r.SetCookie(&http.Cookie{
//	                Name:   "session",
//	                MaxAge: -1,
//	            })
//	            // Terminate all instances
//	            doors.UserSessionEnd(ctx)
//	            return true
//	        },
//	    }
//	}
func UserSessionEnd(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.SessionEnd()
}

// UserInstanceEnd terminates the current instance while keeping the session
// and other instances active. This closes the connection for the specific
// browser tab or window.
//
// Use this when you need to close a specific page without affecting the
// entire session. The session continues with remaining instances until
// they are all closed or the session expires.
//
// Example:
//
//	// Close current tab after completion
//	doors.UserInstanceEnd(ctx)
func UserInstanceEnd(ctx context.Context) {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	inst.End()
}

// UserInstanceId returns the unique identifier for the current instance.
// Each instance represents a single browser tab or window connection.
//
// The ID is a cryptographically secure random string that uniquely identifies
// this instance within the application. Useful for logging, debugging, or
// tracking specific client connections.
//
// Example:
//
//	instanceId := doors.UserInstanceId(ctx)
//	log.Printf("Processing request for instance: %s", instanceId)
func UserInstanceId(ctx context.Context) string {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	return inst.Id()
}

// UserSessionId returns the unique identifier for the current session.
// A session represents a browser session and may contain multiple instances
// (tabs/windows) sharing the same session state.
//
// The ID is a cryptographically secure random string stored in a session cookie.
// All instances within the same browser share this session ID.
//
// Example:
//
//	sessionId := doors.UserSessionId(ctx)
//	analytics.Track("page_view", sessionId)
func UserSessionId(ctx context.Context) string {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	return inst.SessionId()
}

// UserSessionSave stores a key-value pair in the session-scoped storage.
// The storage persists for the session lifetime and is shared across
// all instances (browser tabs) within the same session.
//
// Both key and value can be of any type. The storage is thread-safe
// and can be accessed concurrently from different instances or goroutines.
//
// Returns true if the value was successfully saved, false otherwise.
//
// Example:
//
//	// Store user preferences
//	type Preferences struct {
//	    Theme    string
//	    Language string
//	}
//
//	saved := doors.UserSessionSave(ctx, "prefs", Preferences{
//	    Theme:    "dark",
//	    Language: "en",
//	})
func UserSessionSave(ctx context.Context, key any, value any) bool {
	return ctxstore.Save(ctx, common.SessionStoreCtxKey, key, value)
}

// UserSessionLoad retrieves a value from the session-scoped storage by its key.
// Returns nil if the key doesn't exist in the storage.
//
// The returned value must be type-asserted to its original type.
// The storage is shared across all instances within the same session.
//
// Example:
//
//	// Load user preferences
//	if val := doors.UserSessionLoad(ctx, "prefs"); val != nil {
//	    prefs := val.(Preferences)
//	    applyTheme(prefs.Theme)
//	}
func UserSessionLoad(ctx context.Context, key any) any {
	return ctxstore.Load(ctx, common.SessionStoreCtxKey, key)
}

// UserInstanceSave stores a key-value pair in the instance-scoped storage.
// The storage persists for the instance lifetime and is isolated to
// the current instance (browser tab / page). Each instance has its own separate storage.
//
// Both key and value can be of any type. The storage is thread-safe
// and can be accessed concurrently from different goroutines within the same instance.
//
// Returns true if the value was successfully saved, false otherwise.
//
// Example:
//
//	// Store user preferences for this specific tab
//	type Preferences struct {
//	    Theme    string
//	    Language string
//	}
//
//	saved := doors.UserInstanceSave(ctx, "prefs", Preferences{
//	    Theme:    "dark",
//	    Language: "en",
//	})
func UserInstanceSave(ctx context.Context, key any, value any) bool {
	return ctxstore.Save(ctx, common.InstanceStoreCtxKey, key, value)
}

// UserInstanceLoad retrieves a value from the instance-scoped storage by its key.
// Returns nil if the key doesn't exist in the storage.
//
// The returned value must be type-asserted to its original type.
// The storage is isolated to the current instance and not shared with other instances.
//
// Example:
//
//	// Load user preferences for this specific tab
//	if val := doors.UserInstanceLoad(ctx, "prefs"); val != nil {
//	    prefs := val.(Preferences)
//	    applyTheme(prefs.Theme)
//	}
func UserInstanceLoad(ctx context.Context, key any) any {
	return ctxstore.Load(ctx, common.InstanceStoreCtxKey, key)
}

// Location represents a URL location within the application's routing system.
// It encapsulates the path and query parameters encoded from a path model.
//
// Location provides methods for URL manipulation and can be used with
// navigation functions or href attributes.
type Location = path.Location

// NewLocation creates a Location from a model using the registered path adapter
// for that model's type. The adapter encodes the model's fields into a URL
// according to the path patterns defined in struct tags.
//
// Path models use struct tags to define routing patterns:
//   - `path:"/pattern"` tags on bool fields define path variants
//   - `:FieldName` in patterns captures path segments to struct fields
//   - `query:"name"` tags capture query parameters
//   - `:Field+` captures remaining path segments
//
// Example:
//
//	type ProductPath struct {
//	    List bool `path:"/products"`
//	    Item bool `path:"/products/:Id"`
//	    Id   int
//	    Sort string `query:"sort"`
//	}
//
//	// Create location for product item
//	loc, err := doors.NewLocation(ctx, ProductPath{
//	    Item: true,
//	    Id:   123,
//	    Sort: "price",
//	})
//	// loc.String() returns "/products/123?sort=price"
//
// Returns an error if no adapter is registered for the model's type or if
// encoding fails due to invalid model data.
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

// RandId generates a cryptographically secure random identifier string.
// The generated ID is URL-safe and suitable for use as session IDs,
// instance IDs, tokens, or any other unique identifiers.
//
// The ID format provides sufficient entropy for uniqueness in distributed
// systems and is safe for use in URLs without encoding.
//
// Example:
//
//	// Generate unique token
//	token := doors.RandId()
//
//	// Use for session creation
//	session := Session{
//	    ID:      doors.RandId(),
//	    UserID:  user.ID,
//	    Created: time.Now(),
//	}
func RandId() string {
	return common.RandId()
}
