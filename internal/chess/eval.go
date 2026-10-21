package chess

var pieceValue = [6]int32{
	pawn:   1,
	knight: 3,
	bishop: 3,
	rook:   5,
	queen:  9,
}

func (pos *Position) evaluation() int32 {
	var r int32
	for _, p := range pos.pieces {
		if p == nil || p.captured || p.kind == king {
			continue
		}
		r += pieceValue[p.kind] * (1 - 2*int32(p.color))
	}
	return r
}
