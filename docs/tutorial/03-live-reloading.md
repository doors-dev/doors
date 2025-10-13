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

> Sometimes, the reload happens when your code is broken, causing the browser to show the "Unable to connect" page. In that situation, manually reload the page.
>

**From this point, live reloading is always enabled.**

Next: [Integrations](./04-integrations.md)
---
