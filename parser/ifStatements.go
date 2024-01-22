package parser

import (
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
	elseIfs := []*Node{}

	for {
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}

		// Collect else if
		if p.peek().TokenType == lexer.TT_KW_elif {
			elifPosition := p.consume().Position

			// Collect condition
			p.consume()
			elifCondition := p.parseExpressionRoot()
			p.consume()

			// Check condition type
			elifConditionType := p.getExpressionType(elifCondition)

			if !elifConditionType.Equals(VariableType{DT_Bool, false}) {
				conditionPosition := getExpressionPosition(elifCondition, elifCondition.Position.StartChar, elifCondition.Position.EndChar)
				p.newError(&conditionPosition, "Condition expression isn't of type Bool.")
			}

			// Collect body
			if p.peek().TokenType == lexer.TT_EndOfCommand {
				p.consume()
			}

			elifBody := p.parseScope(true, true).(*Node)

			elseIfs = append(elseIfs, &Node{elifPosition, NT_If, &IfNode{elifCondition, elifBody, nil, nil}})

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

	return &Node{ifPosition, NT_If, &IfNode{condition, body, elseIfs, elseBody}}
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
	conditionType := p.getExpressionType(condition)

	if !conditionType.Equals(VariableType{DT_Bool, false}) {
		conditionPosition := getExpressionPosition(condition, condition.Position.StartChar, condition.Position.EndChar)
		p.newError(&conditionPosition, "Condition expression isn't of type Bool.")
	}

	return condition
}
