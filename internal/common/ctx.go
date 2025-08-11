package common

type ctxKey int

const (
    InstanceCtxKey ctxKey = iota
    NodeCtxKey
    ThreadCtxKey
    RenderMapCtxKey
    BlockingCtxKey
    AdaptersCtxKey
    SessionStoreCtxKey
    InstanceStoreCtxKey
    ParentCtxKey
)

