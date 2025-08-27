# Error Handling

The **On Error API** defines actions that execute when an **error occurs during a hook’s processing**. It supports UI indication and custom client-side callbacks. The API reuses the Indicator model used for temporary UI changes and the call handlers registration system from the **JavaScript** article.

## Concept

* `OnError` attribute field accepts a slice of action (to perform multiple actions on error if needed) 
* **You can create a simple indication with helper functions, such as `doors.OnErrorIndicate(...)`** or or take full control with `[]doors.OnError{/* OnError actions */}`
* Error kinds (not all factual hook errors trigger `OnError` actions):
  * Canceled by scope - no trigger
  * Unauthorized -  no trigger, cause page reload (401 and 410 codes)
  * Not Found  - not trigger (hook expired, 404 code) 
  * Bad Request - trigger, but not expected to occur (400 code)
  * Other - other 4XX code (not used by framework, if you issued)
  * **Network Error - trigger**
  * **Server Error - trigger (5XX code)**
  * Capture Error - trigger, but not expected to occur (means js side event handling issue)
  
* Custom hook error (`$d.hook` and `$d.hookRaw`) handling needs to be done manually; any error throws an exception

## Indicate 

Apply the indication for the specified duration when error occurs

```templ
type IndicateOnError struct {
	Duration  time.Duration
	Indicator []Indicator
}
```

**Helper**: 

```templ
func OnErrorIndicate(duration time.Duration, indicator []Indicator) []OnError 
```

## Call

Call the registered JavaScript handler when an error occurs

```templ
type CallOnError struct {
  // name of handler
	Name string
	// value of the meta field of $D.CaptureErr instance, marshaled to JSON
	// and parced on the front-end
	Meta any
}
```

**Helper**:

```templ
func OnErrorCall(name string, meta any) []OnError 
```

> Call **Registration/Invocation** rules are explained in the [JavaScript](../10-javascript.md) article

## Example

```templ
@doors.Script() {
    <script>
        $d.on("err", (e) => {
        	const message = e.meta
        	alert(message)
        })
    </script>
}

@doors.AClick{
	OnError: doors.OnErrorCall("err", "Click error, please try later!")
	/* setup */
}
<button></button>
```



## Error Type `CaptureErr`

Errors produced by hooks are instances of `CaptureErr` class that inherits `Error`.

### Fields

* `status: number | undefined`
  *http status code, if the request failed*
* `case: Error | undefined`
  Original catched Error if present
* `kind:  "canceled" | "unauthorized" | "not_found" | "network" | "server" | "bad_request" | "capture"`
* `meta: any`
  contains meta information specified in `OnError` action in attribute if provided

### Methods

Help to determine error kind:

* `canceled() : bool`
* `notFound() : bool`
* `unauthorized() : bool`
* `network() : bool`
* `capture() : bool`
* `server() : bool`
* `badRequest() : bool`
* `isOther(): bool`

