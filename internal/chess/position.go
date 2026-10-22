package chess

import (
	"errors"
	"fmt"
)

const maxPieces = 32

const Ranks = 8
const Files = 8

type Position struct {
	pieces      [maxPieces]*piece
	board       [Files][Ranks]*piece
	activeColor color
	// indexed [color][side]: white king, white queen, black king, black queen
	castleRights    [2][2]bool
	enPassantSquare square
	halfmoveClock   uint8
	fullmoveCnt     uint16
	kings           [2]*piece
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
			if p.kind == king {
				pos.kings[p.color] = p
			}
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
var castleLetters = [2][2]rune{white: {'K', 'Q'}, black: {'k', 'q'}}

var enPassantRanks = [2]int8{white: 5, black: 2}

func (pos *Position) generateMoves() []Move {
	var ms []Move
	for _, p := range pos.pieces {
		if p == nil || p.color != pos.activeColor || p.captured {
			continue
		}
		plms := pseudoLegalMoves(pos, p.square)
		for _, m := range plms {
			ud := pos.Make(m)
			if pos.isSafe(pos.kings[pos.activeColor.opposite()].square, pos.activeColor) {
				ms = append(ms, m)
			}
			pos.unmake(ud)
		}
	}
	return ms
}

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
		if pos.enPassantSquare.rank != enPassantRanks[pos.activeColor.opposite()] {
			return fmt.Errorf("its %v's move and en passant filled so it must be %d", pos.activeColor, enPassantRanks[pos.activeColor])
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

	for _, c := range colors {
		var atKingSq *piece
		for i, r := range pos.castleRights[c] {
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

func (pos *Position) Make(m Move) undoInfo {
	if debug {
		if err := pos.Validate(); err != nil {
			panicf(pos, "invalid position: %v", err)
		}
	}
	if debug && !m.from.isValid() {
		panicf(pos, "illegal from sq %+v", m)
	}
	if debug && !m.to.isValid() {
		panicf(pos, "illegal to sq %+v", m)
	}
	if debug && m.promotion == pawn {
		panicf(pos, "illegal promotion %+v", m)
	}

	p := pos.pieceAt(m.from)
	if debug {
		if p == nil {
			panicf(pos, "no piece at from sq %+v", m)
		} else if p.color != black && p.color != white {
			panicf(pos, "piece of unexpected color %v at from sq %+v", p.color, m)
		} else if p.color != pos.activeColor {
			panicf(pos, "piece of inactive color at from sq %+v", m)
		}
	}
	ud := undoInfo{
		castleRights:    pos.castleRights,
		enPassantSquare: pos.enPassantSquare,
		halfmoveClock:   pos.halfmoveClock,
	}
	ud.movedPiece.from = m.from
	ud.movedPiece.piece = p
	ud.movedPiece.promotion = m.promotion
	isPawnMove := p.kind == pawn

	atToSq := pos.pieceAt(m.to)
	if debug && atToSq != nil {
		if atToSq.color != black && atToSq.color != white {
			panicf(pos, "piece of unexpected color %v at to sq %+v", atToSq.color, m)
		}
		if atToSq.color == pos.activeColor {
			panicf(pos, "piece of active color at to sq %+v", m)
		}
		if atToSq.kind == king {
			panicf(pos, "move %+v trying to take a king", m)
		}
	}
	ud.capturedPiece = atToSq

	// move itself
	castleDist := m.to.file - m.from.file
	if p.kind == king && (castleDist == 2 || castleDist == -2) {
		if debug && (m.from != initKingSq[p.color]) {
			panicf(pos, "no king at init sq %+v", initKingSq[p.color])
		}
		side := 0
		if castleDist == -2 {
			side = 1
		}
		if debug && !pos.castleRights[p.color][side] {
			panicf(pos, "no castle right for %v %s, but tried", p.color, castleSide[side])
		}
		rookSq := initRookSqs[p.color][side]
		atRookSq := pos.pieceAt(rookSq)
		dir := int8(1 - 2*side)
		if debug {
			if atRookSq == nil || atRookSq.kind != rook || atRookSq.color != p.color {
				panicf(pos, "no rook of king color at rook sq %+v during castle move %+v", rookSq, m)
			}
			for i := range int8(2) {
				onPathSq := square{file: p.square.file + i*dir, rank: p.square.rank}
				if !pos.isSafe(onPathSq, p.color.opposite()) {
					panicf(pos, "sq %+v on castle path and it is not safe", onPathSq)
				}
			}
		}
		pos.board[rookSq.file][rookSq.rank] = nil
		atRookSq.square = square{file: p.square.file + 1*dir, rank: rookSq.rank}
		pos.board[atRookSq.square.file][atRookSq.square.rank] = atRookSq
		ud.castleRook.rook = atRookSq
		ud.castleRook.from = rookSq
	} else if p.kind == pawn && m.to == pos.enPassantSquare && pos.enPassantSquare != noEnPassantSquare {
		aboveEnPassantSq := square{file: m.to.file, rank: m.from.rank}
		aboveEnPassant := pos.pieceAt(aboveEnPassantSq)
		if debug {
			if aboveEnPassant == nil || aboveEnPassant.kind != pawn || aboveEnPassant.color == p.color {
				panicf(pos, "piece on above en passant sq %+v is not a pawn of %v color", aboveEnPassantSq, p.color.opposite())
			}
			if pos.enPassantSquare.rank != enPassantRanks[p.color.opposite()] {
				panicf(pos, "en passant sq %+v rank do not match with current move color (%v)", pos.enPassantSquare, pos.activeColor)
			}
		}
		aboveEnPassant.captured = true
		ud.capturedPiece = aboveEnPassant
		pos.board[aboveEnPassantSq.file][aboveEnPassantSq.rank] = nil
	}
	if atToSq != nil {
		atToSq.captured = true
	}
	if m.promotion != noPromotion {
		p.kind = m.promotion
	}
	pos.board[m.from.file][m.from.rank] = nil
	p.square = m.to
	pos.board[p.square.file][p.square.rank] = p

	// castle rights
	if p.kind == king {
		pos.castleRights[p.color] = [2]bool{}
	}
	for _, c := range colors {
		for i, rookSq := range initRookSqs[c] {
			if m.from == rookSq || m.to == rookSq {
				pos.castleRights[c][i] = false
			}
		}
	}

	// en passant
	forward := pawnForward[p.color]
	if p.kind == pawn && forward*(m.to.rank-m.from.rank) == 2 {
		pos.enPassantSquare = square{
			rank: m.to.rank - 1*forward,
			file: m.to.file,
		}
	} else {
		pos.enPassantSquare = noEnPassantSquare
	}

	// fullmove count
	if pos.activeColor == black {
		pos.fullmoveCnt += 1
	}

	// active color
	pos.activeColor = pos.activeColor.opposite()

	// halfmove clock
	if isPawnMove || ud.capturedPiece != nil {
		pos.halfmoveClock = 0
	} else {
		pos.halfmoveClock++
	}

	return ud
}

func (pos *Position) unmake(ud undoInfo) {
	pos.castleRights = ud.castleRights
	pos.enPassantSquare = ud.enPassantSquare
	pos.halfmoveClock = ud.halfmoveClock
	if pos.activeColor == black.opposite() {
		pos.fullmoveCnt -= 1
	}
	pos.activeColor = pos.activeColor.opposite()

	if ud.castleRook.rook != nil {
		pos.board[ud.castleRook.rook.square.file][ud.castleRook.rook.square.rank] = nil
		ud.castleRook.rook.square = ud.castleRook.from
		pos.board[ud.castleRook.from.file][ud.castleRook.from.rank] = ud.castleRook.rook
	}

	pos.board[ud.movedPiece.piece.square.file][ud.movedPiece.piece.square.rank] = nil
	ud.movedPiece.piece.square = ud.movedPiece.from
	pos.board[ud.movedPiece.from.file][ud.movedPiece.from.rank] = ud.movedPiece.piece
	if ud.movedPiece.promotion != noPromotion {
		ud.movedPiece.piece.kind = pawn
	}

	if ud.capturedPiece != nil {
		ud.capturedPiece.captured = false
		pos.board[ud.capturedPiece.square.file][ud.capturedPiece.square.rank] = ud.capturedPiece
	}
}
