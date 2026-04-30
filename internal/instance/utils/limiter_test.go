// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLimiterClocklessFrequencyEvictsLeastProtected(t *testing.T) {
	l := NewLimiter(3)
	for _, id := range []string{"a", "b", "c"} {
		if removed := l.Add(id); removed != "" {
			t.Fatalf("unexpected removal while below limit: %q", removed)
		}
	}
	assertLimiterOrder(t, l, "a", "b", "c")

	l.TouchHeavy("a")
	assertLimiterOrder(t, l, "b", "a", "c")

	l.TouchLight("a")
	assertLimiterOrder(t, l, "b", "c", "a")

	if removed := l.Add("d"); removed != "b" {
		t.Fatalf("removed %q, want b", removed)
	}
	assertLimiterOrder(t, l, "c", "a", "d")
	assertLimiterInvariants(t, l)
}

func TestLimiterTouchTypes(t *testing.T) {
	lightLimiter := NewLimiter(3)
	heavyLimiter := NewLimiter(3)
	for _, id := range []string{"a", "b", "c"} {
		lightLimiter.Add(id)
		heavyLimiter.Add(id)
	}

	lightLimiter.TouchLight("a")
	lightLimiter.TouchLight("a")
	assertLimiterOrder(t, lightLimiter, "b", "a", "c")

	heavyLimiter.TouchHeavy("a")
	heavyLimiter.TouchHeavy("a")
	assertLimiterOrder(t, heavyLimiter, "b", "c", "a")
}

func TestLimiterDeleteShiftsChain(t *testing.T) {
	l := NewLimiter(4)
	for _, id := range []string{"a", "b", "c", "d"} {
		l.Add(id)
	}

	l.Delete("b")
	assertLimiterOrder(t, l, "a", "c", "d")

	l.Delete("a")
	assertLimiterOrder(t, l, "c", "d")

	l.Delete("d")
	assertLimiterOrder(t, l, "c")

	l.Delete("c")
	assertLimiterOrder(t, l)
	assertLimiterInvariants(t, l)
}

func TestLimiterDeleteRefillsShiftedPlayerAfterFullDrain(t *testing.T) {
	l := NewLimiter(3)
	for _, id := range []string{"a", "b", "c"} {
		l.Add(id)
	}
	assertLimiterOrder(t, l, "a", "b", "c")

	leader := l.players["c"]
	middle := l.players["b"]
	leader.hp = 0

	l.Delete("b")

	if got := l.players["c"]; got != middle {
		t.Fatalf("shifted c is stored in %#v, want former middle player %#v", got, middle)
	}
	if middle.hp != middle.maxHp {
		t.Fatalf("shifted player hp = %d, want refill to maxHp %d", middle.hp, middle.maxHp)
	}
	assertLimiterOrder(t, l, "a", "c")
}

func TestLimiterOverflowInsertStartsFragile(t *testing.T) {
	l := NewLimiter(2)
	l.Add("a")
	l.Add("b")

	if removed := l.Add("c"); removed != "a" {
		t.Fatalf("removed %q, want a", removed)
	}

	inserted := l.players["c"]
	if inserted.hp != 1 {
		t.Fatalf("overflow inserted hp = %d, want 1", inserted.hp)
	}
	if inserted.maxHp <= inserted.hp {
		t.Fatalf("overflow inserted maxHp = %d, want greater than hp %d", inserted.maxHp, inserted.hp)
	}
	assertLimiterOrder(t, l, "b", "c")
}

func TestLimiterUnderLimitInsertKeepsGraceHp(t *testing.T) {
	l := NewLimiter(3)
	l.Add("a")
	l.Add("b")

	inserted := l.players["b"]
	if inserted.hp != inserted.maxHp/2 {
		t.Fatalf("under-limit inserted hp = %d, want half maxHp %d", inserted.hp, inserted.maxHp/2)
	}
	assertLimiterOrder(t, l, "a", "b")
}

