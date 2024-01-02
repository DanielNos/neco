package parser

import "neko/lexer"

func (p *Parser) parseIfStatement() *Node {
	ifPosition := p.consume().Position

	// Collect condition
	p.consume()
	condition := p.parseExpression(MINIMAL_PRECEDENCE)
	p.consume()
	
	// Check condition type
	conditionType := p.getExpressionType(condition)
	if !conditionType.Equals(VariableType{DT_Bool, false}) {
		conditionPosition := getExpressionPosition(condition, condition.position.StartChar, condition.position.EndChar)
		p.newError(&conditionPosition, "Condition expression isn't of type Bool.")
	}
	
	// Collect body
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(true)}

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
			elifCondition := p.parseExpression(MINIMAL_PRECEDENCE)
			p.consume()

			// Check condition type
			elifConditionType := p.getExpressionType(elifCondition)

			if !elifConditionType.Equals(VariableType{DT_Bool, false}) {
				conditionPosition := getExpressionPosition(elifCondition, elifCondition.position.StartChar, elifCondition.position.EndChar)
				p.newError(&conditionPosition, "Condition expression isn't of type Bool.")
			}
			
			// Collect body
			if p.peek().TokenType == lexer.TT_EndOfCommand {
				p.consume()
			}

			elifBody := &Node{p.peek().Position, NT_Scope, p.parseScope(true)}

			elseIfs = append(elseIfs, &Node{elifPosition, NT_If, &IfNode{elifCondition, elifBody, nil, nil}})

		} else {
			break
		}
	}

	// Collect else
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	var elseBody *Node = nil
	if p.peek().TokenType == lexer.TT_KW_else {
		elsePosition := p.consume().Position
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}
		elseBody = &Node{elsePosition, NT_Scope, p.parseScope(true)}
	}

	return &Node{ifPosition, NT_If, &IfNode{condition, body, elseIfs, elseBody}}
}
