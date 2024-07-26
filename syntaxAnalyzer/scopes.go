package syntaxAnalyzer

import "github.com/DanielNos/NeCo/lexer"

func (sn *SyntaxAnalyzer) analyzeScope() {
	sn.consume()
	sn.analyzeStatementList(true)
	sn.consume()
}

func (sn *SyntaxAnalyzer) analyzeStatementList(isScope bool) {
	start := sn.peekPrevious()

	for sn.peek().TokenType != lexer.TT_EndOfFile {
		if sn.analyzeStatement(isScope) {
			return
		}
	}

	if isScope {
		sn.newError(start, "Scope is missing a closing brace.")
	}
}
