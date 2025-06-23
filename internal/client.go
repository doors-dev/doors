package internal

import (
	"embed"
	"io/fs"
)

//go:embed client/*
var clientSrc embed.FS


var ClientSrc fs.FS
var ClientStyles []byte

func init() {
    ClientSrc, _ = fs.Sub(clientSrc, "client")
    ClientStyles, _ = fs.ReadFile(ClientSrc, "style.css")
}
