package main

import (
	"fmt"
	"log"

	"github.com/desspperate/boberchik/internal/chess"
)

var toColor = [2]string{"white", "black"}

func main() {
	pos, err := chess.DecodeFen("1k1K4/8/8/8/6R1/8/8/8 w - - 0 1")
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
			fmt.Printf("No move at halfmove %d\n", i)
			return
		}
		pos.Make(move)
		fmt.Printf("the best move for %s: %+v\n", toColor[i%2], move)
		i++
	}
}
