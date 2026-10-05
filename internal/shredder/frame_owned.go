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

package shredder

import (
	"log/slog"
)

type ownedOperation int

const (
	scheduleOwned ownedOperation = iota
	releaseOwned
	activateOwned
	callbackOwned
)

type frameOwner interface {
	lock(op ownedOperation)
	unlock(op ownedOperation, done bool)
}

type ownedFrame struct {
	owner    frameOwner
	active   bool
	released bool
	counter  int
	buffer   []executable
}

func (f *ownedFrame) Release() {
	f.owner.lock(releaseOwned)
	if f.released {
		f.owner.unlock(releaseOwned, false)
		return
	}
	f.released = true
	f.owner.unlock(releaseOwned, f.isCompleted())
}

func (f *ownedFrame) schedule(e executable, logger *slog.Logger) {
	f.owner.lock(scheduleOwned)
	if f.isCompleted() {
		f.owner.unlock(scheduleOwned, false)
		logger.Warn(
			"attempted to schedule on completed frame",
		)
		e.execute(func(error) {})
		return
	}
	f.counter += 1
	if !f.active {
		f.buffer = append(f.buffer, e)
		f.owner.unlock(scheduleOwned, false)
		return
	}
	f.owner.unlock(scheduleOwned, false)
	e.execute(f.callback)
}

func (f *ownedFrame) activate() {
	f.owner.lock(activateOwned)
	if f.active {
		f.owner.unlock(activateOwned, f.isCompleted())
		return
	}
	f.active = true
	if f.isCompleted() {
		f.owner.unlock(activateOwned, true)
		return
	}
	f.owner.unlock(activateOwned, false)
	for i, e := range f.buffer {
		e.execute(f.callback)
		f.buffer[i] = nil
	}
	f.buffer = f.buffer[:0]
}

func (f *ownedFrame) callback(error) {
	f.owner.lock(callbackOwned)
	f.counter -= 1
	f.owner.unlock(callbackOwned, f.isCompleted())
}

func (f *ownedFrame) isCompleted() bool {
	return f.released && f.counter == 0 && f.active
}
