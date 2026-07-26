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

package common

import "time"

// DefaultCacheControl is the Cache-Control value used when ServerCacheControl
// is empty.
const DefaultCacheControl string = "public, max-age=31536000, immutable"

// Conf is the Doors runtime configuration.
type Conf struct {
	// SessionInstanceLimit is the max number of live page instances per
	// session. Above it, the least active instance is suspended and reloads
	// on the next interaction. Default: 12.
	SessionInstanceLimit int
	// SessionTTL is how long a session survives its last request. Values at
	// or below InstanceTTL are raised to it.
	SessionTTL time.Duration
	// InstanceConnectTimeout is how long a rendered page has to connect back
	// before its instance is dropped. Default: RequestTimeout.
	InstanceConnectTimeout time.Duration
	// InstanceGoroutineLimit is the max number of goroutines one page
	// instance uses for rendering and reactive work. Default: 8.
	InstanceGoroutineLimit int
	// InstanceTTL is how long an instance is kept without a browser
	// connection before it is dropped. Default: 40m; values below twice
	// RequestTimeout are raised to that.
	InstanceTTL time.Duration
	// ServerCacheControl is the Cache-Control value for managed resources.
	// Default: public, max-age=31536000, immutable.
	ServerCacheControl string
	// ServerDisableGzip disables gzip compression for HTML, Door updates, and
	// managed resources.
	ServerDisableGzip bool
	// ServerSessionCookiePrefix is prepended to the Doors session cookie name.
	// Use it for browser-enforced prefixes such as __Host- or __Secure-.
	ServerSessionCookiePrefix string
	// ServerRequestBodyLimit is the max size in bytes of a hook or form
	// submission request body. Default: 8 MB.
	ServerRequestBodyLimit int
	// ServerSessionCookieNoSecure drops the Secure attribute from the cookies
	// Doors sets. Use it only for plain HTTP development.
	ServerSessionCookieNoSecure bool
	// DisconnectHiddenTimer is how long a hidden page stays connected before it
	// pauses; it reconnects when visible again. Default: InstanceTTL/2.
	DisconnectHiddenTimer time.Duration
	// RequestTimeout is the time limit of one request: page renders on the
	// server, hook and navigation requests in the browser. Default: 30s.
	RequestTimeout time.Duration
	// SolitaireRollTime is how long one live connection request is held open
	// before it is replaced by a fresh one. Lower it behind proxies that cut
	// long requests. Default: 15s.
	SolitaireRollTime time.Duration
	// SolitaireSyncTimeout is how long the server waits for the browser to
	// acknowledge an update before ending the instance. Default: InstanceTTL;
	// larger values are capped at it.
	SolitaireSyncTimeout time.Duration
	// SolitaireFrameTime is how long outgoing updates are batched before they
	// are sent to the browser. Default: 33.3ms.
	SolitaireFrameTime time.Duration
	// SolitaireFrameSize is how many buffered bytes force an immediate send of
	// outgoing updates. Default: 32 KB.
	SolitaireFrameSize int
	// SolitaireDisableGzip disables gzip compression of action payloads sent
	// to the browser.
	SolitaireDisableGzip bool
	// SolitaireQueue is the max number of updates waiting for delivery or
	// acknowledgement. Exceeding it ends the instance. Default: 1024.
	SolitaireQueue int
	// SolitairePending is the max number of delivered but unacknowledged
	// updates; at the limit the server stops sending. Default: 256.
	SolitairePending int
	// SolitaireDisableReportStreaming makes the browser post each message
	// separately instead of streaming one request body. The browser already
	// falls back on its own outside HTTP/2 and HTTP/3.
	SolitaireDisableReportStreaming bool
	// SolitaireReportLimit is the max size in bytes of one message the browser
	// sends back. Default: 8 MB.
	SolitaireReportLimit int
	// SolitaireReportTimeout is how long the server waits to receive one
	// message from the browser before dropping the connection. Default: 5s, or
	// SolitaireRollTime if it is smaller.
	SolitaireReportTimeout time.Duration
	// SolitaireMaxRTT caps the round-trip estimate used to pace waiting on the
	// browser while updates are unacknowledged. Values below twice
	// SolitaireFrameTime are raised to that. Default: 1s.
	SolitaireMaxRTT time.Duration
}

