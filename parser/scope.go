package parser

type symbolTable map[string]*Symbol

func (p *Parser) findSymbol(identifier string) *Symbol {
	stackNode := p.symbolTableStack.Top

	for stackNode != nil {
		symbol, exists := stackNode.Value.(symbolTable)[identifier]
		
		if exists {
			return symbol
		}

		stackNode = stackNode.Previous
	}

	return nil
}

func (p *Parser) getSymbol(identifier string) *Symbol {
	symbol, exists := p.symbolTableStack.Top.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}

	return nil
}

func (p *Parser) getGlobalSymbol(identifier string) *Symbol {
	symbol, exists := p.symbolTableStack.Bottom.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}

	return nil
}
