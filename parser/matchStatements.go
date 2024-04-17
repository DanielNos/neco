package parser

import "neco/lexer"

func (p *Parser) parseMatch() *Node {
	startPosition := p.consume().Position

	expression := p.parseExpressionRoot()
	cases := []*Node{}

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	p.consume() // {
	var defaultCase *Node = nil
	caseCount := 0

	// Collect cases
	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Skip empty lines
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
			continue
		}

		if p.peek().TokenType == lexer.TT_KW_default {
			p.consume() // default
			p.consume() // :

			defaultCase = p.parseStatement(false)

			p.scopeCounter++
			// Parse case
		} else {
			// Default case has to be the last case
			if defaultCase != nil {
				p.newError(defaultCase.Position, "Default case has to be the last case.")
			}

			caseCount++
			casePosition := p.peek().Position

			// Collect case expressions
			expressions := []*Node{}
			expressions = append(expressions, p.parseExpressionRoot())

			for p.peek().TokenType == lexer.TT_DL_Comma {
				p.consume()
				expressions = append(expressions, p.parseExpressionRoot())
				caseCount++
			}

			p.consume()

			cases = append(cases, &Node{casePosition, NT_Case, &CaseNode{expressions, p.parseStatement(false)}})
		}
	}

	p.consume() // }

	return &Node{startPosition, NT_Match, &MatchNode{expression, cases, caseCount, defaultCase}}
}
