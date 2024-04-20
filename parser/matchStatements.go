package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseMatch(isExpression bool) *Node {
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

			if isExpression {
				defaultCase = p.parseExpressionRoot()
			} else {
				defaultCase = p.parseStatement(false)
			}
			caseCount++

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

			var caseContent *Node

			if isExpression {
				caseContent = p.parseExpressionRoot()
			} else {
				caseContent = p.parseStatement(false)
			}

			cases = append(cases, &Node{casePosition, NT_Case, &CaseNode{expressions, caseContent}})
		}
	}

	p.consume() // }

	return &Node{startPosition, NT_Match, &MatchNode{expression, cases, caseCount, defaultCase, nil}}
}

func (p *Parser) parseMatchExpression() *Node {
	matchNode := p.parseMatch(true)
	match := matchNode.Value.(*MatchNode)

	if match.Default != nil {
		match.Cases = append(match.Cases, &Node{match.Default.Position, NT_Case, &CaseNode{[]*Node{}, match.Default}})
	}

	expressionTypes := map[string]*dataTypeCount{}
	elementType := &data.DataType{data.DT_Unknown, nil}

	for _, caseExpression := range match.Cases {
		elementType = GetExpressionType(caseExpression.Value.(*CaseNode).Statement)
		signature := elementType.Signature()

		typeAndCount, exists := expressionTypes[signature]
		if !exists {
			expressionTypes[signature] = &dataTypeCount{elementType, 1}
		} else {
			typeAndCount.Count++
		}
	}

	// Check if all list elements have the same type and set element type to the most common type
	if len(expressionTypes) > 1 {
		elementType = p.checkElementTypes(expressionTypes, match.Cases, "match")
	}

	match.DataType = elementType
	if match.Default != nil {
		match.Cases = match.Cases[:len(match.Cases)-1]
	}

	return matchNode
}
