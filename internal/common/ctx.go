package common

type instanceCtxKey struct{}
type nodeCtxKey struct{}
type threadCtxKey struct{}
type hookThreadCtxKey struct{}
type renderMapCtxKey struct{}
type cinemaCtxKey struct{}
type blockingCtxKey struct{}
type adaptersCtxKey struct{}

var InstanceCtxKey = instanceCtxKey{}
var NodeCtxKey = nodeCtxKey{}
var CinemaCtxKey = cinemaCtxKey{}
var ThreadCtxKey = threadCtxKey{}
var HookThreadCtxKey = hookThreadCtxKey{}
var RenderMapCtxKey = renderMapCtxKey{}
var AdaptersCtxKey = adaptersCtxKey{}
