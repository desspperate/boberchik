package chess

import (
	"errors"
	"fmt"
)

const maxPieces = 32

const maxRanks = 8
const maxFiles = 8

type Position struct {
	pieces           [maxPieces]*piece
	board            [maxFiles][maxRanks]*piece
	activeColor      color
	whiteCastleKing  bool
	whiteCastleQueen bool
	blackCastleKing  bool
	blackCastleQueen bool
	enPassantSquare  square
	halfmoveClock    uint8
	fullmoveCnt      uint16
}

func (p *Position) putPiece(piece *piece, s square) error {
	if s.file >= maxFiles {
		return errors.New("file out of bounds")
	}
	if s.rank >= maxRanks {
		return errors.New("rank out of bounds")
	}

	if p.board[s.file][s.rank] != nil {
		return fmt.Errorf("square [%d;%d] on a board is already taken", s.file, s.rank)
	}

	for i := range maxPieces {
		if p.pieces[i] == nil {
			piece.square.file = s.file
			piece.square.rank = s.rank
			p.pieces[i] = piece
			p.board[s.file][s.rank] = piece
			return nil
		}
	}
	return fmt.Errorf("%d pieces already on a board", maxPieces)
}
