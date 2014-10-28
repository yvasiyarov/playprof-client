package profile

import (
	"debug/elf"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	ColCPU        = 0
	ColLiveObj    = 0
	ColLiveBytes  = 1
	ColAllocObj   = 2
	ColAllocBytes = 3
)

type Resolver struct {
	Symbols map[uint64]string
}

func NewResolver() *Resolver {
	return &Resolver{
		Symbols: make(map[uint64]string),
	}
}

func (r *Resolver) Resolve(addr uint64) string {
	return r.Symbols[addr]
}

func (r *Resolver) LoadSymbols(symbolsData []byte) error {
	lines := strings.Split(string(symbolsData), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "num_symbols") {
			continue
		}
		// Lines have the form "0xabcdef symbol_name"
		words := strings.Fields(line)
		if len(words) != 2 {
			return fmt.Errorf("bad symbol file format")
		}

		addr, err := strconv.ParseUint(words[0], 0, 64)
		if err != nil {
			return err
		}
		r.Symbols[addr] = words[1]
	}
	return nil
}

func (r *Resolver) LoadSymbolsFromExeFile(addresses []uint64, filename string) error {
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
	elfResolver := elfSymbolTable(symbols)

	for _, addr := range addresses {
		r.Symbols[addr] = elfResolver.Resolve(addr)
	}
	return nil
}

type elfSymbolTable []elf.Symbol

func (s elfSymbolTable) Len() int {
	return len(s)
}
func (s elfSymbolTable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s elfSymbolTable) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

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
