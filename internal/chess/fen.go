package chess

import (
	"errors"
	"fmt"
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

var enPassantLetterSegmentAllowedChars = map[rune]bool{
	'a': true,
	'b': true,
	'c': true,
	'd': true,
	'e': true,
	'f': true,
	'g': true,
	'h': true,
}

const maxHalfmoveClock = 150

const minFullmoveCnt = 1
const maxFullmoveCnt = 8849 // https://tomrocksmaths.com/wp-content/uploads/2026/06/longestchessgame-katie-cam.pdf

type pieceDesc struct {
	kind  kind
	color color
}

var pieceByChar = map[rune]pieceDesc{
	'K': {king, white}, 'Q': {queen, white}, 'R': {rook, white},
	'B': {bishop, white}, 'N': {knight, white}, 'P': {pawn, white},
	'k': {king, black}, 'q': {queen, black}, 'r': {rook, black},
	'b': {bishop, black}, 'n': {knight, black}, 'p': {pawn, black},
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

	// activeColor
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
	for _, c := range castlingSegment {
		switch c {
		case 'K':
			if position.whiteCastleKing {
				return Position{}, errors.New("castling segment contains 'K' more than once")
			}
			if position.whiteCastleQueen || position.blackCastleKing || position.blackCastleQueen {
				return Position{}, errors.New("'K' is not in a right order, right order for castling segment \"KQkq\"")
			}
			position.whiteCastleKing = true
		case 'Q':
			if position.whiteCastleQueen {
				return Position{}, errors.New("castling segment contains 'Q' more than once")
			}
			if position.blackCastleKing || position.blackCastleQueen {
				return Position{}, errors.New("'Q' is not in a right order, right order for castling segment \"KQkq\"")
			}
			position.whiteCastleQueen = true
		case 'k':
			if position.blackCastleKing {
				return Position{}, errors.New("castling segment contains 'k' more than once")
			}
			if position.blackCastleQueen {
				return Position{}, errors.New("'k' is not in a right order, right order for castling segment \"KQkq\"")
			}
			position.blackCastleKing = true
		case 'q':
			if position.blackCastleQueen {
				return Position{}, errors.New("castling segment contains 'q' more than once")
			}
			position.blackCastleQueen = true
		}
	}

	// enPassant
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
		if !enPassantLetterSegmentAllowedChars[rune(letterSegment)] {
			return Position{}, errors.New("enPassant letter segment contains invalid characters")
		}
		switch numberSegment {
		case '3':
			if position.activeColor == white {
				return Position{}, errors.New("last move was black, so enPassant rank must be 6")
			}
		case '6':
			if position.activeColor == black {
				return Position{}, errors.New("last move was white, so enPassant rank must be 3")
			}
		default:
			return Position{}, errors.New("enPassant number segment contains invalid characters")
		}
		position.enPassantSquare = square{
			file: letterSegment - 'a',
			rank: '8' - numberSegment,
		}
	}

	// halfmoveClock
	halfmoveClockNum, err := strconv.ParseUint(halfmoveClockSegment, 10, 8)
	if err != nil || halfmoveClockNum > maxHalfmoveClock {
		return Position{}, fmt.Errorf("halfmoveClock out of bounds 0-%d", maxHalfmoveClock)
	}
	position.halfmoveClock = uint8(halfmoveClockNum)

	// fullmoveCnt
	fullmoveCntNum, err := strconv.ParseUint(fullmoveCntSegment, 10, 16)
	if err != nil || fullmoveCntNum < minFullmoveCnt || fullmoveCntNum > maxFullmoveCnt {
		return Position{}, fmt.Errorf("fullmoveCnt out of bounds %d-%d", minFullmoveCnt, maxFullmoveCnt)
	}
	position.fullmoveCnt = uint16(fullmoveCntNum)

	// placement
	ranks := strings.Split(placementSegment, "/")
	if len(ranks) != Ranks {
		return Position{}, fmt.Errorf("placement segment must be split on %d ranks by '/'", Ranks)
	}
	for i, rank := range ranks {
		var described int32 = 0

		prevSkip := false
		for _, c := range rank {
			if p, ok := pieceByChar[c]; ok {
				pieceSquare := square{
					file: uint8(described),
					rank: uint8(i),
				}
				puttingErr := position.putPiece(
					&piece{
						kind:   p.kind,
						color:  p.color,
						square: pieceSquare,
					},
					pieceSquare,
				)
				if puttingErr != nil {
					return Position{}, fmt.Errorf("cannot place a piece on a [%d;%d]: %w", described, i, puttingErr)
				}
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

func EncodeFen(position Position) (string, error) {
	return "", errors.New("not implemented")
}
