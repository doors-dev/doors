package instance2

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/door2"
	"github.com/doors-dev/gox"
)

type renderState int

const (
	lookingForHead renderState = iota
	insideHead
	headDone
)

func (inst *Instance[M]) render(w io.Writer, r *http.Request, js door2.JobStream) error {
	var state renderState
	var headId uint64
	for {
		job, ok := js.Next()
		if !ok {
			break
		}
		switch state {
		case lookingForHead:
			job, ok := job.(*gox.JobHeadOpen)
			switch true {
			case !ok:
			case strings.EqualFold(job.Tag, "head"):
				headId = job.Id
				state = insideHead
			case strings.EqualFold(job.Tag, "body"):
				if err := inst.renderResources(w, true); err != nil {
					return err
				}
				state = headDone
			}
		case insideHead:
			openJob, openOk := job.(*gox.JobHeadOpen)
			closeJob, closeOk := job.(*gox.JobHeadClose)
			if (openOk && strings.EqualFold(openJob.Tag, "script")) || (closeOk && closeJob.Id == headId) {
				if err := inst.renderResources(w, true); err != nil {
					return err
				}
				state = headDone
			}
		}
		if err := job.Output(w); err != nil {
			return err
		}
	}
	return nil
}

func (inst *Instance[M]) renderResources(w io.Writer, addHeadTags bool) error {
	return nil
}
