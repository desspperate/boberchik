// Package uci implements the subset of the Universal Chess Interface protocol
// that GUIs and lichess-bot need to play against the engine.
package uci

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/desspperate/boberchik/internal/chess"
)

const (
	engineName   = "Boberchik"
	engineAuthor = "desspperate"

	// DefaultDepth is used when "go" does not specify a depth.
	// Time controls (wtime, btime, movetime) are ignored for now.
	DefaultDepth = 5

	startFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

// Run reads UCI commands from in and writes responses to out until "quit" or EOF.
// Diagnostics go to the standard logger (stderr), never to out.
func Run(in io.Reader, out io.Writer) error {
	w := bufio.NewWriter(out)
	defer w.Flush()

	pos, err := chess.DecodeFen(startFen)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(in)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "uci":
			fmt.Fprintf(w, "id name %s\n", engineName)
			fmt.Fprintf(w, "id author %s\n", engineAuthor)
			fmt.Fprintln(w, "uciok")
		case "isready":
			fmt.Fprintln(w, "readyok")
		case "ucinewgame":
			pos, err = chess.DecodeFen(startFen)
			if err != nil {
				return err
			}
		case "position":
			p, err := parsePosition(fields[1:])
			if err != nil {
				log.Printf("uci: position: %v", err)
				break
			}
			pos = p
		case "go":
			m := pos.BestMove(parseDepth(fields[1:]))
			if m == chess.NoMove {
				fmt.Fprintln(w, "bestmove 0000")
			} else {
				fmt.Fprintf(w, "bestmove %s\n", m)
			}
		case "quit":
			return nil
		default:
			// setoption, debug, stop, ponderhit, ... are not supported yet
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}
	return sc.Err()
}

// parsePosition handles the arguments of
// "position startpos [moves m1 m2 ...]" and "position fen <6 fields> [moves m1 m2 ...]".
func parsePosition(args []string) (chess.Position, error) {
	if len(args) == 0 {
		return chess.Position{}, fmt.Errorf("missing startpos or fen")
	}

	var fen string
	var rest []string
	switch args[0] {
	case "startpos":
		fen, rest = startFen, args[1:]
	case "fen":
		if len(args) < 1+chess.FenSegments {
			return chess.Position{}, fmt.Errorf("fen must have %d fields", chess.FenSegments)
		}
		fen = strings.Join(args[1:1+chess.FenSegments], " ")
		rest = args[1+chess.FenSegments:]
	default:
		return chess.Position{}, fmt.Errorf("expected startpos or fen, got %q", args[0])
	}

	pos, err := chess.DecodeFen(fen)
	if err != nil {
		return chess.Position{}, err
	}
	if err := pos.Validate(); err != nil {
		return chess.Position{}, err
	}

	if len(rest) == 0 {
		return pos, nil
	}
	if rest[0] != "moves" {
		return chess.Position{}, fmt.Errorf("expected moves, got %q", rest[0])
	}
	for _, s := range rest[1:] {
		m, err := pos.ParseMove(s)
		if err != nil {
			return chess.Position{}, err
		}
		pos.Make(m)
	}
	return pos, nil
}

// parseDepth returns N from "go ... depth N ...", or DefaultDepth.
func parseDepth(args []string) int32 {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != "depth" {
			continue
		}
		if d, err := strconv.ParseInt(args[i+1], 10, 32); err == nil && d > 0 {
			return int32(d)
		}
	}
	return DefaultDepth
}
