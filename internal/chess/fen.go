package chess

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const FenSegments = 6

type segmentType byte

const (
	placement segmentType = iota
	activeColor
	castling
	enPassant
	halfmoveClock
	fullmoveCnt
)

var segmentAllowedChars = [FenSegments]map[rune]bool{
	placement: {
		'P': true,
		'N': true,
		'B': true,
		'R': true,
		'Q': true,
		'K': true,
		'p': true,
		'n': true,
		'b': true,
		'r': true,
		'q': true,
		'k': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'/': true,
	},
	activeColor: {
		'w': true,
		'b': true,
	},
	castling: {
		'K': true,
		'Q': true,
		'k': true,
		'q': true,
		'-': true,
	},
	enPassant: {
		'-': true,
		'a': true,
		'b': true,
		'c': true,
		'd': true,
		'e': true,
		'f': true,
		'g': true,
		'h': true,
		'3': true,
		'6': true,
	},
	halfmoveClock: {
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
	},
	fullmoveCnt: {
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
	},
}
var segmentSizeBorders = [FenSegments][2]int{
	placement:     {15, 71},
	activeColor:   {1, 1},
	castling:      {1, 4},
	enPassant:     {1, 2},
	halfmoveClock: {1, 3},
	fullmoveCnt:   {1, 4},
}

type pieceDesc struct {
	kind  pieceKind
	color color
}

var pieceByChar = map[rune]pieceDesc{
	'K': {king, white}, 'Q': {queen, white}, 'R': {rook, white},
	'B': {bishop, white}, 'N': {knight, white}, 'P': {pawn, white},
	'k': {king, black}, 'q': {queen, black}, 'r': {rook, black},
	'b': {bishop, black}, 'n': {knight, black}, 'p': {pawn, black},
}

var charByPiece = map[pieceDesc]rune{
	{king, white}: 'K', {queen, white}: 'Q', {rook, white}: 'R',
	{bishop, white}: 'B', {knight, white}: 'N', {pawn, white}: 'P',
	{king, black}: 'k', {queen, black}: 'q', {rook, black}: 'r',
	{bishop, black}: 'b', {knight, black}: 'n', {pawn, black}: 'p',
}