func TestLimiterUnknownIDsAreNoops(t *testing.T) {
	l := NewLimiter(3)
	for _, id := range []string{"a", "b", "c"} {
		l.Add(id)
	}

	l.TouchLight("missing")
	l.TouchHeavy("missing")
	l.Delete("missing")

	assertLimiterOrder(t, l, "a", "b", "c")
}

func TestLimiterLimitOneEvictsPreviousAndStaysUsable(t *testing.T) {
	l := NewLimiter(1)
	if removed := l.Add("a"); removed != "" {
		t.Fatalf("removed %q, want none", removed)
	}
	if removed := l.Add("b"); removed != "a" {
		t.Fatalf("removed %q, want a", removed)
	}

	l.TouchLight("b")
	l.TouchHeavy("b")
	assertLimiterOrder(t, l, "b")

	l.Delete("b")
	assertLimiterOrder(t, l)

	if removed := l.Add("c"); removed != "" {
		t.Fatalf("removed %q after empty add, want none", removed)
	}
	assertLimiterOrder(t, l, "c")
}

func TestLimiterInvariantsAcrossMixedOperations(t *testing.T) {
	l := NewLimiter(5)
	live := map[string]bool{}
	nextID := 0

	for step := 0; step < 160; step++ {
		order := limiterOrder(l)
		switch {
		case len(order) == 0 || step%4 == 0:
			id := fmt.Sprintf("id-%03d", nextID)
			nextID++
			removed := l.Add(id)
			live[id] = true
			if removed != "" {
				if !live[removed] {
					t.Fatalf("step %d removed non-live id %q", step, removed)
				}
				delete(live, removed)
			}
		case step%4 == 1:
			l.TouchHeavy(order[step%len(order)])
		case step%4 == 2:
			l.TouchLight(order[step%len(order)])
		default:
			id := order[step%len(order)]
			l.Delete(id)
			delete(live, id)
		}
		assertLimiterInvariants(t, l)
		if len(l.players) != len(live) {
			t.Fatalf("step %d live count = %d, players = %d", step, len(live), len(l.players))
		}
		for id := range l.players {
			if !live[id] {
				t.Fatalf("step %d player map has non-live id %q", step, id)
			}
		}
	}
}

func assertLimiterOrder(t *testing.T, l Limiter, want ...string) {
	t.Helper()
	got := limiterOrder(l)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("limiter order = %v, want %v", got, want)
	}
	assertLimiterInvariants(t, l)
}

func limiterOrder(l Limiter) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return limiterOrderLocked(l)
}

func limiterOrderLocked(l Limiter) []string {
	var ids []string
	for player := l.tail; player != nil && player.alive(); player = player.leader {
		ids = append(ids, player.id)
	}
	return ids
}

func assertLimiterInvariants(t *testing.T, l Limiter) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.players) == 0 {
		if l.hasTail() {
			t.Fatal("empty limiter has a live tail")
		}
		return
	}
	if !l.hasTail() {
		t.Fatal("non-empty limiter has no live tail")
	}

	seenIDs := map[string]bool{}
	seenPlayers := map[*player]bool{}
	for player := l.tail; player != nil && player.alive(); player = player.leader {
		if seenPlayers[player] {
			t.Fatalf("cycle at player for id %q", player.id)
		}
		seenPlayers[player] = true
		if player.id == "" {
			t.Fatal("live player has empty id")
		}
		if seenIDs[player.id] {
			t.Fatalf("duplicate live id %q in chain", player.id)
		}
		seenIDs[player.id] = true
		if l.players[player.id] != player {
			t.Fatalf("player map for %q points to %#v, want %#v", player.id, l.players[player.id], player)
		}
	}
	if len(seenIDs) != len(l.players) {
		t.Fatalf("chain has %d ids, player map has %d", len(seenIDs), len(l.players))
	}
	for id := range l.players {
		if !seenIDs[id] {
			t.Fatalf("player map id %q is missing from chain", id)
		}
	}
}
