package solitaire

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
)

type frameSyncer struct {
	Conf    *common.SolitaireConf
	rtt     time.Duration
	mu      sync.Mutex
	reports []uint64
	last    time.Time
}

func (g *frameSyncer) RTT() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.last.IsZero() {
		if g.rtt == 0 {
			return g.minRTT()
		}
		return g.rtt
	}
	if time.Since(g.last) > g.rtt {
		return g.calcRTT(g.last)
	}
	return g.rtt
}

func (g *frameSyncer) Report(id uint64, ts int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.reports = append(g.reports, id)

	if g.last.UnixMilli() == ts {
		g.last = time.Time{}
	}
	g.rtt = g.calcRTT(time.UnixMilli(ts))
}

const alpha = 0.4

func (g *frameSyncer) minRTT() time.Duration {
	return g.Conf.FlushTime * 2
}

func (g *frameSyncer) calcRTT(point time.Time) time.Duration {
	newRTT := g.rtt + time.Duration(alpha*float64(time.Since(point)-g.rtt))
	return min(max(newRTT, g.minRTT()), g.Conf.MaxRTT)
}

func (g *frameSyncer) Collect() []byte {
	g.mu.Lock()
	defer g.mu.Unlock()
	tuple := make([]any, 0, 2)
	g.last = time.Now()
	tuple = append(tuple, g.last.UnixMilli())
	if len(g.reports) > 0 {
		tuple = append(tuple, g.reports)
	}
	data, err := json.Marshal(tuple)
	if err != nil {
		panic(err)
	}
	g.reports = g.reports[:0]
	return data
}

type result struct {
	output json.RawMessage
	err    error
}

func (r *result) UnmarshalJSON(data []byte) error {
	var a [2]json.RawMessage
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	var e *string
	err = json.Unmarshal(a[1], &e)
	if err != nil {
		return err
	}
	if e != nil {
		r.err = errors.New(*e)
		return nil
	}
	r.output = a[0]
	return nil
}

type gap struct {
	beg uint64
	end uint64
}

func (m gap) MarshalJSON() ([]byte, error) {
	if m.beg == m.end {
		return json.Marshal([1]uint64{m.beg})
	}
	return json.Marshal([2]uint64{m.beg, m.end})
}

func (m *gap) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return errors.New("empty result array")
	}
	err = json.Unmarshal(parts[0], &m.beg)
	if err != nil {
		return err
	}
	if len(parts) > 1 {
		err = json.Unmarshal(parts[1], &m.end)
		if err != nil {
			return err
		}
		return nil
	} else {
		m.end = m.beg
	}
	return nil
}

type report struct {
	ID      uint64
	TS      int64
	Gaps    []gap
	Results map[uint64]result
}

func (m *report) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("report must be an array: %w", err)
	}
	if len(arr) < 2 {
		return fmt.Errorf("report array must contain at least [ID, TS]")
	}
	if len(arr) > 4 {
		return fmt.Errorf("report array must contain at most [ID, TS, Results, Gaps]")
	}
	if err := json.Unmarshal(arr[0], &m.ID); err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}
	if err := json.Unmarshal(arr[1], &m.TS); err != nil {
		return fmt.Errorf("invalid report TS: %w", err)
	}
	if len(arr) > 2 {
		if err := json.Unmarshal(arr[2], &m.Results); err != nil {
			return fmt.Errorf("invalid report Results: %w", err)
		}
	}
	if len(arr) > 3 {
		if err := json.Unmarshal(arr[3], &m.Gaps); err != nil {
			return fmt.Errorf("invalid report Gaps: %w", err)
		}
	}
	return nil
}
