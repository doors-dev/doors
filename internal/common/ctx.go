package common

type instanceCtxKey struct{}
type nodeCtxKey struct{}
type threadCtxKey struct{}
type renderMapCtxKey struct{}
type blockingCtxKey struct{}
type adaptersCtxKey struct{}
type sessionStoreCtxKey struct{}
type instanceStoreCtxKey struct{}
type parentCtxKey struct{}

var InstanceCtxKey = instanceCtxKey{}
var NodeCtxKey = nodeCtxKey{}
var ThreadCtxKey = threadCtxKey{}
var RenderMapCtxKey = renderMapCtxKey{}
var AdaptersCtxKey = adaptersCtxKey{}
var SessionStoreCtxKey = sessionStoreCtxKey{}
var InstanceStoreCtxKey = instanceStoreCtxKey{}
var ParentCtxKey = parentCtxKey{}
