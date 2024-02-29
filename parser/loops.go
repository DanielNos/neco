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

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Enter loop scope
	p.enterScope()

	// Construct condition using if node
	condition = &Node{condition.Position, NT_Not, &BinaryNode{nil, condition, data.DataType{data.DT_Bool, nil}}}
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

	// Collect init expression if it exists
	p.consume()
	p.enterScope()

	if p.peek().TokenType != lexer.TT_EndOfCommand {
		p.appendScope(p.parseStatement(false))
	}

	// Statements were added to scope, move them to variable
	initStatement := p.scopeNodeStack.Top.Value.(*ScopeNode).Statements
	p.scopeNodeStack.Top.Value.(*ScopeNode).Statements = []*Node{}

	// Consume EOC
	p.consume()

	// Collect condition expression
	var condition *Node = nil
	if p.peek().TokenType != lexer.TT_EndOfCommand {
		condition = p.parseCondition(false)
	}
	p.consume()

	// Collect step statement
	var stepStatement *Node = nil
	if p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		stepStatement = p.parseStatement(false)
	}

	p.consume()

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Construct condition using if node
	if condition != nil {
		condition = &Node{condition.Position, NT_Not, &BinaryNode{nil, condition, data.DataType{data.DT_Bool, nil}}}
		breakNode := &Node{condition.Position, NT_Break, 1}

		ifBlock := &Node{condition.Position, NT_Scope, &ScopeNode{p.scopeCounter, []*Node{breakNode}}}
		p.scopeCounter++

		// Create and insert negated if node into loop body
		p.appendScope(&Node{condition.Position, NT_If, &IfNode{[]*IfStatement{{condition, ifBlock}}, nil}})
	}

	// Parse body
	body := p.parseScope(false, true).(*Node)

	// Append step to body
	if stepStatement != nil {
		p.appendScope(stepStatement)
	}

	p.leaveScope()

	return &Node{forPosition, NT_ForLoop, &ForLoopNode{initStatement, body}}
}
