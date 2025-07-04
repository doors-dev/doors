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

func testSpawner(limit int) *Spawner {
	p := NewPool(limit)
	return p.Spawner()
}

func TestWrite(t *testing.T) {
	th := testThread(10)
	ch := make(chan int, 10)
	th.Write(func(t *Thread) {
		testWait(300)
		ch <- 1
	})
	th.Write(func(t *Thread) {
		testWait(200)
		ch <- 2
	})
	th.Write(func(t *Thread) {
		ch <- 3
	})
	th.Write(func(t *Thread) {
		close(ch)
	})
	testCheckOrder(ch, t)
}

func TestRead(t *testing.T) {
	th := testThread(3)
	ch := make(chan int, 10)
	th.Read(func(t *Thread) {
		testWait(100)
		ch <- 3
	})
	th.Read(func(t *Thread) {
		testWait(50)
		ch <- 2
	})
	th.Read(func(t *Thread) {
		ch <- 1
	})
	th.Write(func(t *Thread) {
		testWait(50)
		ch <- 4
	})
	th.Read(func(t *Thread) {
		testWait(50)
		ch <- 6
	})
	th.Read(func(t *Thread) {
		ch <- 5
	})
	th.Write(func(t *Thread) {
		close(ch)
	})
	testCheckOrder(ch, t)
}

func TestInternal(t *testing.T) {
	th := testThread(10)
	ch := make(chan int, 10)
	th.Write(func(t *Thread) {
		testWait(200)
		t.Write(func(t *Thread) {
			t.Read(func(t *Thread) {
				testWait(50)
				ch <- 2
			})
			t.Read(func(t *Thread) {
				ch <- 1
			})
			t.Write(func(t *Thread) {
				ch <- 3
			})
		})
		t.Write(func(t *Thread) {
			ch <- 4
		})
	})
	th.Read(func(t *Thread) {
		ch <- 5
	})
	th.Write(func(t *Thread) {
		ch <- 6
	})
	th.Write(func(t *Thread) {
		close(ch)
	})
	testCheckOrder(ch, t)
}
func TestStarving(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t1.Read(func(t *Thread) {
		ch <- 1
		testWait(100)
	})
	t1.WriteStarving(func(t *Thread) {
		ch <- 3
	})
	t1.Read(func(t *Thread) {
		testWait(50)
		ch <- 2
	})
	t1.WriteStarving(func(t *Thread) {
		ch <- 4
	})
	t1.Write(func(t *Thread) {
		ch <- 5
	})
	t1.Read(func(t *Thread) {
		ch <- 6
	})
	t1.Write(func(t *Thread) {
		ch <- 7
		close(ch)
	})
	testCheckOrder(ch, t)
}
func TestAfterStarve(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()

	t1.Read(func(t *Thread) {
		testWait(100)
		ch <- 1
	})
	t1.WriteStarving(func(t *Thread) {
		testWait(200)
		ch <- 2
	})
	testWait(150)
	t1.Read(func(t *Thread) {
		ch <- 3
	})
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
	t3.Write(func(t *Thread) {
		testWait(50)
		ch <- 2
	})
	t3.Read(func(t *Thread) {
		testWait(200)
		ch <- 6
	})
	t1.Write(func(t *Thread) {
		testWait(100)
		ch <- 3
	})
	t2.Read(func(t *Thread) {
		ch <- 1
	})
	t2.Write(func(t *Thread) {
		ch <- 4
		t.Write(func(t *Thread) {
			ch <- 5
		}, R(t3))
	}, W(t1))
	t2.Write(func(t *Thread) {
		close(ch)
	}, W(t3))
	testCheckOrder(ch, t)
}

func TestJoinInstant(t *testing.T) {
	ch := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t3 := s.NewThead()
	t3.Write(func(t *Thread) {
		testWait(50)
		ch <- 2
	})
	t3.Read(func(t *Thread) {
		testWait(200)
		ch <- 6
	})
	t1.Write(func(t *Thread) {
		testWait(100)
		ch <- 3
	})
	t2.Read(func(t *Thread) {
		ch <- 1
	})
	t2.WriteInstant(func(t *Thread) {
		ch <- 4
		t.WriteInstant(func(t *Thread) {
			ch <- 5
		}, R(t3))
	}, W(t1))
	t2.Write(func(t *Thread) {
		close(ch)
	}, W(t3))
	testCheckOrder(ch, t)
}

func TestKill1(t *testing.T) {
	ch := make(chan bool, 10)
	order := make(chan int, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t2.Write(func(*Thread) {
		testWait(100)
	})
	t1.Read(func(*Thread) {
		testWait(50)
		t1.Kill(func() {
			order <- 3
		})
		testWait(20)
		order <- 1
	})
	t1.Read(func(t *Thread) {
		t.Read(func(t *Thread) {
			ch <- t == nil
		}, W(t2))
		t.Read(func(t *Thread) {
			testWait(150)
			order <- 2
			t.Write(func(t *Thread) {
				ch <- t == nil
				close(ch)
			})
		})
	})
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
	t1.Write(func(t *Thread) {
		ch <- t == nil
		close(ch)
	}, W(t2), W(t3))
	testCheckTrue(ch, t)
}

func TestKill3(t *testing.T) {
	ch := make(chan bool, 10)
	s := testSpawner(10)
	t1 := s.NewThead()
	t2 := s.NewThead()
	t2.Write(func(*Thread) {
		testWait(50)
		t2.Kill(nil)
	})
	t3 := s.NewThead()
	t1.Write(func(t *Thread) {
		ch <- t == nil
		close(ch)
	}, W(t2), W(t3))
	testCheckTrue(ch, t)
}
