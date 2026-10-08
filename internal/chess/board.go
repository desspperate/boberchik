package chess

import (
	"errors"
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

func (pp *Position) putPiece(p *piece, sq square) error {
	if sq.file >= Files {
		return errors.New("file out of bounds")
	}
	if sq.rank >= Ranks {
		return errors.New("rank out of bounds")
	}

	if pp.board[sq.file][sq.rank] != nil {
		return fmt.Errorf("square [%d;%d] on a board is already taken", sq.file, sq.rank)
	}

	for i := range maxPieces {
		if pp.pieces[i] == nil {
			p.square.file = sq.file
			p.square.rank = sq.rank
			pp.pieces[i] = p
			pp.board[sq.file][sq.rank] = p
			return nil
		}
	}
	return fmt.Errorf("%d pieces already on a board", maxPieces)
}

func (pp *Position) pieceAt(sq square) *piece {
	return pp.board[sq.file][sq.rank]
}
