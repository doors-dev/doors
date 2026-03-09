package printer

import (
	"io"

	"github.com/doors-dev/gox"
)

type defaultPrinter struct {
	w io.Writer
}

func (d defaultPrinter) Send(job gox.Job) error {
	if job.Context().Err() != nil {
		return nil
	}
	return job.Output(d.w)
}
