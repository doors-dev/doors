// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import "time"

// SystemConf defines global configuration for sessions, instances,
// client-server communication, and performance. Defaults are auto-initialized.
type SystemConf struct {
	// SessionInstanceLimit is the max number of page instances per session.
	// When exceeded, the oldest inactive ones are suspended.
	// Default: 12.
	SessionInstanceLimit int

	// SessionTTL controls how long session lives.
	// Default behavior (value 0): session ends, when no more
	// instances left, cookie expires when browser closes.
	SessionTTL time.Duration

	// InstanceGoroutineLimit is the max goroutines per page instance.
	// Controls resource use for rendering and reactive updates. Default: 16.
	InstanceGoroutineLimit int

	// InstanceTTL is how long an inactive instance is kept before cleanup.
	// Active = browser connected. Default: 40minutes or ≥2×RequestTimeout.
	InstanceTTL time.Duration

	// ServerDisableGzip disables gzip compression for HTML, JS, and CSS if true.
	ServerDisableGzip bool

	// DisconnectHiddenTimer is how long hidden/background instances stay connected.
	// Default: InstanceTTL/2.
	DisconnectHiddenTimer time.Duration

	// RequestTimeout is the max duration of a client-server request (hook).
	// Default: 30s.
	RequestTimeout time.Duration

	// SolitairePing is the max idle time before rolling the request.
	// Default: 15s.
	SolitairePing time.Duration

	// SolitaireSyncTimeout is the max pending duration of a server→client sync task,
	// including user calls. Exceeding this kills the instance.
	// Default: InstanceTTL.
	SolitaireSyncTimeout time.Duration

	// SolitaireRollTimeout is how long an active sync connection lasts before
	// rolling to a new one if the queue is long. Default: 1s.
	SolitaireRollTimeout time.Duration

	// SolitaireFlushTimeout is the max time before forcing a flush.
	// Default: 30ms
	SolitaireFlushTimeout time.Duration

	// SolitaireFlushSizeLimit is the max buffered bytes before forcing a flush.
	// Default: 32 KB
	SolitaireFlushSizeLimit int

	// SolitaireQueue is the max queued server→client sync task.
	// Exceeding this kills the instance. Default: 1024.
	SolitaireQueue int

	// SolitairePending is the max unresolved server→client sync tasks.
	// Throttles sending when reached. Default: 256.
	SolitairePending int
}

type SolitaireConf struct {
	Ping         time.Duration
	FlushSize    int
	RollDuration time.Duration
	FlushTimeout time.Duration
	Queue        int
	Pending      int
	SyncTimeout  time.Duration
}

func GetSolitaireConf(s *SystemConf) *SolitaireConf {
	return &SolitaireConf{
		SyncTimeout:  s.SolitaireSyncTimeout,
		Ping:         s.SolitairePing,
		FlushSize:    s.SolitaireFlushSizeLimit,
		RollDuration: s.SolitaireRollTimeout,
		FlushTimeout: s.SolitaireFlushTimeout,
		Queue:        s.SolitaireQueue,
		Pending:      s.SolitairePending,
	}
}

type ClientConf struct {
	TTL            time.Duration
	RequestTimeout time.Duration
	Ping           time.Duration
	SleepTimeout   time.Duration
	Detached       bool
}

func GetClientConf(s *SystemConf) *ClientConf {
	return &ClientConf{
		TTL:            s.InstanceTTL,
		SleepTimeout:   s.DisconnectHiddenTimer,
		RequestTimeout: s.RequestTimeout,
		Ping:           s.SolitairePing,
	}
}

func (s *SystemConf) solitaireDefaults() {
	if s.SolitaireFlushSizeLimit <= 0 {
		s.SolitaireFlushSizeLimit = 32 * 1024
	}
	if s.SolitaireSyncTimeout <= 0 {
		s.SolitaireSyncTimeout = s.InstanceTTL
	}
	if s.SolitaireFlushTimeout <= 0 {
		s.SolitaireFlushTimeout = 30 * time.Millisecond
	}
	if s.SolitaireRollTimeout <= 0 {
		s.SolitaireRollTimeout = 1 * time.Second
	}
	if s.SolitaireQueue <= 0 {
		s.SolitaireQueue = 1024
	}
	if s.SolitairePending <= 0 {
		s.SolitairePending = 256
	}
	if s.SolitairePing <= 0 {
		s.SolitairePing = 15 * time.Second
	}
}

func InitDefaults(s *SystemConf) {
	if s.RequestTimeout <= 0 {
		s.RequestTimeout = 30 * time.Second
	}
	if s.SessionInstanceLimit < 1 {
		s.SessionInstanceLimit = 12
	}
	if s.InstanceGoroutineLimit <= 0 {
		s.InstanceGoroutineLimit = 16
	}
	if s.InstanceTTL <= 0 {
		s.InstanceTTL = 40 * time.Minute
	}
	if s.InstanceTTL < s.RequestTimeout*2 {
		s.InstanceTTL = s.RequestTimeout * 2
	}
	if s.DisconnectHiddenTimer <= 0 {
		s.DisconnectHiddenTimer = s.InstanceTTL / 2
	}
	s.solitaireDefaults()
}
