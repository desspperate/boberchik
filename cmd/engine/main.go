package main

import (
	"log"
	"os"

	"github.com/desspperate/boberchik/internal/uci"
)

func main() {
	if err := uci.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
