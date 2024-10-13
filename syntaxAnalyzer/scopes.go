package syntaxAnalyzer

func (sn *SyntaxAnalyzer) analyzeScope() {
	sn.consume()
	sn.analyzeStatementList(true)
	sn.consume()
}

func (sn *SyntaxAnalyzer) analyzeStatementList(isScope bool) {
	start := sn.peekPrevious()

	for !sn.peek().IsEndOfFileOf(sn.tokens[0]) {
		if sn.analyzeStatement(isScope) {
			return
		}
	}

	if isScope {
		sn.newError(start, "Scope is missing a closing brace.")
	}
}
