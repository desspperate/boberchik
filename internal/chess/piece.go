package chess

import "fmt"

type square struct {
	file int8
	rank int8
}

func (sq square) isValid() bool {
	return sq.file >= 0 && sq.file < Files && sq.rank >= 0 && sq.rank < Ranks
}

type pieceKind byte

const (
	king pieceKind = iota
	pawn
	queen
	rook
	bishop
	knight
)

func (k pieceKind) String() string {
	switch k {
	case king:
		return "king"
	case pawn:
		return "pawn"
	case queen:
		return "queen"
	case rook:
		return "rook"
	case bishop:
		return "bishop"
	case knight:
		return "knight"
	default:
		panic(fmt.Sprintf("unexpected piece kind: %d", k))
	}
}

var pieceKinds = []pieceKind{king, pawn, queen, rook, bishop, knight}

type color byte

const (
	white color = iota
	black
)

func (c color) String() string {
	switch c {
	case white:
		return "white"
	case black:
		return "black"
	default:
		panic(fmt.Sprintf("unexpected color %d", c))
	}
}

var colors = []color{white, black}

func (c color) opposite() color {
	return c ^ 1
}

type piece struct {
	kind     pieceKind
	color    color
	captured bool
	square   square
}

var pawnForward = [2]int8{white: -1, black: 1}
