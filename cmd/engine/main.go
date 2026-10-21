package main

import (
	"fmt"
	"log"

	"github.com/desspperate/boberchik/internal/chess"
)

var toColor = [2]string{"white", "black"}

func main() {
	pos, err := chess.DecodeFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		log.Fatal(err)
	}
	if err := pos.Validate(); err != nil {
		log.Fatal(err)
	}

	i := 0
	for {
		move := pos.BestMove()
		if move == chess.NoMove {
			i += 1
			fmt.Printf("No move at halfmove, %s win %d. FEN: %s\n", toColor[i%2], i, pos.Fen())
			return
		}
		if i == 9000 {
			fmt.Printf("Draw by limit. FEN: %s\n", pos.Fen())
			return
		}
		pos.Make(move)
		fmt.Printf("the best move for %s: %+v\n. FEN: %s\n", toColor[i%2], move, pos.Fen())
		i++
	}
}
