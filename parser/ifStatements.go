package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseIfStatement(enteredScope bool) *Node {
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

			if elifConditionType.Type != data.DT_Unknown && elifConditionType.Type != data.DT_Bool {
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

	// Collect else body
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

	// Optimize if statements
	if p.optimize {
		optimizeIfs(&ifs)

		// Removed all if statements
		if len(ifs) == 0 {
			// No else statement, return next statement
			if elseBody == nil {
				return p.parseStatement(enteredScope)
				// Return else body
			} else {
				return elseBody
			}
		}

		// Only one always true if remains, return if body
		if len(ifs) == 1 && ifs[0].Condition.NodeType == NT_Literal && ifs[0].Condition.Value.(*LiteralNode).Value.(bool) {
			return ifs[0].Body
		}
	}

	return &Node{ifPosition, NT_If, &IfNode{ifs, elseBody}}
}

func optimizeIfs(ifStatements *[]*IfStatement) {
	forRemoval := make([]bool, len(*ifStatements))

	for i, statement := range *ifStatements {
		if statement.Condition.NodeType == NT_Literal && statement.Condition.Value.(*LiteralNode).PrimitiveType == data.DT_Bool {
			// Condition is always true
			if statement.Condition.Value.(*LiteralNode).Value.(bool) {
				// Discard all following if statements
				*ifStatements = (*ifStatements)[:i+1]
				return
			}

			// Condition is always false, mark it for removal
			forRemoval[i] = true
		}
	}

	// Remove marked statements
	for i := len(forRemoval) - 1; i >= 0; i-- {
		if forRemoval[i] {
			*ifStatements = append((*ifStatements)[:i], (*ifStatements)[i+1:]...)
		}
	}
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

	if conditionType.Type != data.DT_Unknown && conditionType.Type != data.DT_Bool {
		p.newError(condition.Position, "Condition expression data type has to be Bool.")
	}

	return condition
}
