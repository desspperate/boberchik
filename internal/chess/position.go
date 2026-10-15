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

	for _, d := range lateralBlackPawnDeltas {
		attackerSq := square{
			file: sq.file - d.file,
			rank: sq.rank - d.rank*pawnForward[from],
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

const maxHalfmoveClock = 150

const minFullmoveCnt = 1
const maxFullmoveCnt = 8849 // https://tomrocksmaths.com/wp-content/uploads/2026/06/longestchessgame-katie-cam.pdf

const kingExactly = 1

const pawnMax = 8

const rbkInit = 2
const qInit = 1

const colorMax = 16

var initKingSq = []square{white: {file: 4, rank: 7}, black: {file: 4, rank: 0}}

// do not change low level (in color) order
var initRookSqs = [][]square{
	white: {{file: 7, rank: 7}, {file: 0, rank: 7}},
	black: {{file: 7, rank: 0}, {file: 0, rank: 0}},
}
var castleSide = []string{0: "king", 1: "queen"}

const whiteEnPassantRank = 5
const blackEnPassantRank = 2

func (pos *Position) Validate() error {
	if pos.halfmoveClock > maxHalfmoveClock {
		return fmt.Errorf("halfmove clock must be in range [0,%d]", maxHalfmoveClock)
	}
	if pos.fullmoveCnt < minFullmoveCnt || pos.fullmoveCnt > maxFullmoveCnt {
		return fmt.Errorf("fullmove count must be in range [%d,%d]", minFullmoveCnt, maxFullmoveCnt)
	}

	if pos.enPassantSquare != noEnPassantSquare {
		if pos.halfmoveClock != 0 {
			return errors.New("en passant set, so halfmove clock must be 0")
		}
		if pos.activeColor == white && pos.enPassantSquare.rank != blackEnPassantRank {
			return fmt.Errorf("its %v's move and en passant filled so it must be %d", pos.activeColor, blackEnPassantRank)
		}
		if pos.activeColor == black && pos.enPassantSquare.rank != whiteEnPassantRank {
			return fmt.Errorf("its %v's move and en passant filled so it must be %d", pos.activeColor, whiteEnPassantRank)
		}
		atEnPassantSq := pos.pieceAt(pos.enPassantSquare)
		if atEnPassantSq != nil {
			return fmt.Errorf("en passant set, so it's square must be empty")
		}
		aboveEnPassant := pos.pieceAt(square{
			file: pos.enPassantSquare.file,
			rank: pos.enPassantSquare.rank + pawnForward[pos.activeColor.opposite()],
		})
		if aboveEnPassant == nil || aboveEnPassant.kind != pawn || aboveEnPassant.color != pos.activeColor.opposite() {
			return fmt.Errorf("its %v's move and en passant filled so above it must be a pawn of %v (inactive) color", pos.activeColor, pos.activeColor.opposite())
		}
	}

	played := pos.fullmoveCnt*2 - 2
	if pos.activeColor == black {
		played++
	}
	if uint16(pos.halfmoveClock) > played {
		return fmt.Errorf("halfmove clock %d exceeds halfmoves played %d", pos.halfmoveClock, played)
	}

	castleRights := [2][2]bool{
		white: {pos.whiteCastleKing, pos.whiteCastleQueen},
		black: {pos.blackCastleKing, pos.blackCastleQueen},
	}
	for _, c := range colors {
		var atKingSq *piece
		for i, r := range castleRights[c] {
			if !r {
				continue
			}
			if atKingSq == nil {
				atKingSq = pos.pieceAt(initKingSq[c])
				if atKingSq == nil || atKingSq.kind != king || atKingSq.color != c {
					return fmt.Errorf("%v can not have their king not at init square and %sside castle right", c, castleSide[i])
				}
			}
			atRookSq := pos.pieceAt(initRookSqs[c][i])
			if atRookSq == nil || atRookSq.kind != rook || atRookSq.color != c {
				return fmt.Errorf("%v can not have their rook not at init square and %sside castle right", c, castleSide[i])
			}
		}
	}

	var pieceCount [2][6]int
	var colorCount [2]int
	for _, p := range pos.pieces {
		if p == nil || p.captured {
			continue
		}

		pieceCount[p.color][p.kind]++
		colorCount[p.color]++

		switch {
		case p.kind == pawn:
			forward := pawnForward[p.color]
			if p.square.rank == (7+7*forward)/2 {
				return fmt.Errorf("no %v pawn allowed at %d (promotion) rank", p.color, p.square.rank)
			}
			if p.square.rank == (7-7*forward)/2 {
				return fmt.Errorf("no %v pawn allowed at %d (under initial) rank", p.color, p.square.rank)
			}
		case p.kind == king:
			if p.color != pos.activeColor && !pos.isSafe(p.square, pos.activeColor) {
				return fmt.Errorf("no checks for king of %v (inactive) side allowed", p.color)
			}
		default:
		}

	}
	for _, c := range colors {
		if colorCount[c] > colorMax {
			return fmt.Errorf("side %v must have %d pieces or fewer", c, colorMax)
		}
	}
	for _, k := range pieceKinds {
		for _, c := range colors {
			pc := pieceCount[c][k]
			switch k {
			case king:
				if pc != kingExactly {
					return fmt.Errorf("%v side must have exactly %d king", c, kingExactly)
				}
			case queen, rook, bishop, knight, pawn:
				initCnt := rbkInit
				if k == queen {
					initCnt = qInit
				} else if k == pawn {
					initCnt = 0
				}
				if pc > initCnt+pawnMax {
					return fmt.Errorf("either %v side must have %d %vs or fewer", c, initCnt+pawnMax, k)
				}
			default:
				panicf(pos, "unexpected piece kind %+v", k)
			}
		}
	}

	return nil
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
