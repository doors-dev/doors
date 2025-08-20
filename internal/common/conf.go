
package common

import "time"

// SystemConf contains system-wide configuration options that control various aspects
// of the doors framework including session management, instance lifecycle, performance
// settings, and client-server communication parameters.
type SystemConf struct {
	// SessionInstanceLimit controls the maximum number of page instances that can
	// exist simultaneously within a single session. When this limit is exceeded,
	// the oldest and least active instances are suspended to make room for new ones.
	// Default: 6.
	SessionInstanceLimit int

	// SessionExpiration sets the session expiration timeout. The session will be
	// automatically cleaned up after this duration of inactivity, regardless of
	// whether instances are present. 
	// Default: 0 (session cleaned up when no more instances left).
	SessionExpiration time.Duration

	// SessionCookieExpiration sets the expiration time for the session cookie sent
	// to the client's browser. This should typically be longer than SessionExpiration
	// to allow for proper session restoration. 
	// Default and recommended: 0 (session cookies expire when browser closes).
	SessionCookieExpiration time.Duration

	// InstanceGoroutineLimit sets the maximum number of goroutines that can be spawned
	// per page instance. This controls resource usage for render tasks and reactive
	// updates within each instance. Default: 16.
	InstanceGoroutineLimit int

	// InstanceTTL (Time To Live) sets how long an inactive page instance remains alive
	// before being automatically killed. Instances are considered active when they
	// have ongoing connections or recent client interactions. Default: 15 minutes
	// or at least twice SolitaireRequestTimeout (whichever is greater).
	InstanceTTL time.Duration

	// ServerDisableGzip disables gzip compression for HTTP responses when set to true.
	// Compression is enabled by default to reduce bandwidth usage, but may be disabled
	// for debugging or when using external compression. Default: false.
	ServerDisableGzip bool

	// DisconnectHiddenTimer controls how long hidden/background page instances keep
	// connection active. This helps manage CPU/RAM usage when many tabs are open 
	// but not actively viewed. Default: 10 minutes.
	DisconnectHiddenTimer time.Duration

	// SolitaireRollSize sets the maximum number of bytes that can be sent in a single
	// response before commanding the client to roll to a new request. Default: 8 KB.
	SolitaireRollSize int

	// SolitaireRequestTimeout sets the maximum duration a client-server communication
	// request can remain open. Default: 30 seconds.
	SolitaireRequestTimeout time.Duration

	// SolitaireRollPendingTime sets how long the server waits if there are pending calls
	// before commanding the client to roll to a new request, even if SolitaireRollSize
	// hasn't been reached. This balances latency against request efficiency.
	// Default: 100 milliseconds.
	SolitaireRollPendingTime time.Duration

	// SolitaireQueue sets the maximum number of client calls that can be queued
	// waiting to be sent to the client. When this limit is exceeded, the instance
	// will be terminated to prevent memory exhaustion. Default: 1024.
	SolitaireQueue int

	// SolitairePending sets the maximum number of calls that can be sent to the
	// client but not yet acknowledged. Default: 256.
	SolitairePending int
}

type SolitaireConf struct {
	Request         time.Duration
	RollSize        int
	RollTime        time.Duration
	RollPendingTime time.Duration
	Queue           int
	Pending         int
}

func GetSolitaireConf(s *SystemConf) *SolitaireConf {
	return &SolitaireConf{
		Request:         (s.SolitaireRequestTimeout * 2) / 3,
		RollSize:        s.SolitaireRollSize,
		RollTime:        s.SolitaireRequestTimeout / 2,
		RollPendingTime: s.SolitaireRollPendingTime,
		Queue:           s.SolitaireQueue,
		Pending:         s.SolitairePending,
	}
}

type ClientConf struct {
	TTL            time.Duration
	RequestTimeout time.Duration
	SleepTimeout   time.Duration
}

func GetClientConf(s *SystemConf) *ClientConf {
	return &ClientConf{
		TTL:            s.InstanceTTL,
		SleepTimeout:   s.DisconnectHiddenTimer,
		RequestTimeout: s.SolitaireRequestTimeout,
	}
}

func (s *SystemConf) solitaireDefaults() {
	if s.SolitaireRollSize <= 0 {
		s.SolitaireRollSize = 8 * 1024
	}
	if s.SolitaireRequestTimeout <= 0 {
		s.SolitaireRequestTimeout = 30 * time.Second
	}
	if s.SolitaireRollPendingTime <= 0 {
		s.SolitaireRollPendingTime = 100 * time.Millisecond
	}
	if s.SolitaireQueue <= 0 {
		s.SolitaireQueue = 1024
	}
	if s.SolitairePending <= 0 {
		s.SolitairePending = 256
	}
}

func InitDefaults(s *SystemConf) {
	s.solitaireDefaults()
	if s.SessionInstanceLimit < 1 {
		s.SessionInstanceLimit = 6
	}
	if s.InstanceGoroutineLimit <= 0 {
		s.InstanceGoroutineLimit = 16
	}
	if s.InstanceTTL <= 0 {
		s.InstanceTTL = 15 * time.Minute
	}
	if s.InstanceTTL < s.SolitaireRequestTimeout * 2{
		s.InstanceTTL = s.SolitaireRequestTimeout * 2
	}
	if s.DisconnectHiddenTimer <= 0 {
		s.DisconnectHiddenTimer = 10 * time.Minute
	}
}
