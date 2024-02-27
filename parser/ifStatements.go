package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseIfStatement() *Node {
	ifPosition := p.consume().Position

	// Collect condition
	condition := p.parseCondition(true)

	// Collect body
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := p.parseScope(true, true).(*Node)

	// Collec else ifs
	ifs := []*IfStatement{{condition, body}}

	for {
		if p.peek().TokenType == lexer.TT_EndOfCommand && p.peekNext().TokenType == lexer.TT_KW_elif {
			p.consume()
		}

		// Collect else if
		if p.peek().TokenType == lexer.TT_KW_elif {
			p.consume()

			// Collect condition
			p.consume()
			elifCondition := p.parseExpressionRoot()
			p.consume()

			// Check condition type
			elifConditionType := p.GetExpressionType(elifCondition)

			if elifConditionType.DType != data.DT_Bool {
				p.newError(elifCondition.Position, "Condition expression data type has to be Bool.")
			}

			// Collect body
			if p.peek().TokenType == lexer.TT_EndOfCommand {
				p.consume()
			}

			elifBody := p.parseScope(true, true).(*Node)

			ifs = append(ifs, &IfStatement{elifCondition, elifBody})

		} else {
			break
		}
	}

	// Collect else
	if p.peek().TokenType == lexer.TT_EndOfCommand && p.peekNext().TokenType == lexer.TT_KW_else {
		p.consume()
	}

	var elseBody *Node = nil
	if p.peek().TokenType == lexer.TT_KW_else {
		elsePosition := p.consume().Position
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}
		elseBody = p.parseScope(true, true).(*Node)
		elseBody.Position = elsePosition
	}

	return &Node{ifPosition, NT_If, &IfNode{ifs, elseBody}}
}

func (p *Parser) parseCondition(removeParenthesis bool) *Node {
	// Collect condition
	var condition *Node
	if removeParenthesis {
		p.consume()
		condition = p.parseExpressionRoot()
		p.consume()
	} else {
		condition = p.parseExpressionRoot()
	}

	// Check condition type
	conditionType := p.GetExpressionType(condition)

	if conditionType.DType != data.DT_Bool {
		p.newError(condition.Position, "Condition expression data type has to be Bool.")
	}

	return condition
}
