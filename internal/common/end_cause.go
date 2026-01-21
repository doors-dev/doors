package common

import "fmt"

type EndCause int

const (
	EndCauseKilled EndCause = iota
	EndCauseSuspend
	EndCauseSyncError
)

func (c EndCause) Error() string {
	return fmt.Sprint("cause: ", int(c))
}

