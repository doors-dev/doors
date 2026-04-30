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
	"log"
	"sync"
)

const hpStep int = 2

// player is one rank in the limiter's clockless frequency tournament. The tail
// player is the next eviction candidate. Each leader is harder to displace
// because its maxHp grows by hpStep.
//
// A player intentionally stores no timestamp, counter window, or access log.
// Frequency is represented by position in the chain and by carried hp: a
// touched id attacks only its leader, and repeated touches are needed to climb
// through protected ranks.
type player struct {
	leader  *player
	hp      int
	id      string
	maxHp   int
	limiter *limiter
}

func (p *player) shiftUnsafe() {
	if !p.hasLeader() {
		p.id = ""
		return
	}
	p.id = p.leader.id
	p.hp = p.leader.hp
	p.normalize()
	p.limiter.players[p.id] = p
	p.leader.shiftUnsafe()
}

func (p *player) alive() bool {
	return p.id != ""
}

func (p *player) hasLeader() bool {
	return p.leader != nil && p.leader.alive()
}

func (p *player) normalize() {
	p.hp = max(p.hp, p.maxHp)
}

func (p *player) promote(new *player, decayRanks bool) {
	new.maxHp += hpStep
	if decayRanks {
		p.maxHp -= hpStep
		p.normalize()
	}
	if !p.hasLeader() {
		p.leader = new
		if decayRanks {
			new.resetMin()
		} else {
			new.resetLow()
		}
		return
	}
	p.leader.promote(new, decayRanks)
}

func (p *player) upLight() {
	if p.hp < p.maxHp {
		p.hp += 1
	}
	if !p.hasLeader() {
		return
	}
	if p.leader.hit() {
		p.swapWithLeader()
	}
}

func (p *player) upHeavy() {
	if p.hp < p.maxHp {
		p.hp += 1
	}
	player := p
	for player.hasLeader() {
		if player.leader.hit() {
			player.swapWithLeader()
		}
		player = player.leader
	}
}

func (p *player) swapWithLeader() {
	leaderId := p.leader.id
	p.leader.id = p.id
	p.id = leaderId
	p.resetFull()
	p.leader.resetLow()
	p.limiter.players[p.id] = p
	p.limiter.players[p.leader.id] = p.leader
}

func (p *player) resetFull() {
	p.hp = p.maxHp
}

func (p *player) resetLow() {
	p.hp = max(p.maxHp/2, 1)
}

func (p *player) resetMin() {
	p.hp = 1
}

func (p *player) hit() bool {
	p.hp -= 1
	return p.hp <= 0
}

// Limiter is an approximate LFU limiter without time-series data. It keeps a
// bounded frequency ranking of active instance ids and returns the current tail
// id from Add when the configured limit is exceeded.
type Limiter = *limiter

// limiter holds the private implementation behind Limiter.
//
// Invariants while the mutex is held:
//   - an empty limiter has no live tail; tail may still point to a dead player.
//   - a non-empty limiter's tail points to a live player.
//   - every live player id is present in players and maps back to that player.
//   - every id in players appears exactly once in the tail -> leader chain.
//   - tail is the next id add will evict when len(players) exceeds limit.
type limiter struct {
	mu      sync.Mutex
	tail    *player
	players map[string]*player
	limit   int
}

func NewLimiter(limit int) Limiter {
	if limit < 1 {
		log.Fatalf("limiter: limit can't be less than 1")
	}
	return &limiter{
		mu:      sync.Mutex{},
		players: make(map[string]*player),
		limit:   limit,
	}
}

func (l *limiter) TouchLight(id string) {
	l.touch(id, false)
}

func (l *limiter) TouchHeavy(id string) {
	l.touch(id, true)
}

func (l *limiter) touch(id string, heavy bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	player, ok := l.players[id]
	if !ok {
		return
	}
	if heavy {
		player.upHeavy()
		return
	}
	player.upLight()
}

func (l *limiter) Delete(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	player, has := l.players[id]
	if !has {
		return
	}
	player.shiftUnsafe()
	delete(l.players, id)
}

func (l *limiter) Add(id string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	player := &player{
		leader:  nil,
		id:      id,
		limiter: l,
	}
	_, has := l.players[id]
	if has {
		log.Fatal("limiter: duplicate")
	}
	l.players[id] = player
	removed := ""
	if len(l.players) > l.limit {
		removed = l.evictTailLocked()
	}
	if !l.hasTail() {
		l.tail = player
		return removed
	}
	l.tail.promote(player, removed != "")
	return removed

}

func (l *limiter) evictTailLocked() string {
	removed := l.tail.id
	l.tail = l.tail.leader
	if l.tail != nil && removed == l.tail.id {
		log.Fatal("limiter: removed ", removed, " new tail", l.tail.id)
	}
	delete(l.players, removed)
	return removed
}

func (l *limiter) hasTail() bool {
	return l.tail != nil && l.tail.alive()
}
