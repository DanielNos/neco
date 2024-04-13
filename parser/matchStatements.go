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

		// Parse case
		if p.peek().TokenType == lexer.TT_KW_case {
			// Default case has to be the last case
			if defaultCase != nil {
				p.newError(defaultCase.Position, "Default case has to be the last case.")
			}

			caseCount++
			casePosition := p.consume().Position

			// Collect case expressions
			expressions := []*Node{}
			expressions = append(expressions, p.parseExpressionRoot())

			for p.peek().TokenType == lexer.TT_DL_Comma {
				p.consume()
				expressions = append(expressions, p.parseExpressionRoot())
				caseCount++
			}

			colonPosition := p.consume().Position // :

			p.enterScope()

			// Parse statements
			for p.peek().TokenType != lexer.TT_KW_case && p.peek().TokenType != lexer.TT_KW_default && p.peek().TokenType != lexer.TT_DL_BraceClose {
				if p.peek().TokenType == lexer.TT_EndOfCommand {
					p.consume()
					continue
				}

				statement := p.parseStatement(true)

				if statement != nil {
					p.appendScope(statement)
				}
			}

			scope := p.leaveScope()
			scopeNode := &Node{colonPosition, NT_Scope, scope}

			cases = append(cases, &Node{casePosition, NT_Case, &CaseNode{expressions, scopeNode}})
			// Parse default
		} else if p.peek().TokenType == lexer.TT_KW_default {
			defaultPosition := p.consume().Position
			p.consume() // :

			p.enterScope()

			// Parse statements
			for p.peek().TokenType != lexer.TT_KW_case && p.peek().TokenType != lexer.TT_KW_default && p.peek().TokenType != lexer.TT_DL_BraceClose {
				if p.peek().TokenType == lexer.TT_EndOfCommand {
					p.consume()
					continue
				}

				// Collect statement
				statement := p.parseStatement(true)

				if statement != nil {
					p.appendScope(statement)
				}
			}

			scope := p.leaveScope()
			defaultCase = &Node{defaultPosition, NT_Scope, scope}

			p.scopeCounter++
		}
	}

	p.consume() // }

	return &Node{startPosition, NT_Match, &MatchNode{expression, cases, caseCount, defaultCase}}
}
