package chess

type square struct {
	file byte
	rank byte
}

type pType byte

const (
	king pType = iota
	pawn
	queen
	rook
	bishop
	knight
)

type color byte

const (
	white color = iota
	black
)

type piece struct {
	pType      pType
	color      color
	isCaptured bool
	square     square
}
