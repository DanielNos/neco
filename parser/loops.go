package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseLoop() *Node {
	loopPosition := p.consume().Position

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := p.parseScope(true, true).(*Node)

	return &Node{loopPosition, NT_Loop, body}
}

func (p *Parser) parseWhile() *Node {
	startPosition := p.consume().Position

	// Collect condition
	condition := p.parseCondition(true)
	condition = &Node{condition.Position, NT_Not, &BinaryNode{nil, condition, data.DataType{data.DT_Bool, nil}}}

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Enter loop scope
	p.enterScope()

	// Construct condition using if node
	breakNode := &Node{condition.Position, NT_Break, 1}

	ifBlock := &Node{condition.Position, NT_Scope, &ScopeNode{p.scopeCounter, []*Node{breakNode}}}
	p.scopeCounter++

	// Create and insert negated if node into loop body
	p.appendScope(&Node{condition.Position, NT_If, &IfNode{[]*IfStatement{{condition, ifBlock}}, nil}})

	body := p.parseScope(false, true).(*Node)

	p.leaveScope()

	return &Node{startPosition, NT_Loop, body}
}

func (p *Parser) parseFor() *Node {
	forPosition := p.consume().Position

	// Collect init, condition and step
	p.enterScope()
	p.consume()

	initStatement := p.parseStatement(false)
	p.consume()

	condition := p.parseCondition(false)

	stepStatement := p.parseStatement(false)

	p.consume()

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := p.parseScope(false, true).(*Node)

	p.leaveScope()

	return &Node{forPosition, NT_For, &ForNode{initStatement, condition, stepStatement, body}}
}