// SolitaireConf holds the effective update stream settings taken from Conf.
type SolitaireConf struct {
	// Roll is the effective SolitaireRollTime.
	Roll time.Duration
	// FrameSize is the effective SolitaireFrameSize.
	FrameSize int
	// FlushTime is the effective SolitaireFrameTime.
	FlushTime time.Duration
	// DisableGzip is the effective SolitaireDisableGzip.
	DisableGzip bool
	// DisableReportStreaming is the effective SolitaireDisableReportStreaming.
	DisableReportStreaming bool
	// Queue is the effective SolitaireQueue.
	Queue int
	// Pending is the effective SolitairePending.
	Pending int
	// SyncTimeout is the effective SolitaireSyncTimeout.
	SyncTimeout time.Duration
	// ReportSize is the effective SolitaireReportLimit.
	ReportSize int
	// ReportTimeout is the effective SolitaireReportTimeout.
	ReportTimeout time.Duration
	// MaxRTT is the effective SolitaireMaxRTT.
	MaxRTT time.Duration
}

// GetSolitaireConf returns the effective update stream settings of s.
func GetSolitaireConf(s *Conf) *SolitaireConf {
	return &SolitaireConf{
		SyncTimeout:            s.SolitaireSyncTimeout,
		Roll:                   s.SolitaireRollTime,
		FrameSize:              s.SolitaireFrameSize,
		FlushTime:              s.SolitaireFrameTime,
		DisableGzip:            s.SolitaireDisableGzip,
		DisableReportStreaming: s.SolitaireDisableReportStreaming,
		Queue:                  s.SolitaireQueue,
		Pending:                s.SolitairePending,
		ReportSize:             s.SolitaireReportLimit,
		ReportTimeout:          s.SolitaireReportTimeout,
		MaxRTT:                 s.SolitaireMaxRTT,
	}
}

func (s *Conf) solitaireDefaults() {
	if s.SolitaireFrameSize <= 0 {
		s.SolitaireFrameSize = 32 * 1024
	}
	if s.SolitaireSyncTimeout <= 0 {
		s.SolitaireSyncTimeout = s.InstanceTTL
	}
	if s.SolitaireSyncTimeout > s.InstanceTTL {
		s.SolitaireSyncTimeout = s.InstanceTTL
	}
	if s.SolitaireFrameTime <= 0 {
		s.SolitaireFrameTime = time.Second / 30
	}
	if s.SolitaireQueue <= 0 {
		s.SolitaireQueue = 1024
	}
	if s.SolitairePending <= 0 {
		s.SolitairePending = 256
	}
	if s.SolitaireRollTime <= 0 {
		s.SolitaireRollTime = 15 * time.Second
	}
	if s.SolitaireReportLimit <= 0 {
		s.SolitaireReportLimit = 8 * 1024 * 1024
	}
	if s.SolitaireMaxRTT <= 0 {
		s.SolitaireMaxRTT = time.Second
	}
	s.SolitaireMaxRTT = max(s.SolitaireFrameTime*2, s.SolitaireMaxRTT)
	if s.SolitaireReportTimeout <= 0 {
		s.SolitaireReportTimeout = min(5*time.Second, s.SolitaireRollTime)
	}
}

// InitDefaults fills zero or invalid values in s with Doors defaults and
// applies the documented clamps.
func InitDefaults(s *Conf) {
	if s.RequestTimeout <= 0 {
		s.RequestTimeout = 30 * time.Second
	}
	if s.SessionInstanceLimit < 1 {
		s.SessionInstanceLimit = 12
	}
	if s.InstanceGoroutineLimit <= 0 {
		s.InstanceGoroutineLimit = 8
	}
	if s.InstanceConnectTimeout <= 0 {
		s.InstanceConnectTimeout = s.RequestTimeout
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
	if s.ServerCacheControl == "" {
		s.ServerCacheControl = DefaultCacheControl
	}
	if s.SessionTTL <= s.InstanceTTL {
		s.SessionTTL = s.InstanceTTL
	}
	if s.ServerRequestBodyLimit <= 0 {
		s.ServerRequestBodyLimit = 8 * 1024 * 1024
	}
	s.solitaireDefaults()
}
