package instance

import (
	"log"
	"sync"
)

type player struct {
	leader  *player
	hp      int
	id      string
	maxHp   int
	limiter *limiter
}

func (p *player) shift() {
	if !p.hasLeader() {
		p.id = ""
		return
	}
	p.id = p.leader.id
	p.hp = p.leader.hp
	p.normalize()
	p.limiter.players[p.id] = p
	p.leader.shift()
	return
}

func (p *player) isAlive() bool {
	return p.id != ""
}

func (p *player) hasLeader() bool {
	return p.leader != nil && p.leader.isAlive()
}

func (p *player) normalize() {
	if p.hp > p.maxHp {
		p.hp = p.maxHp
	}
}

func (p *player) promote(new *player, debuff bool) {
	new.maxHp += 1
	if debuff {
		p.maxHp--
		p.normalize()
	}
	if !p.hasLeader() {
		p.leader = new
		new.reset()
		return
	}
	p.leader.promote(new, debuff)
}

func (p *player) up() {
	if p.hp < p.maxHp {
		p.hp += 1
	}
	player := p
	for player.hasLeader() {
		if player.leader.hit() {
			leaderId := player.leader.id
			player.leader.id = player.id
			player.id = leaderId
			player.reset()
			player.leader.reset()
			p.limiter.players[player.id] = player
			p.limiter.players[player.leader.id] = player.leader
		}
		player = player.leader
	}
}

func (p *player) reset() {
	p.hp = p.maxHp
}

func (p *player) hit() bool {
	p.hp -= 1
	return p.hp == 0
}

type limiter struct {
	mu      sync.Mutex
	tail    *player
	players map[string]*player
	limit   int
}

func newLimiter(limit int) *limiter {
	if limit < 1 {
		log.Fatalf("Limit can't be less than 1")
	}
	return &limiter{
		mu:      sync.Mutex{},
		players: make(map[string]*player),
		limit:   limit,
	}
}

func (l *limiter) touch(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	p, ok := l.players[id]
	if !ok {
		return
	}
	p.up()

}

func (l *limiter) delete(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	p, has := l.players[id]
	if !has {
		return
	}
	p.shift()
	delete(l.players, id)
}

func (l *limiter) add(id string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	p := &player{
		leader:  nil,
		maxHp:   0,
		id:      id,
		limiter: l,
	}
	_, has := l.players[id]
	if has {
		log.Fatal("duplicate")
	}
	l.players[id] = p
	removed := ""
	if len(l.players) > l.limit {
		removed = l.tail.id
		l.tail = l.tail.leader
		if removed == l.tail.id {
			log.Fatal("REMOVED ", removed, " new tail", l.tail.id)
		}
		delete(l.players, removed)
	}
	if !l.hasTail() {
		l.tail = p
		return removed
	}
	l.tail.promote(p, removed != "")
	return removed

}

func (l *limiter) hasTail() bool {
	return l.tail != nil && l.tail.isAlive()
}
