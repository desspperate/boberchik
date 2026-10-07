package chess

import (
	"errors"
	"fmt"
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

const enPassantSegments = 2

type enPassantSegmentType byte

const (
	letter enPassantSegmentType = iota
	number
)

var enPassantSegmentAllowedChars = [enPassantSegments]map[rune]bool{
	letter: {
		'a': true,
		'b': true,
		'c': true,
		'd': true,
		'e': true,
		'f': true,
		'g': true,
		'h': true,
	},
	number: {
		'3': true,
		'6': true,
	},
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

	if activeColorSegment[0] == 'w' {
		position.activeColor = white
	}
	if activeColorSegment[0] == 'b' {
		position.activeColor = black
	}

	if strings.IndexRune(castlingSegment, '-') != -1 {
		if len(castlingSegment) > 1 {
			return Position{}, errors.New("castling segment contains '-', in that case that must be the only character")
		}
	}
	for _, c := range castlingSegment {
		switch c {
		case 'K':
			if position.whiteCastleKing {
				return Position{}, errors.New("castling segment containts 'K' more than once")
			}
			position.whiteCastleKing = true
		case 'Q':
			if position.whiteCastleQueen {
				return Position{}, errors.New("castling segment contains 'Q' more than once")
			}
			position.whiteCastleQueen = true
		case 'k':
			if position.blackCastleKing {
				return Position{}, errors.New("castling segment contains 'k' more than once")
			}
			position.blackCastleKing = true
		case 'q':
			if position.blackCastleQueen {
				return Position{}, errors.New("castling segment contains 'q' more than once")
			}
			position.blackCastleQueen = true
		}
	}

	if strings.IndexRune(enPassantSegment, '-') != -1 {
		if len(enPassantSegment) > 1 {
			return Position{}, errors.New("enPassant segment contains '-', in that case that must be the only character")
		}
	} else {
		if len(enPassantSegment) == 1 {
			return Position{}, errors.New("enPassantSegment did not contain '-', so its length must be 2")
		}
	}
	letterSegment := enPassantSegment[0]
	numberSegment := enPassantSegment[1]
	if !enPassantSegmentAllowedChars[letter][rune(letterSegment)] {
		return Position{}, errors.New("enPassant letter segment contains invalid characters")
	}
	if !enPassantSegmentAllowedChars[number][rune(numberSegment)] {
		return Position{}, errors.New("enPassant number segment contains invalid characters")
	}
	position.enPassantSquare = square{1, 1}

	return position, nil
}

func EncodeFen(position Position) (string, error) {
	return "", errors.New("not implemented")
}
