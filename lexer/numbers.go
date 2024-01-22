package lexer

import (
	"fmt"
	"strconv"
	"unicode"
)

var DIGIT_VALUE = map[rune]int{
	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
	'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	'a': 10, 'A': 10, 'b': 11, 'B': 11,
	'c': 12, 'C': 12, 'd': 13, 'D': 13,
	'e': 14, 'E': 14, 'f': 15, 'F': 15,
	'g': 16, 'G': 16, 'h': 17, 'H': 17,
	'i': 18, 'I': 18, 'j': 19, 'J': 19,
	'k': 20, 'K': 20, 'l': 21, 'L': 21,
	'm': 22, 'M': 22, 'n': 23, 'N': 23,
	'o': 24, 'O': 24, 'p': 25, 'P': 25,
	'q': 26, 'Q': 26, 'r': 27, 'R': 27,
	's': 28, 'S': 28, 't': 29, 'T': 29,
	'u': 30, 'U': 30, 'v': 31, 'V': 31,
	'w': 32, 'W': 32, 'x': 33, 'X': 33,
	'y': 34, 'Y': 34, 'z': 35, 'Z': 35,
}

func (l *Lexer) lexNumber() {
	startLine := l.lineIndex
	startChar := l.charIndex

	// Collect number/base
	var base string

	for i := 0; i < 2; i++ {
		// Digit
		if unicode.IsDigit(l.currRune) {
			l.token.WriteRune(l.currRune)
			l.advance()
			// Space
		} else if l.currRune == '_' {
			l.advance()
			// Base end
		} else if l.currRune == 'x' {
			base = l.token.String()
			l.token.Reset()
			l.advance()
			break
			// Float
		} else if l.currRune == '.' {
			l.lexFloat(startLine, startChar)
			return
			// End of number
		} else if isTokenBreaker(l.currRune) {
			break
			// Invalid character
		} else {
			l.collectRestOfToken()
			l.newError(startLine, startChar, true, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
			l.newToken(startLine, startChar, TT_LT_Int)
			return
		}
	}

	if base != "" {
		l.lexBaseInt(startLine, startChar, base)
		return
	}

	// Collect number
	for unicode.IsDigit(l.currRune) || l.currRune == '_' {
		if l.currRune != '_' {
			l.token.WriteRune(l.currRune)
		}
		l.advance()
	}

	// Create token
	if isTokenBreaker(l.currRune) {
		l.newToken(startLine, startChar, TT_LT_Int)
		return
		// Float
	} else if l.currRune == '.' {
		l.lexFloat(startLine, startChar)
		return
		// Invalid characters in number
	} else {
		l.collectRestOfToken()
		l.newError(startLine, startChar, true, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
	}
}

func (l *Lexer) lexBaseInt(startLine, startChar uint, baseString string) {
	// Convert base to int
	base, _ := strconv.Atoi(baseString)

	// Invalid base
	if base < 2 || base > 36 {
		l.collectRestOfToken()
		l.newError(startLine, startChar, true, fmt.Sprintf("Invalid integer base %d. Only bases in range <2, 36> are supported.", base))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Collect number
	digitValue, valid := DIGIT_VALUE[l.currRune]
	invalidDigits := false
	for valid || l.currRune == '_' {
		if l.currRune != '_' {
			if digitValue >= base {
				invalidDigits = true
			}
			l.token.WriteRune(unicode.ToLower(l.currRune))
		}
		l.advance()
		digitValue, valid = DIGIT_VALUE[l.currRune]
	}

	// Digits exceed base
	if invalidDigits {
		l.newError(startLine, startChar+uint(len(baseString))+1, true, fmt.Sprintf("Digit/s of integer \"%s\" exceed its base.", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Invalid characters in number
	if !isTokenBreaker(l.currRune) {
		l.collectRestOfToken()
		l.newError(startLine, startChar, true, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Convert and create token
	value, _ := strconv.ParseInt(l.token.String(), base, 64)
	l.token.Reset()

	l.newTokenFrom(startLine, startChar, TT_LT_Int, fmt.Sprintf("%d", value))
}

func (l *Lexer) lexFloat(startLine, startChar uint) {
	l.token.WriteRune(l.currRune)
	l.advance()

	// Collect rest of float
	for unicode.IsDigit(l.currRune) || l.currRune == '_' {
		if l.currRune != '_' {
			l.token.WriteRune(l.currRune)
		}
		l.advance()
	}

	// Invalid characters
	if !isTokenBreaker(l.currRune) {
		l.collectRestOfToken()
		l.newError(startLine, startChar, true, fmt.Sprintf("Invalid character/s in float literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Float)
		return
	}

	l.newToken(startLine, startChar, TT_LT_Float)
}
