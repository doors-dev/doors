package common

import "time"

type SystemConf struct {
	SessionInstanceLimit     int
	SessionExpiration        time.Duration
	SessionCookieExpiration  time.Duration
	InstanceGoroutineLimit   int
	InstanceTTL              time.Duration
	ServerDisableGzip        bool
	ClientHiddenSleepTimer   time.Duration
	SolitaireRollSize        int
	SolitaireRequestTimeout  time.Duration
	SolitaireRollPendingTime time.Duration
	SolitaireQueue           int
	SolitairePending         int
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
		SleepTimeout:   s.ClientHiddenSleepTimer,
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
	if s.InstanceTTL == 0 {
		s.InstanceTTL = 15 * time.Minute
	}
	if s.InstanceTTL < s.SolitaireRequestTimeout * 2{
		s.InstanceTTL = s.SolitaireRequestTimeout * 2
	}
	if s.ClientHiddenSleepTimer == 0 {
		s.ClientHiddenSleepTimer = 3 * time.Minute
	}
}
