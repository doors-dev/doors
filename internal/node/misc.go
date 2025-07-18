package node

var closedCh = make(chan error)

func init() {
	close(closedCh)
}
