package renderer

/*
import (
	"io"
	"sync"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"golang.org/x/net/context"
)

type scope struct {
	jobs []gox.Job
	done chan struct{}
}



func (s *scope) send(j gox.Job) {
	s.jobs = append(s.jobs, j)
}

type Proxy struct {
	p      gox.Printer
	thread *shredder.Thread
	scopes []scope
	mu     sync.Mutex
	ctx    context.Context
}

func (x *Proxy) Output(w io.Writer) error {
	for {

	}
}


func (x *Proxy) Init(p gox.Printer) {
	x.p = p
}

func (x *Proxy) Send(j gox.Job) (done bool, err error) {
	switch job := j.(type) {
	case *gox.JobElem:
		go func() {

		}()
	}
}

func (x *Proxy) Terminate() {
	panic("unimplemented")
}

var _ gox.Proxy = &Proxy{} */
