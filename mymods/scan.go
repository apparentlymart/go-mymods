package mymods

import (
	"bufio"
	"bytes"
)

var sep = []byte{'\t'}

type scanner struct {
	sc  *bufio.Scanner
	cur entry
}

type entry struct {
	kw      []byte
	path    []byte
	version []byte
}

func scanTable(buf []byte) *scanner {
	var sc scanner
	r := bytes.NewReader(buf)
	sc.sc = bufio.NewScanner(r)
	return &sc
}

func (sc *scanner) Scan() bool {
	for {
		more := sc.sc.Scan()
		if !more {
			return false
		}
		line := sc.sc.Bytes()
		tab := bytes.Index(line, sep)
		if tab == -1 {
			// Line seems invalid, so skip
			continue
		}
		sc.cur.kw = line[:tab]
		line = line[tab+1:]
		tab = bytes.Index(line, sep)
		if tab == -1 {
			sc.cur.path = line
			sc.cur.version = nil
			return true
		}
		sc.cur.path = line[:tab]
		line = line[tab+1:]
		tab = bytes.Index(line, sep)
		if tab == -1 {
			sc.cur.version = line
		} else {
			sc.cur.version = line[:tab]
		}
		return true
	}
}

// HasKeyword compares the given string with the current keyword without
// allocating any memory.
func (sc *scanner) HasKeyword(kw string) bool {
	if len(kw) != len(sc.cur.kw) {
		return false
	}
	for i, got := range sc.cur.kw {
		if kw[i] != got {
			return false
		}
	}
	return true
}

// Entry returns the current entry from the scanner. The bytes returned may be
// used only until the next call to Scan.
func (sc *scanner) Entry() entry {
	return sc.cur
}
