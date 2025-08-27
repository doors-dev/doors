# After Actions

In any event handler, you can schedule a browser action to be executed after the browser receives a response.

* `r.After(doors.AfterLocationReload())` 

  Reload page.

* `r.After(doors.AfterLocationAssign(model any))`

  Assign a new location from the provided **path model** (creates a history entry)

* `r.After(doors.AfterLocationReload(model any))`

  Replace location with the provided **path model** (without history entry)

* `r.After(doors.AfterScrollInto(selector string, smooth bool))`

  Scrolls the page into the element queried by the selector.

