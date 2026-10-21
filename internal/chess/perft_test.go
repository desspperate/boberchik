package chess

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

// Reference values: https://www.chessprogramming.org/Perft_Results
var perftCases = []struct {
	name   string
	fen    string
	counts []uint64 // counts[d-1] = perft(d)
	short  int      // depth used with go test -short
}{
	{
		name:   "startpos",
		fen:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		counts: []uint64{20, 400, 8902, 197281, 4865609},
		short:  4,
	},
	{
		name:   "kiwipete",
		fen:    "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		counts: []uint64{48, 2039, 97862, 4085603},
		short:  3,
	},
	{
		name:   "pos3",
		fen:    "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		counts: []uint64{14, 191, 2812, 43238, 674624},
		short:  4,
	},
	{
		name:   "pos4",
		fen:    "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		counts: []uint64{6, 264, 9467, 422333},
		short:  3,
	},
	{
		name:   "pos5",
		fen:    "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		counts: []uint64{44, 1486, 62379, 2103487},
		short:  3,
	},
	{
		name:   "pos6",
		fen:    "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		counts: []uint64{46, 2079, 89890, 3894594},
		short:  3,
	},
}

func TestPerft(t *testing.T) {
	for _, tc := range perftCases {
		t.Run(tc.name, func(t *testing.T) {
			pos, err := DecodeFen(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			if err := pos.Validate(); err != nil {
				t.Fatal(err)
			}

			maxDepth := len(tc.counts)
			if testing.Short() {
				maxDepth = tc.short
			}
			for depth := 1; depth <= maxDepth; depth++ {
				want := tc.counts[depth-1]
				got := perft(&pos, depth)
				if got != want {
					t.Errorf("depth %d: got %d, want %d\ndivide:\n%s", depth, got, want, divideString(&pos, depth))
					return // already wrong, no point going deeper
				}
			}
		})
	}
}

// TestPerftDivide prints a per-root-move breakdown for comparison with Stockfish
// (`position fen ...`, then `go perft N`):
//
//	PERFT_FEN="..." PERFT_DEPTH=3 go test -run TestPerftDivide -v ./internal/chess
func TestPerftDivide(t *testing.T) {
	fen := os.Getenv("PERFT_FEN")
	if fen == "" {
		t.Skip("PERFT_FEN is not set")
	}
	depth := 1
	if d := os.Getenv("PERFT_DEPTH"); d != "" {
		if _, err := fmt.Sscan(d, &depth); err != nil {
			t.Fatal(err)
		}
	}
	pos, err := DecodeFen(fen)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + divideString(&pos, depth))
}

func perft(pos *Position, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	moves := pos.generateMoves()
	if depth == 1 {
		return uint64(len(moves))
	}

	var nodes uint64
	for _, m := range moves {
		before := ""
		if debug {
			before = pos.Fen()
		}

		ud := pos.Make(m)
		nodes += perft(pos, depth-1)
		pos.unmake(ud)

		if debug {
			if after := pos.Fen(); after != before {
				panicf(pos, "unmake mismatch after %s\nbefore: %s\nafter:  %s", moveString(m), before, after)
			}
		}
	}
	return nodes
}

func divideString(pos *Position, depth int) string {
	type row struct {
		move  string
		nodes uint64
	}
	var rows []row
	var total uint64
	for _, m := range pos.generateMoves() {
		ud := pos.Make(m)
		n := perft(pos, depth-1)
		pos.unmake(ud)

		rows = append(rows, row{moveString(m), n})
		total += n
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].move < rows[j].move })

	s := ""
	for _, r := range rows {
		s += fmt.Sprintf("%s: %d\n", r.move, r.nodes)
	}
	s += fmt.Sprintf("\nmoves: %d, nodes: %d\n", len(rows), total)
	return s
}

// moveString returns the move in UCI notation (e2e4, e7e8q), the same format Stockfish prints.
func moveString(m move) string {
	sq := func(s square) string {
		return string(rune('a'+s.file)) + string(rune('8'-s.rank))
	}
	s := sq(m.from) + sq(m.to)
	switch m.promotion {
	case queen:
		s += "q"
	case rook:
		s += "r"
	case bishop:
		s += "b"
	case knight:
		s += "n"
	}
	return s
}
