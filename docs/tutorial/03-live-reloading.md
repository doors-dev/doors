# Live Reloading

The easiest way to enable live reloading is by using [https://github.com/bokwoon95/wgo](https://github.com/bokwoon95/wgo) tool. 

## 1. Install

```bash
go install github.com/bokwoon95/wgo@latest
```

## 2. Run

Run in your project dir:

```bash
 wgo -file=.go -file=.templ  -xfile=_templ.go templ generate :: go run .
```

> Sometimes, the reload happens at the wrong moment, causing the browser to show the "Unable to connect" page. In that situation, manually reload the page.
>
> Additionally, the **templ** included watcher struggles with my lib as a dependency.
>
> I will research ways to improve this experience later.

## 3. Check

Open https://localhost:8443/, then edit the `<h1>` content in your `./home/page.templ`; the page should automatically reload with the new version.