func DecodeFen(fen string) (Position, error) {
	splitFen := strings.Split(fen, " ")
	if len(splitFen) != FenSegments {
		return Position{}, fmt.Errorf("input have %d segments, but must have %d", len(splitFen), FenSegments)
	}

	for i, s := range splitFen {
		allowedChars := segmentAllowedChars[i]
		minSegmentLen, maxSegmentLen := segmentSizeBorders[i][0], segmentSizeBorders[i][1]
		if len(s) > maxSegmentLen || len(s) < minSegmentLen {
			return Position{}, fmt.Errorf("segment %d is not of the right size %d, allowed size %d-%d", i, len(s), minSegmentLen, maxSegmentLen)
		}
		for _, c := range s {
			if !allowedChars[c] {
				return Position{}, fmt.Errorf("input contains invalid character '%c' in segment %d", c, i)
			}
		}
	}

	placementSegment := splitFen[placement]
	activeColorSegment := splitFen[activeColor]
	castlingSegment := splitFen[castling]
	enPassantSegment := splitFen[enPassant]
	halfmoveClockSegment := splitFen[halfmoveClock]
	fullmoveCntSegment := splitFen[fullmoveCnt]

	position := Position{}

	// active color
	switch activeColorSegment[0] {
	case 'w':
		position.activeColor = white
	case 'b':
		position.activeColor = black
	}

	// castling
	if strings.IndexRune(castlingSegment, '-') != -1 {
		if len(castlingSegment) > 1 {
			return Position{}, errors.New("castling segment contains '-', in that case that must be the only character")
		}
	}
	for _, ch := range castlingSegment {
		for _, c := range colors {
			for i, letter := range castleLetters[c] {
				if ch != letter {
					continue
				}
				if position.castleRights[c][i] {
					return Position{}, fmt.Errorf("castling segment contains '%c' more than once", ch)
				}
				for _, laterC := range colors {
					for laterI := range position.castleRights[laterC] {
						if laterC*2+color(laterI) > c*2+color(i) && position.castleRights[laterC][laterI] {
							return Position{}, fmt.Errorf("'%c' is not in a right order, right order for castling segment \"KQkq\"", ch)
						}
					}
				}
				position.castleRights[c][i] = true
			}
		}
	}

	// en passant
	if strings.IndexRune(enPassantSegment, '-') != -1 {
		if len(enPassantSegment) > 1 {
			return Position{}, errors.New("enPassant segment contains '-', in that case that must be the only character")
		}
	} else {
		if len(enPassantSegment) == 1 {
			return Position{}, errors.New("enPassantSegment did not contain '-', so its length must be 2")
		}
		letterSegment := enPassantSegment[0]
		numberSegment := enPassantSegment[1]

		if letterSegment < 'a' || letterSegment > 'h' {
			return Position{}, errors.New("enPassant letter segment contains invalid characters")
		}
		if numberSegment != '3' && numberSegment != '6' {
			return Position{}, errors.New("enPassant number segment contains invalid characters")
		}

		position.enPassantSquare = square{
			file: int8(letterSegment - 'a'),
			rank: int8('8' - numberSegment),
		}
	}

	// halfmove clock
	halfmoveClockNum, err := strconv.ParseUint(halfmoveClockSegment, 10, 8)
	if err != nil {
		return Position{}, fmt.Errorf("halfmoveClock out of uint8 range 0-%d", math.MaxUint8)
	}
	position.halfmoveClock = uint8(halfmoveClockNum)

	// fullmove count
	fullmoveCntNum, err := strconv.ParseUint(fullmoveCntSegment, 10, 16)
	if err != nil {
		return Position{}, fmt.Errorf("fullmoveCnt out of uint16 range 0-%d", math.MaxUint16)
	}
	position.fullmoveCnt = uint16(fullmoveCntNum)

	// placement
	ranks := strings.Split(placementSegment, "/")
	if len(ranks) != Ranks {
		return Position{}, fmt.Errorf("placement segment must be split on %d ranks by '/'", Ranks)
	}
	var pieces int32
	for i, rank := range ranks {
		var described int32

		prevSkip := false
		for _, c := range rank {
			if p, ok := pieceByChar[c]; ok {
				pieces++
				if pieces > maxPieces {
					return Position{}, fmt.Errorf("FEN describes more than %d pieces", maxPieces)
				}
				if described >= Files {
					return Position{}, fmt.Errorf("rank %d describes more than %d squares", i, Files)
				}
				pieceSquare := square{
					file: int8(described),
					rank: int8(i),
				}
				position.putPiece(
					&piece{
						kind:   p.kind,
						color:  p.color,
						square: pieceSquare,
					},
					pieceSquare,
				)
				described++
				prevSkip = false

			} else {
				if prevSkip {
					return Position{}, fmt.Errorf("rank %d must not contain 2 or more skip numbers in a row", i)
				}
				described += c - '0'
				prevSkip = true
			}
		}

		if described != Files {
			return Position{}, fmt.Errorf("rank %d must describe exactly %d squares", i, Files)
		}
	}

	return position, nil
}

var noEnPassantSquare = square{0, 0}

func (pos *Position) Fen() string {
	// placement
	ranks := [Ranks]string{}
	for rank := range int8(Ranks) {
		empty := 0
		for file := range int8(Files) {
			p := pos.pieceAt(square{file: file, rank: rank})
			if p == nil {
				empty++
				continue
			}
			if empty > 0 {
				ranks[rank] += string(rune(empty + '0'))
				empty = 0
			}
			ranks[rank] += string(charByPiece[pieceDesc{color: p.color, kind: p.kind}])
		}
		if empty > 0 {
			ranks[rank] += string(rune(empty + '0'))
		}
	}
	placementSegment := strings.Join(ranks[:], "/")

	// active color
	activeColorSegment := ""
	if pos.activeColor == white {
		activeColorSegment = "w"
	} else if pos.activeColor == black {
		activeColorSegment = "b"
	}

	// castling
	castlingSegment := ""
	for _, c := range colors {
		for i, r := range pos.castleRights[c] {
			if r {
				castlingSegment += string(castleLetters[c][i])
			}
		}
	}
	if castlingSegment == "" {
		castlingSegment = "-"
	}

	// en passant
	enPassantSegment := "-"
	if pos.enPassantSquare != noEnPassantSquare {
		enPassantLetter := uint8(pos.enPassantSquare.file) + 'a'
		enPassantNumber := '8' - uint8(pos.enPassantSquare.rank)
		enPassantSegment = string(enPassantLetter) + string(enPassantNumber)
	}

	// fullmove count
	fullmoveCntSegment := strconv.Itoa(int(pos.fullmoveCnt))

	// halfmove clock
	halfmoveClockSegment := strconv.Itoa(int(pos.halfmoveClock))

	return strings.Join(
		[]string{
			placementSegment,
			activeColorSegment,
			castlingSegment,
			enPassantSegment,
			halfmoveClockSegment,
			fullmoveCntSegment,
		}, " ",
	)
}
