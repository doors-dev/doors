# Live Reloading

The easiest way to enable live reloading is by using [https://github.com/bokwoon95/wgo](https://github.com/bokwoon95/wgo) tool. 

### Install

```bash
go install github.com/bokwoon95/wgo@latest
```

### Run

Run in your project dir:

```bash
 wgo -file=.go -file=.templ  -xfile=_templ.go templ generate :: go run .
```

> Sometimes, the reload occurs at the wrong moment, and the browser displays the "Unable to connect" page. In that case, manually reload the page.
>
> Additionally, the **templ** included watcher struggles with my lib as a dependency.
>
> I will research ways to improve this experience later.

### Check

Open https://localhost:8443/ , then edit `<h1>`  content in your `./home/page.templ` ,  page should be automatically reloaded with the new version.


