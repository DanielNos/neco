package lexer

func (l *Lexer) skipComment() {
	for l.currRune != '\n' && l.currRune != '\r' {
		l.advance()
	}
}

func (l *Lexer) skipMultiLineComment() {
	for l.currRune != EOF {
		switch l.currRune {

		case '*': // End of comment
			l.advance()
			if l.currRune == '/' {
				l.advance()
				return
			}

		case '/': // Start of new multiline comment
			l.advance()
			if l.currRune == '*' {
				l.advance()
				l.skipMultiLineComment()
			}

		case '\n': // New line
			l.lineIndex++
			l.charIndex = 1
			l.advance()

		case '\r': // Windows new line
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
