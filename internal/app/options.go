package app

import (
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/evanw/esbuild/pkg/api"
)

// SessionTracker is notified when sessions are created and removed.
type SessionTracker interface {
	// Create is called when a session is created. The request is the one
	// that triggered creation; it must not be retained beyond the call,
	// and its body must not be read.
	Create(id string, r *http.Request)

	// Delete is called when a session is removed.
	Delete(id string)
}

type Options struct {
	Conf           common.Conf
	CSP            *common.CSP
	ESBuild        func(profile string) api.BuildOptions
	SessionTracker SessionTracker
	ID             string
	ErrorPage      ErrorPage
}

type notracker struct{}

func (n notracker) Create(id string, r *http.Request) {
}

func (n notracker) Delete(id string) {
}

var _ SessionTracker = notracker{}

func (o *Options) initDefaults() {
	common.InitDefaults(&o.Conf)
	if o.ESBuild == nil {
		o.ESBuild = resources.DefaultProfile
	}
	if o.ID == "" {
		o.ID = "doors"
	}
	if o.SessionTracker == nil {
		o.SessionTracker = notracker{}
	}
}
