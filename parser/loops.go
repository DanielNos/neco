package parser

import "neko/lexer"

func (p *Parser) parseLoop() *Node {
	loopPosition := p.consume().Position

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(true, false)}

	return &Node{loopPosition, NT_Loop, body}
}

func (p *Parser) parseWhile() *Node {
	p.consume()

	// Collect condition
	condition := p.parseCondition(true)

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Enter loop scope
	p.enterScope()

	// Construct condition using if node
	breakNode := &Node{condition.Position, NT_Drop, 1}

	ifBlock := &Node{condition.Position, NT_Scope, &ScopeNode{p.scopeCounter, []*Node{breakNode}}}
	p.scopeCounter++

	// Create and insert if node into loop body
	ifStatement := &Node{condition.Position, NT_If, &IfNode{condition, ifBlock, nil, nil}}
	p.appendScope(ifStatement)

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false, false)}

	return body
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

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false, false)}

	p.leaveScope()

	return &Node{forPosition, NT_For, &ForNode{initStatement, condition, stepStatement, body}}
}
