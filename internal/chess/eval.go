package chess

var pieceValue = [6]int32{
	king:   0,
	pawn:   100,
	knight: 330,
	bishop: 350,
	rook:   500,
	queen:  900,
}

func (pos *Position) evaluation() int32 {
	var r int32
	for _, p := range pos.pieces {
		if p == nil || p.captured || p.kind == king {
			continue
		}
		rank := p.square.rank
		if p.color == black {
			rank = 7 - rank
		}
		r += (pieceValue[p.kind] + pst[p.kind][p.square.file][rank]) * (1 - 2*int32(p.color))
	}
	return r
}
