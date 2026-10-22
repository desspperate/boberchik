package chess

import "math"

const searchDepth = 7

func (pos *Position) BestMove() move {
	moveColor := pos.activeColor
	var r struct {
		move move
		eval int32
	}
	if moveColor == white {
		r.eval = math.MinInt32
	} else {
		r.eval = math.MaxInt32
	}

	for _, p := range pos.pieces {
		if p == nil || p.color != pos.activeColor || p.captured {
			continue
		}
		plms := pseudoLegalMoves(pos, p.square)

		for _, m := range plms {
			ud := pos.Make(m)
			if !pos.isSafe(pos.kings[moveColor].square, moveColor.opposite()) {
				pos.unmake(ud)
				continue
			}
			if moveColor == white {
				eval := pos.alphaBeta(searchDepth-1, 1, r.eval, math.MaxInt32)
				if r.move == NoMove || eval > r.eval {
					r.eval = eval
					r.move = m
				}
			} else {
				eval := pos.alphaBeta(searchDepth-1, 1, math.MinInt32, r.eval)
				if r.move == NoMove || eval < r.eval {
					r.eval = eval
					r.move = m
				}
			}

			pos.unmake(ud)
		}
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

	var mvCnt int
	for _, p := range pos.pieces {
		if p == nil || p.color != pos.activeColor || p.captured {
			continue
		}
		plms := pseudoLegalMoves(pos, p.square)

		flag := false
		for _, m := range plms {
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
					flag = true
					pos.unmake(ud)
					break
				}
			} else {
				r = min(r, eval)
				beta = min(beta, eval)
				if r <= alpha {
					flag = true
					pos.unmake(ud)
					break
				}
			}

			pos.unmake(ud)
		}
		if flag {
			break
		}
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
