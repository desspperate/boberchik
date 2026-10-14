package chess

import "fmt"

const debug = true

var outer string

func panicf(pos *Position, format string, args ...any) {
	m := fmt.Sprintf(format, args...)
	if outer != "" {
		panic(outer + "\n(while dumping) " + m)
	}
	outer = m
	if pos != nil {
		m += "\nFEN: " + pos.Fen()
	}
	outer = ""
	panic(m)
}
