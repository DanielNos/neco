package parser

import "neko/lexer"

func (p *Parser) parseLoop() *Node {
	loopPosition := p.consume().Position

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(true)}

	return &Node{loopPosition, NT_Loop, body}
}

func (p *Parser) parseWhile() *Node {
	p.consume()

	// Collect condition
	condition := p.parseCondition()

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}
	
	// Enter loop scope
	p.enterScope()

	// Construct condition using if node
	breakNode := &Node{condition.position, NT_Drop, 1}

	ifBlock := &Node{condition.position, NT_Scope, &ScopeNode{p.scopeCounter, []*Node{breakNode}}}
	p.scopeCounter++

	// Create and insert if node into loop body
	ifStatement := &Node{condition.position, NT_If, &IfNode{condition, ifBlock, nil, nil}}
	p.appendScope(ifStatement)

	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false)}

	return body
}
