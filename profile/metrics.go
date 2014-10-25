package profile

import (
	//"debug/elf"
	"fmt"
	//"io"
	//"io/ioutil"
	//"net/http"
	//"sort"
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

/*
func NewResolverTransferFromSymbolMap(symbolsMap map[uint64]string) ResolverTransfer {
	symbols := make([]SymbolInfo, 0, len(symbolsMap))
	for addr, label := range symbolsMap {
		si := SymbolInfo{
			Address: int64(addr),
			Label:   label,
		}
		symbols = append(symbols, si)
	}
	return ResolverTransfer{
		Symbols: symbols,
	}
}
func NewResolverTransferFromSymbolList(s elfSymbolTable) ResolverTransfer {
	symbols := make([]SymbolInfo, 0, len(s))
	for _, symbol := range s {
		si := SymbolInfo{
			Address: int64(symbol.Value),
			Label:   symbol.Name,
		}
		symbols = append(symbols, si)
	}
	return ResolverTransfer{
		Symbols: symbols,
	}
}

*/

/*
func (r *Reporter) Total(col int) (t int64) {
	for _, s := range r.stats {
		t += s.Self[col]
	}
	return
}


func (r *Reporter) SetExecutable(filename string) error {
	f, err := elf.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	symbols, err := f.Symbols()
	if err != nil {
		return err
	}
	sort.Sort(elfSymbolTable(symbols))
	r.Resolver = elfSymbolTable(symbols)
	return nil
}
*/

/*
func (r *Reporter) ReportByFunc(column int) []ReportLine {
	lines := make(map[string][2][4]float64, len(r.stats))
	for a, v := range r.stats {
		name := r.Resolver.Resolve(a)
		s := lines[name]
		for i := 0; i < 4; i++ {
			s[0][i] += float64(v.Self[i])
			s[1][i] += float64(v.Cumul[i])
		}
		lines[name] = s
	}
	entries := make([]ReportLine, 0, len(lines))
	for name, values := range lines {
		entries = append(entries, ReportLine{Name: name,
			Self: values[0], Cumul: values[1]})
	}
	sort.Sort(bySelf{entries, column})
	return entries
}

func (r *Reporter) ReportByPC() []ReportLine {
	return nil
}

type ReportLine struct {
	Name        string
	Self, Cumul [4]float64
}

type bySelf struct {
	slice []ReportLine
	col   int
}

func (s bySelf) Len() int      { return len(s.slice) }
func (s bySelf) Swap(i, j int) { s.slice[i], s.slice[j] = s.slice[j], s.slice[i] }

func (s bySelf) Less(i, j int) bool {
	left, right := s.slice[i].Self[s.col], s.slice[j].Self[s.col]
	if left > right {
		return true
	}
	if left == right {
		return s.slice[i].Name < s.slice[j].Name
	}
	return false
}

type byCumul bySelf

func (s byCumul) Len() int      { return len(s.slice) }
func (s byCumul) Swap(i, j int) { s.slice[i], s.slice[j] = s.slice[j], s.slice[i] }

func (s byCumul) Less(i, j int) bool {
	left, right := s.slice[i].Cumul[s.col], s.slice[j].Cumul[s.col]
	if left > right {
		return true
	}
	if left == right {
		return s.slice[i].Name < s.slice[j].Name
	}
	return false
}
func (r *Reporter) ExportToTransferObject() *ReporterTransfer {
	transfer := new(ReporterTransfer)

	transfer.FreeStats = r.freeStats
	transfer.Stats = r.stats

	if resolver, ok := r.Resolver.(*RemoteResolver); ok {
		transfer.Resolver = NewResolverTransferFromSymbolMap(resolver.Symbols)
	} else if resolver, ok := r.Resolver.(elfSymbolTable); ok {
		transfer.Resolver = NewResolverTransferFromSymbolList(resolver)
	}
	return transfer
}

func (r *Reporter) ImportFromTransferObject(transfer *ReporterTransfer) {
	r.freeStats = transfer.FreeStats
	r.stats = transfer.Stats
}
*/

/*
// A Resolver associates a symbol to a numeric address.
type Resolver interface {
	Resolve(uint64) string
}

type TableResolver struct {
	Symbols map[uint64]string
}

func NewTableResolver(symbols map[uint64]string) *TableResolver {
	return &TableResolver{
		Symbols: symbols,
	}
}

func (r *TableResolver) Resolve(addr uint64) string {
	return r.Symbols[addr]
}

// Symbol resolvers.

type elfSymbolTable []elf.Symbol

func (s elfSymbolTable) Len() int           { return len(s) }
func (s elfSymbolTable) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s elfSymbolTable) Less(i, j int) bool { return s[i].Value < s[j].Value }

func (s elfSymbolTable) Resolve(addr uint64) string {
	min, max := 0, len(s)
	for max-min > 1 {
		med := (min + max) / 2
		a := s[med].Value
		if a < addr {
			min = med
		} else {
			max = med
		}
	}
	if s[min].Value > addr {
		return "N/A"
	}
	return s[min].Name
}

type RemoteResolver struct {
	Url     string
	Symbols map[uint64]string
}


*/
