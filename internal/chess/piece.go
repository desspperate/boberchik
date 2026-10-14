package chess

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

type color byte

const (
	white color = iota
	black
)

func (c color) opposite() color {
	return c ^ 1
}

type piece struct {
	kind     pieceKind
	color    color
	captured bool
	square   square
}
