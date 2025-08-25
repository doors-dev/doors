// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import "sync"

func NewFuncCollector() *FuncCollector {
	return &FuncCollector{
		collection: make([]func(), 0),
	}
}

type FuncCollector struct {
	mu         sync.Mutex
	collection []func()
}

func (c *FuncCollector) Add(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, f)
}

func (c *FuncCollector) Apply() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, f := range c.collection {
		f()
	}
	c.collection = make([]func(), 0)
}
