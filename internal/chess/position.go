package chess

import (
	"fmt"
)

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

func (pos *Position) putPiece(p *piece, sq square) error {
	if !sq.isValid() {
		return fmt.Errorf("square [%d;%d] is not legal", sq.file, sq.rank)
	}

	if pos.board[sq.file][sq.rank] != nil {
		return fmt.Errorf("square [%d;%d] on a board is already taken", sq.file, sq.rank)
	}

	for i := range maxPieces {
		if pos.pieces[i] == nil {
			p.square.file = sq.file
			p.square.rank = sq.rank
			pos.pieces[i] = p
			pos.board[sq.file][sq.rank] = p
			return nil
		}
	}
	return fmt.Errorf("%d pieces already on a board", maxPieces)
}

func (pos *Position) pieceAt(sq square) (*piece, error) {
	if !sq.isValid() {
		return nil, fmt.Errorf("square [%d;%d] is not legal", sq.file, sq.rank)
	}

	return pos.board[sq.file][sq.rank], nil
}

// en passant is not considered: it can only attack the pawn, so it never threatens the king
func (pos *Position) isSafe(sq square, from color) (bool, error) {
	if !sq.isValid() {
		return false, fmt.Errorf("square [%d;%d] is not legal", sq.file, sq.rank)
	}

	for _, d := range knightDeltas {
		attackerSq := square{
			file: sq.file + d.file,
			rank: sq.rank + d.rank,
		}
		if p, err := pos.pieceAt(attackerSq); err == nil && p != nil && p.kind == knight && p.color == from {
			return false, nil
		}
	}

	for i, d := range queenDirections {
		for j := int8(1); ; j++ {
			attackerSq := square{
				file: sq.file + d.file*j,
				rank: sq.rank + d.rank*j,
			}
			if p, err := pos.pieceAt(attackerSq); err != nil {
				break
			} else if p != nil {
				if p.color == from && (p.kind == queen || p.kind == bishop && i < 4 || p.kind == rook && i >= 4 || p.kind == king && j == 1) {
					return false, nil
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
		if p, err := pos.pieceAt(attackerSq); err == nil && p != nil && p.kind == pawn && p.color == from {
			return false, nil
		}
	}

	return true, nil
}

func (pos *Position) PseudoLegalMoves() ([]move, error) {
	var mvs []move
	for _, p := range pos.pieces {
		if p != nil && pos.activeColor == p.color && !p.captured {
			pmvs, err := pseudoLegalMoves(pos, p.square)
			if err != nil {
				return nil, fmt.Errorf("pseudoLegalMoves failed %w", err)
			}
			if pmvs != nil {
				mvs = append(mvs, pmvs...)
			}
		}
	}
	return mvs, nil
}
