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

const DefaultCacheControl string = "public, max-age=31536000, immutable"

// Conf defines global configuration for sessions, instances,
// client-server communication, and performance. Defaults are auto-initialized.
type Conf struct {
	// SessionInstanceLimit is the max number of page instances per session.
	// When exceeded, the oldest inactive ones are suspended.
	// Default: 12.
	SessionInstanceLimit int
	// SessionTTL controls how long session lives after last activity.
	// Default behavior (value 0): InstanceTTL
	SessionTTL time.Duration
	// InstanceConnectTimeout controls how long new instance waits
	// before shutdown for the first client connection.
	// Default: RequestTimeout
	InstanceConnectTimeout time.Duration
	// InstanceGoroutineLimit is the max goroutines per page instance.
	// Controls resource use for rendering and reactive updates. Default: 8.
	InstanceGoroutineLimit int
	// InstanceTTL is how long an inactive instance is kept before cleanup.
	// Active = browser connected. Default: 40minutes.
	InstanceTTL time.Duration
	// ServerCacheControl defines cache control header value for JS and CSS
	// resources prepared by the framework.
	// Default "public, max-age=31536000, immutable"
	ServerCacheControl string
	// ServerDisableGzip disables gzip compression for HTML, JS, and CSS if true.
	ServerDisableGzip bool
	// ServerSessionCookiePrefix sets the internal Doors session cookie name prefix.
	// Use it when you want browser-enforced cookie prefix rules, such as
	// __Host- or __Secure-. Empty by default.
	ServerSessionCookiePrefix string
	// ServerSessionCookieNoSecure disables the Secure attribute on the internal
	// Doors session cookie. Use only for plain HTTP development.
	ServerSessionCookieNoSecure bool
	// DisconnectHiddenTimer is how long hidden/background instances stay connected.
	// Default: InstanceTTL/2.
	DisconnectHiddenTimer time.Duration
	// RequestTimeout is the max duration of a client-server request.
	// Default: 30s.
	RequestTimeout time.Duration
	// SolitaireRollTime is the max lifetime of a solitaire sender or report
	// receiver request before rolling to a replacement request.
	// Default: 15s.
	SolitaireRollTime time.Duration
	// SolitaireSyncTimeout is the max pending duration of a server-to-client
	// sync task, including user calls. Exceeding this kills the instance.
	// Default: InstanceTTL.
	SolitaireSyncTimeout time.Duration
	// SolitaireFrameTime is the max time to buffer an outgoing server-to-client
	// frame before forcing a sync flush.
	// Default: ~33.3ms (30 FPS).
	SolitaireFrameTime time.Duration
	// SolitaireFrameSize is the max buffered bytes in an outgoing
	// server-to-client frame before forcing a sync flush.
	// Default: 32 KB.
	SolitaireFrameSize int
	// SolitaireDisableGzip disables gzip compression for solitaire sync payloads if true.
	SolitaireDisableGzip bool
	// SolitaireQueue is the max queued server-to-client sync task.
	// Exceeding this kills the instance. Default: 1024.
	SolitaireQueue int
	// SolitairePending is the max unresolved server-to-client sync tasks.
	// Throttles sending when reached. Default: 256.
	SolitairePending int
	// SolitaireDisableReportStreaming disables browser streaming request bodies
	// for client-to-server reports. When true, each report uses a standalone
	// JSON POST. Default: false.
	SolitaireDisableReportStreaming bool
	// SolitaireReportSize is the max size of one client-to-server report in a
	// streaming report body. Default: 8 MB.
	SolitaireReportSize int
	// SolitaireReportTimeout is the max time to receive and decode one
	// client-to-server report. Default: min(5s, SolitaireRollTime).
	SolitaireReportTimeout time.Duration
	// SolitaireMaxRTT caps the RTT estimate used for sync probing while the
	// server has pending work but no frame ready to flush. Default: 1s.
	SolitaireMaxRTT time.Duration
}

type SolitaireConf struct {
	// Roll is the effective solitaire request roll interval.
	Roll time.Duration
	// FrameSize is the effective outgoing server-to-client frame size limit.
	FrameSize int
	// FlushTime is the effective outgoing server-to-client frame time limit.
	FlushTime time.Duration
	// DisableGzip is the effective solitaire payload gzip flag.
	DisableGzip bool
	// DisableReportStreaming is the effective report streaming flag.
	DisableReportStreaming bool
	// Queue is the effective queued sync task limit.
	Queue int
	// Pending is the effective unresolved sync task limit.
	Pending int
	// SyncTimeout is the effective sync task timeout.
	SyncTimeout time.Duration
	// ReportSize is the effective streaming report message size limit.
	ReportSize int
	// ReportTimeout is the effective single-report receive timeout.
	ReportTimeout time.Duration
	// MaxRTT is the effective RTT estimate cap.
	MaxRTT time.Duration
}

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
		ReportSize:             s.SolitaireReportSize,
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
	if s.SolitaireReportSize <= 0 {
		s.SolitaireReportSize = 8 * 1024 * 1024
	}
	if s.SolitaireMaxRTT <= 0 {
		s.SolitaireMaxRTT = time.Second
	}
	if s.SolitaireReportTimeout <= 0 {
		s.SolitaireReportTimeout = min(5*time.Second, s.SolitaireRollTime)
	}
}

// InitDefaults fills zero or invalid values in s with Doors defaults.
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
	s.solitaireDefaults()
}
