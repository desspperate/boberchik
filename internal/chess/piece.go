package chess

type square struct {
	file uint8
	rank uint8
}

type kind byte

const (
	king kind = iota
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
	kind     kind
	color    color
	Captured bool
	square   square
}
