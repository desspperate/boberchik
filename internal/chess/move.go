package chess

import "fmt"

type Move struct {
	from      square
	to        square
	promotion pieceKind
}

var NoMove = Move{}

type undoInfo struct {
	castleRights    [2][2]bool
	enPassantSquare square
	halfmoveClock   uint8

	movedPiece struct {
		from      square
		piece     *piece
		promotion pieceKind
	}
	capturedPiece *piece
	castleRook    struct {
		from square
		rook *piece
	}
}

const noPromotion = king

type vector square

var knightDeltas = []vector{
	{2, 1},
	{2, -1},
	{-2, 1},
	{-2, -1},
	{1, 2},
	{1, -2},
	{-1, 2},
	{-1, -2},
}

var rookDirections = [4]vector{
	{0, 1},
	{0, -1},
	{1, 0},
	{-1, 0},
}

var bishopDirections = [4]vector{
	{1, 1},
	{1, -1},
	{-1, 1},
	{-1, -1},
}

// do NOT CHANGE the order grouped by piece type [b,r] in queenDirections, function Position.isSafe relies on it
var queenDirections = [8]vector{
	bishopDirections[0], bishopDirections[1], bishopDirections[2], bishopDirections[3],
	rookDirections[0], rookDirections[1], rookDirections[2], rookDirections[3],
}

var lateralBlackPawnDeltas = []vector{
	{1, 1},
	{-1, 1},
}

func pseudoLegalMoves(pos *Position, sq square) []Move {
	p := pos.pieceAt(sq)

	if p == nil {
		return nil
	}

	var mvs []Move
	switch p.kind {
	case knight:
		for _, d := range knightDeltas {
			to := square{
				file: sq.file + d.file,
				rank: sq.rank + d.rank,
			}
			if !to.isValid() {
				continue
			}
			if pp := pos.pieceAt(to); pp == nil || pp.color != p.color {
				mvs = append(mvs, Move{sq, to, noPromotion})
			}
		}
	case bishop, rook, queen:
		var dirs []vector
		switch p.kind {
		case bishop:
			dirs = bishopDirections[:]
		case rook:
			dirs = rookDirections[:]
		case queen:
			dirs = queenDirections[:]
		default:
			panicf(pos, "unexpected piece kind %+v on square %+v", p.kind, sq)
		}

		for _, d := range dirs {
			for i := int8(1); ; i++ {
				to := square{
					file: sq.file + d.file*i,
					rank: sq.rank + d.rank*i,
				}
				if !to.isValid() {
					break
				}
				if pp := pos.pieceAt(to); pp == nil || pp.color != p.color {
					mvs = append(mvs, Move{sq, to, noPromotion})
					if pp != nil {
						break
					}
				} else {
					break
				}
			}
		}
	case pawn:
		forward := pawnForward[p.color]

		promotionRank := (7 + 7*forward) / 2

		add := func(to square) {
			if to.rank == promotionRank {
				mvs = append(
					mvs,
					Move{sq, to, knight},
					Move{sq, to, bishop},
					Move{sq, to, rook},
					Move{sq, to, queen},
				)
			} else {
				mvs = append(mvs, Move{sq, to, noPromotion})
			}
		}

		to := square{
			file: sq.file,
			rank: sq.rank + 1*forward,
		}
		if pp := pos.pieceAt(to); pp == nil {
			add(to)

			startRank := (7 - 5*forward) / 2
			if sq.rank == startRank {
				to.rank += 1 * forward
				if pp := pos.pieceAt(to); pp == nil {
					mvs = append(mvs, Move{sq, to, noPromotion})
				}
			}
		}

		for _, d := range lateralBlackPawnDeltas {
			to := square{
				file: sq.file + d.file,
				rank: sq.rank + d.rank*forward,
			}
			if !to.isValid() {
				continue
			}
			pp := pos.pieceAt(to)
			if pp != nil && pp.color != p.color || pos.enPassantSquare == to && pos.enPassantSquare != noEnPassantSquare {
				add(to)
			}
		}
	case king:
		filesStart := max(0, sq.file-1)
		filesEnd := min(Files-1, sq.file+1)
		ranksStart := max(0, sq.rank-1)
		ranksEnd := min(Ranks-1, sq.rank+1)

		for file := filesStart; file <= filesEnd; file++ {
			for rank := ranksStart; rank <= ranksEnd; rank++ {
				to := square{file: file, rank: rank}
				if pp := pos.pieceAt(to); pp == nil || pp.color != p.color {
					mvs = append(mvs, Move{sq, to, noPromotion})
				}
			}
		}

		if pos.isSafe(sq, p.color.opposite()) {
			var sideMultipliers []int8
			for i, multiplier := range [2]int8{1, -1} {
				if pos.castleRights[p.color][i] {
					sideMultipliers = append(sideMultipliers, multiplier)
				}
			}
			for _, multiplier := range sideMultipliers {
				var somethingOnPath bool
				for file := sq.file + 1*multiplier; file > 0 && file < Files-1; file += 1 * multiplier {
					onPath := square{file: file, rank: sq.rank}
					if pp := pos.pieceAt(onPath); pp != nil {
						somethingOnPath = true
						break
					}
				}
				if somethingOnPath {
					continue
				}

				onPath := square{file: sq.file + 1*multiplier, rank: sq.rank}
				if pos.isSafe(onPath, p.color.opposite()) {
					to := square{file: sq.file + 2*multiplier, rank: sq.rank}
					mvs = append(mvs, Move{sq, to, noPromotion})
				}

			}
		}
	default:
		panicf(pos, "unexpected piece kind %+v on square %+v", p.kind, sq)
	}

	return mvs
}

// String returns the move in UCI notation (e2e4, e7e8q).
func (m Move) String() string {
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

// ParseMove finds the legal move written in UCI notation (e2e4, e7e8q) in the position.
func (pos *Position) ParseMove(s string) (Move, error) {
	for _, m := range pos.generateMoves() {
		if m.String() == s {
			return m, nil
		}
	}
	return NoMove, fmt.Errorf("move %q is not legal in %s", s, pos.Fen())
}
