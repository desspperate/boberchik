package main

import (
	"fmt"
	"log"

	"github.com/desspperate/chess-engine/internal/chess"
)

func main() {
	pos, err := chess.DecodeFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		log.Fatal(err)
	}
	if err := pos.Validate(); err != nil {
		log.Fatal(err)
	}
	mvs := pos.PseudoLegalMoves()
	fmt.Println(len(mvs))
}
