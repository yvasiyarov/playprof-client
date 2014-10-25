package profile

import (
	"fmt"
)

// A ReporterTransfer is used to transfer reports from client to play-prof server.
type Metrics struct {
	FuncStats map[uint64]*StatItem // stats per PC.
	MemStats  []StatItem           // allocation pool.
}
type StatItem struct {
	Self    [4]int64
	Cumul   [4]int64
	Callees map[uint64][4]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		FuncStats: make(map[uint64]*StatItem),
		MemStats:  make([]StatItem, 64),
	}
}

func (m *Metrics) Symbols() (as []uint64) {
	seen := make(map[uint64]bool, len(m.FuncStats))
	for a, v := range m.FuncStats {
		if !seen[a] {
			seen[a] = true
			as = append(as, a)
		}
		for b := range v.Callees {
			if !seen[b] {
				seen[b] = true
				as = append(as, b)
			}
		}
	}
	return
}

// Add registers data for a given stack trace. There may be at most
// 4 count arguments, as needed in heap profiles.
func (m *Metrics) Add(trace []uint64, count ...int64) error {
	if len(count) > 4 {
		return fmt.Errorf("too many counts (%d) to register in reporter", len(count))
	}
	// Only the last point.
	s := m.getStats(trace[0])
	for i, n := range count {
		s.Self[i] += n
	}
	// Record cumulated stats.
	seen := make(map[uint64]bool, len(trace))
	for i, a := range trace {
		s := m.getStats(a)
		if !seen[a] {
			seen[a] = true
			for j, n := range count {
				s.Cumul[j] += n
			}
		}
		if i > 0 {
			callee := trace[i-1]
			if s.Callees == nil {
				s.Callees = make(map[uint64][4]int64)
			}
			edges := s.Callees[callee]
			for j, n := range count {
				edges[j] += n
			}
			s.Callees[callee] = edges
		}
	}
	return nil
}

func (m *Metrics) getStats(key uint64) *StatItem {
	if p := m.FuncStats[key]; p != nil {
		return p
	}
	s := &m.MemStats[0]
	m.MemStats = m.MemStats[1:]
	m.FuncStats[key] = s
	return s
}
