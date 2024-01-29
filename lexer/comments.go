package lexer

func (l *Lexer) skipComment() {
	for l.currRune != '\n' && l.currRune != '\r' {
		l.advance()
	}
}

func (l *Lexer) skipMultiLineComment() {
	for l.currRune != EOF {
		switch l.currRune {

		// End of comment
		case '*':
			l.advance()
			if l.currRune == '/' {
				l.advance()
				return
			}

		// Start of new multiline comment
		case '/':
			l.advance()
			if l.currRune == '*' {
				l.advance()
				l.skipMultiLineComment()
			}

		// New line
		case '\n':
			l.lineIndex++
			l.charIndex = 1
			l.advance()

		// Windows new line
		case '\r':
			l.advance()
			if l.currRune != '\n' {
				l.newError(l.lineIndex, l.charIndex-1, true, "Invalid Windows line ending.")
			} else {
				l.advance()
			}

			l.lineIndex++
			l.charIndex = 1

		default:
			l.advance()
		}
	}
}
