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

func (p *Position) putPiece(piece *piece, square square) error {
	if square.file >= Files {
		return errors.New("file out of bounds")
	}
	if square.rank >= Ranks {
		return errors.New("rank out of bounds")
	}

	if p.board[square.file][square.rank] != nil {
		return fmt.Errorf("square [%d;%d] on a board is already taken", square.file, square.rank)
	}

	for i := range maxPieces {
		if p.pieces[i] == nil {
			piece.square.file = square.file
			piece.square.rank = square.rank
			p.pieces[i] = piece
			p.board[square.file][square.rank] = piece
			return nil
		}
	}
	return fmt.Errorf("%d pieces already on a board", maxPieces)
}
