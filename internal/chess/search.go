package chess

import (
	"math"
)

func (pos *Position) scoredMoves() ([]Move, []int32) {
	var plms []Move
	var mvvlvas []int32
	for _, p := range pos.pieces {
		if p == nil || p.color != pos.activeColor || p.captured {
			continue
		}
		for _, m := range pseudoLegalMoves(pos, p.square) {
			a := pos.pieceAt(m.from)
			v := pos.pieceAt(m.to)

			mvvlva := int32(math.MaxInt32)
			if v != nil {
				mvvlva = pieceValue[a.kind] - 10*pieceValue[v.kind]
			} else if a.kind == pawn && pos.enPassantSquare != noEnPassantSquare && m.to == pos.enPassantSquare {
				mvvlva = pieceValue[a.kind] - 10*pieceValue[pawn]
			}
			plms = append(plms, m)
			mvvlvas = append(mvvlvas, mvvlva)
		}
	}
	return plms, mvvlvas
}

func pickNext[T any](arr []T, scores []int32, start int) T {
	best := start
	for i := start + 1; i < len(arr); i++ {
		if scores[i] < scores[best] {
			best = i
		}
	}
	arr[start], arr[best] = arr[best], arr[start]
	scores[start], scores[best] = scores[best], scores[start]

	return arr[start]
}

func (pos *Position) BestMove(depth int32) Move {
	moveColor := pos.activeColor
	var r struct {
		move Move
		eval int32
	}
	if moveColor == white {
		r.eval = math.MinInt32
	} else {
		r.eval = math.MaxInt32
	}

	plms, mvvlvas := pos.scoredMoves()
	var m Move
	for i := range plms {
		m = pickNext(plms, mvvlvas, i)

		ud := pos.Make(m)
		if !pos.isSafe(pos.kings[moveColor].square, moveColor.opposite()) {
			pos.unmake(ud)
			continue
		}
		if moveColor == white {
			eval := pos.alphaBeta(depth-1, 1, r.eval, math.MaxInt32)
			if r.move == NoMove || eval > r.eval {
				r.eval = eval
				r.move = m
			}
		} else {
			eval := pos.alphaBeta(depth-1, 1, math.MinInt32, r.eval)
			if r.move == NoMove || eval < r.eval {
				r.eval = eval
				r.move = m
			}
		}

		pos.unmake(ud)
	}

	return r.move
}

func (pos *Position) alphaBeta(depth int32, ply int32, alpha int32, beta int32) int32 {
	moveColor := pos.activeColor

	// 50-halfmove = draw by chess rules
	if pos.halfmoveClock > 100 {
		return 0
	}
	if depth == 0 {
		inCheck := !pos.isSafe(pos.kings[moveColor].square, moveColor.opposite())
		if !inCheck {
			return pos.evaluation()
		}
		depth++
	}

	var r int32
	if moveColor == white {
		r = math.MinInt32
	} else {
		r = math.MaxInt32
	}

	plms, mvvlvas := pos.scoredMoves()

	var (
		mvCnt int
		m     Move
	)
	for i := range plms {
		m = pickNext(plms, mvvlvas, i)

		ud := pos.Make(m)
		if !pos.isSafe(pos.kings[moveColor].square, moveColor.opposite()) {
			pos.unmake(ud)
			continue
		}
		mvCnt++

		eval := pos.alphaBeta(depth-1, ply+1, alpha, beta)
		if moveColor == white {
			r = max(r, eval)
			alpha = max(alpha, eval)
			if r >= beta {
				pos.unmake(ud)
				break
			}
		} else {
			r = min(r, eval)
			beta = min(beta, eval)
			if r <= alpha {
				pos.unmake(ud)
				break
			}
		}

		pos.unmake(ud)
	}

	if mvCnt == 0 {
		if pos.isSafe(pos.kings[moveColor].square, moveColor.opposite()) {
			return 0
		}
		if moveColor == white {
			return math.MinInt32 + ply
		}
		return math.MaxInt32 - ply
	}

	return r
}
