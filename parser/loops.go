package parser

func (p *Parser) parseLoop() *Node {
	loopPosition := p.consume().Position
	body := &Node{p.peek().Position, NT_Scope, p.parseScope(true)}

	return &Node{loopPosition, NT_Loop, body}
}
