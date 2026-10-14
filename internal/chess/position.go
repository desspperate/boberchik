package chess

const maxPieces = 32

const Ranks = 8
const Files = 8

type Position struct {
	pieces           [maxPieces]*piece
	board            [Files][Ranks]*piece
	activeColor      color
	whiteCastleKing  bool
	whiteCastleQueen bool
	blackCastleKing  bool
	blackCastleQueen bool
	enPassantSquare  square
	halfmoveClock    uint8
	fullmoveCnt      uint16
}

func (pos *Position) putPiece(p *piece, sq square) {
	if debug && !sq.isValid() {
		panicf(pos, "square %+v is invalid", sq)
	}
	if debug && pos.board[sq.file][sq.rank] != nil {
		panicf(pos, "square %+v already taken", sq)
	}

	for i := range maxPieces {
		if pos.pieces[i] == nil {
			p.square.file = sq.file
			p.square.rank = sq.rank
			pos.pieces[i] = p
			pos.board[sq.file][sq.rank] = p
			return
		}
	}

	panicf(pos, "%d pieces already on a board", maxPieces)
}

func (pos *Position) pieceAt(sq square) *piece {
	if debug && !sq.isValid() {
		panicf(pos, "square %+v is not legal", sq)
	}

	return pos.board[sq.file][sq.rank]
}

// en passant is not considered: it can only attack the pawn, so it never threatens the king
func (pos *Position) isSafe(sq square, from color) bool {
	if debug && !sq.isValid() {
		panicf(pos, "square %+v is not legal", sq)
	}

	for _, d := range knightDeltas {
		attackerSq := square{
			file: sq.file + d.file,
			rank: sq.rank + d.rank,
		}
		if !attackerSq.isValid() {
			continue
		}
		if p := pos.pieceAt(attackerSq); p != nil && p.kind == knight && p.color == from {
			return false
		}
	}

	for i, d := range queenDirections {
		for j := int8(1); ; j++ {
			attackerSq := square{
				file: sq.file + d.file*j,
				rank: sq.rank + d.rank*j,
			}
			if !attackerSq.isValid() {
				break
			}
			if p := pos.pieceAt(attackerSq); p != nil {
				if p.color == from && (p.kind == queen || p.kind == bishop && i < 4 || p.kind == rook && i >= 4 || p.kind == king && j == 1) {
					return false
				}
				break
			}
		}
	}

	var rankMultiplier int8
	if from == black {
		rankMultiplier = 1
	} else {
		rankMultiplier = -1
	}
	for _, d := range lateralBlackPawnDeltas {
		attackerSq := square{
			file: sq.file - d.file,
			rank: sq.rank - d.rank*rankMultiplier,
		}
		if !attackerSq.isValid() {
			continue
		}
		if p := pos.pieceAt(attackerSq); p != nil && p.kind == pawn && p.color == from {
			return false
		}
	}

	return true
}

func (pos *Position) PseudoLegalMoves() []move {
	var mvs []move
	for _, p := range pos.pieces {
		if p != nil && pos.activeColor == p.color && !p.captured {
			pmvs := pseudoLegalMoves(pos, p.square)
			if pmvs != nil {
				mvs = append(mvs, pmvs...)
			}
		}
	}
	return mvs
}
