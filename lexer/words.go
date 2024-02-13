package lexer

import "unicode"

func (l *Lexer) lexLetter() {
	startLine := l.lineIndex
	startChar := l.charIndex

	// Collect identifier/keyword
	l.token.WriteRune(l.currRune)
	l.advance()

	for unicode.IsLetter(l.currRune) || unicode.IsDigit(l.currRune) || l.currRune == '_' || l.currRune == '.' {
		l.token.WriteRune(l.currRune)
		l.advance()
	}

	// Check if token is a keyword
	value := l.token.String()
	keyword, isKeyword := KEYWORDS[value]

	// Identifier/Literal
	if !isKeyword {
		if value == "true" { // Literal true
			l.newTokenFrom(startLine, startChar, TT_LT_Bool, "1")
			l.token.Reset()
		} else if value == "false" { // Literal false
			l.newTokenFrom(startLine, startChar, TT_LT_Bool, "0")
			l.token.Reset()
		} else { // Identifier
			l.newToken(startLine, startChar, TT_Identifier)
		}
		// Keyword
	} else {
		l.newTokenFrom(startLine, startChar, keyword, "")
		l.token.Reset()
	}
}

func (l *Lexer) lexString() {
	startLine := l.lineIndex
	startChar := l.charIndex
	l.advance()

	// Collect string
	for l.currRune != '"' {
		if l.currRune == '\r' {
			l.newError(l.lineIndex, l.charIndex, false, "Multi-line strings are not allowed.")
			l.advance()
			l.advance()
			l.lineIndex++
			l.charIndex = 1
		} else if l.currRune == '\n' {
			l.newError(l.lineIndex, l.charIndex, false, "Multi-line strings are not allowed.")
			l.advance()
			l.lineIndex++
			l.charIndex = 1
		} else {
			l.token.WriteRune(l.currRune)
			l.advance()
		}
	}
	l.advance()

	l.newToken(startLine, startChar, TT_LT_String)
}
