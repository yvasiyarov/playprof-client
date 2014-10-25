package profile

import (
	"debug/elf"
	"fmt"
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

func (r *Resolver) LoadSymbolsFromExeFile(filename string) error {
	f, err := elf.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	symbols, err := f.Symbols()
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		r.Symbols[symbol.Value] = symbol.Name
	}
	return nil
}
