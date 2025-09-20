// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package shredder

import (
	"testing"
	"time"
)

func testCheckTrue(ch chan bool, t *testing.T) {
	i := 0
	for val := range ch {
		i += 1
		if val != true {

			t.Error("True check failed for ", i, "value")
		}
	}
}
func testCheckOrder(ch chan int, t *testing.T) {
	prev := 0
	for val := range ch {
		//println("Received ", val)
		expected := prev + 1
		if val != expected {
			t.Error("Order broken got ", val, " after ", prev)
		}
		prev = val
	}
}

func testWait(ms int) {
	<-time.NewTimer(time.Millisecond * time.Duration(ms)).C
}

func testThread(limit int) *Thread {
	s := testSpawner(limit)
	return s.NewThead()
}

type dummy struct{}

func (d dummy) OnPanic(err error) {
	panic(err)
}

func testSpawner(limit int) *Spawner {
	p := NewPool(limit)
	return p.Spawner(dummy{})
}

func TestWrite(t *testing.T) {
	th := testThread(10)
	ch := make(chan int, 10)
	Run(func(t *Thread) {
		testWait(300)
		ch <- 1
	}, W(th))
	Run(func(t *Thread) {
		testWait(200)
		ch <- 2
	}, W(th))
	Run(func(t *Thread) {
		ch <- 3
	}, W(th))
	Run(func(t *Thread) {
		close(ch)
	}, W(th))
	testCheckOrder(ch, t)
}

func TestRead(t *testing.T) {
	th := testThread(3)
	ch := make(chan int, 10)
	Run(func(t *Thread) {
		testWait(100)
		ch <- 3
	}, R(th))
	Run(func(t *Thread) {
		testWait(50)
		ch <- 2
	}, R(th))
	Run(func(t *Thread) {
		ch <- 1
	}, R(th))
	Run(func(t *Thread) {
		testWait(50)
		ch <- 4
	}, W(th))
	Run(func(t *Thread) {
		testWait(50)
		ch <- 6
	}, R(th))
	Run(func(t *Thread) {
		ch <- 5
	}, R(th))
	Run(func(t *Thread) {
		close(ch)
	}, W(th))
	testCheckOrder(ch, t)
}

func TestInternal(t *testing.T) {
	th := testThread(10)
	ch := make(chan int, 10)
	Run(func(t *Thread) {
		testWait(200)
		Run(func(t *Thread) {
			Run(func(t *Thread) {
				testWait(50)
				ch <- 2
			}, R(t))
			Run(func(t *Thread) {
				ch <- 1
			}, R(t))
			Run(func(t *Thread) {
				ch <- 3
			}, W(t))
		}, W(t))
		Run(func(t *Thread) {
			ch <- 4
		}, W(t))
	}, W(th))
	Run(func(t *Thread) {
		ch <- 5
	}, R(th))
	Run(func(t *Thread) {
		ch <- 6
	}, W(th))
	Run(func(t *Thread) {
		close(ch)
	}, W(th))
	testCheckOrder(ch, t)
}
func TestStarving(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	Run(func(t *Thread) {
		ch <- 1
		testWait(100)
	}, R(t1))
	Run(func(t *Thread) {
		ch <- 3
	}, Ws(t1))
	Run(func(t *Thread) {
		testWait(50)
		ch <- 2
	}, R(t1))
	Run(func(t *Thread) {
		ch <- 4
	}, Ws(t1))
	Run(func(t *Thread) {
		ch <- 5
	}, W(t1))
	Run(func(t *Thread) {
		ch <- 6
	}, R(t1))
	Run(func(t *Thread) {
		ch <- 7
		close(ch)
	}, W(t1))
	testCheckOrder(ch, t)
}
func TestAfterStarve(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()

	Run(func(t *Thread) {
		testWait(100)
		ch <- 1
	}, R(t1))
	Run(func(t *Thread) {
		testWait(200)
		ch <- 2
	}, Ws(t1))
	testWait(150)
	Run(func(t *Thread) {
		ch <- 3
	}, R(t1))
	testWait(300)
	close(ch)
	testCheckOrder(ch, t)
}

func TestJoin(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t3 := s.NewThead()
	Run(func(t *Thread) {
		testWait(50)
		ch <- 2
	}, W(t3))
	Run(func(t *Thread) {
		testWait(200)
		ch <- 6
	}, R(t3))
	Run(func(t *Thread) {
		testWait(100)
		ch <- 3
	}, W(t1))
	Run(func(t *Thread) {
		ch <- 1
	}, R(t2))
	Run(func(t *Thread) {
		ch <- 4
		Run(func(t *Thread) {
			ch <- 5
		}, W(t), R(t3))
	}, W(t2), W(t1))
	Run(func(t *Thread) {
		close(ch)
	}, W(t2), W(t3))
	testCheckOrder(ch, t)
}

func TestJoinInstant(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t3 := s.NewThead()
	Run(func(t *Thread) {
		testWait(50)
		ch <- 2
	}, W(t3))
	Run(func(t *Thread) {
		testWait(200)
		ch <- 6
	}, R(t3))
	Run(func(t *Thread) {
		testWait(100)
		ch <- 3
	}, W(t1))
	Run(func(t *Thread) {
		ch <- 1
	}, R(t2))
	Run(func(t *Thread) {
		ch <- 4
		Run(func(t *Thread) {
			ch <- 5
		}, W(t), Ri(t3))
	}, W(t2), Wi(t1))
	Run(func(t *Thread) {
		close(ch)
	}, W(t2), W(t3))
	testCheckOrder(ch, t)
}

func TestKill1(t *testing.T) {
	ch := make(chan bool, 10)
	order := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	Run(func(*Thread) {
		testWait(100)
	}, W(t2))
	Run(func(*Thread) {
		testWait(50)
		t1.Kill(func() {
			order <- 3
		})
		testWait(20)
		order <- 1
	}, R(t1))
	Run(func(t *Thread) {
		Run(func(t *Thread) {
			ch <- t == nil
		}, R(t), W(t2))
		Run(func(t *Thread) {
			testWait(150)
			order <- 2
			Run(func(t *Thread) {
				ch <- t == nil
				close(ch)
			}, W(t))
		}, R(t))
	}, R(t1))
	testCheckTrue(ch, t)
	testWait(200)
	close(order)
	testCheckOrder(order, t)
}
func TestKill2(t *testing.T) {
	ch := make(chan bool, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t3 := s.NewThead()
	t2.Kill(nil)
	Run(func(t *Thread) {
		ch <- t == nil
		close(ch)
	}, W(t1), W(t2), W(t3))
	testCheckTrue(ch, t)
}

func TestKill3(t *testing.T) {
	ch := make(chan bool, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	Run(func(*Thread) {
		testWait(50)
		t2.Kill(nil)
	}, W(t2))
	t3 := s.NewThead()
	Run(func(t *Thread) {
		ch <- t == nil
		close(ch)
	}, W(t1), W(t2), W(t3))
	testCheckTrue(ch, t)
}
