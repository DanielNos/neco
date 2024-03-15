package parser

import (
	"fmt"
	data "neco/dataStructures"
	"neco/lexer"
	"neco/logger"
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
	condition = &Node{condition.Position, NT_Not, &TypedBinaryNode{nil, condition, data.DataType{data.DT_Bool, nil}}}
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
		condition = &Node{condition.Position, NT_Not, &TypedBinaryNode{nil, condition, data.DataType{data.DT_Bool, nil}}}
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

	return &Node{forPosition.SetEndPos(p.peekPrevious().Position), NT_ForLoop, &ForLoopNode{initStatement, body}}
}

func (p *Parser) parseForEach() *Node {
	startPosition := p.consume().Position

	// Consume (
	p.consume()
	iteratorPosition := p.peek().Position

	// Generate iterator index variable declaration
	indexIdentifier := fmt.Sprintf("@LOOP_ITERATOR_%d", p.tokenIndex)
	indexIdentifierVariable := &Node{iteratorPosition, NT_Variable, &VariableNode{indexIdentifier, data.DataType{data.DT_Int, nil}}}
	p.appendScope(&Node{iteratorPosition, NT_VariableDeclare, &VariableDeclareNode{data.DataType{data.DT_Int, nil}, false, []string{indexIdentifier}}})

	// Generate assignment of zero to iterator index variable
	zeroLiteral := &Node{iteratorPosition, NT_Literal, &LiteralNode{data.DT_Int, int64(0)}}
	p.IntConstants[0] = -1 // Store zero in constants
	p.appendScope(&Node{iteratorPosition, NT_Assign, &BinaryNode{indexIdentifierVariable, zeroLiteral}})

	// Generate variable declaration for list size
	sizeIdentifier := fmt.Sprintf("@LIST_SIZE_%d", p.tokenIndex)
	sizeDeclaration := &Node{iteratorPosition, NT_VariableDeclare, &VariableDeclareNode{data.DataType{data.DT_Int, nil}, false, []string{sizeIdentifier}}}
	p.appendScope(sizeDeclaration)

	// Set list size variable to list size
	functionCallNode := &FunctionCallNode{-1, "length", nil, nil, &data.DataType{data.DT_Int, nil}}
	sizeFunctionCall := &Node{iteratorPosition, NT_FunctionCall, functionCallNode}
	p.appendScope(&Node{iteratorPosition, NT_Assign, &BinaryNode{indexIdentifierVariable, sizeFunctionCall}})

	// Enter loop scope
	p.enterScope()

	// Generate: if (index == size) { break }
	indexVariable := &Node{iteratorPosition, NT_Variable, &VariableNode{indexIdentifier, data.DataType{data.DT_Int, nil}}}
	sizeVaraible := &Node{iteratorPosition, NT_Variable, &VariableNode{sizeIdentifier, data.DataType{data.DT_Int, nil}}}
	condition := &Node{iteratorPosition, NT_Equal, &TypedBinaryNode{indexVariable, sizeVaraible, data.DataType{data.DT_Bool, nil}}}
	breakNode := &Node{iteratorPosition, NT_Break, nil}
	ifBody := &Node{iteratorPosition, NT_Scope, &ScopeNode{-1, []*Node{breakNode}}}
	ifStatement := &IfStatement{condition, ifBody}
	ifNode := &Node{iteratorPosition, NT_If, &IfNode{[]*IfStatement{ifStatement}, nil}}
	p.appendScope(ifNode)

	// Collect iterator variable
	typePosition := p.peek().Position
	iteratorType := p.parseType()
	typePosition = typePosition.SetEndPos(p.peekPrevious().Position)

	iteratorIdentifier := p.consume().Value
	iteratorVariable := &Node{p.peekPrevious().Position, NT_Variable, &VariableNode{iteratorIdentifier, iteratorType}}

	// Declare it and insert to scope
	iteratorDeclaration := &Node{iteratorPosition, NT_VariableDeclare, &VariableDeclareNode{iteratorType, false, []string{iteratorIdentifier}}}
	p.appendScope(iteratorDeclaration)

	// Insert it into symbol table
	p.insertSymbol(iteratorIdentifier, &Symbol{ST_Variable, &VariableSymbol{iteratorType, true, false}})

	// Consume in
	p.consume()

	// Collect enumerated expression
	expression := p.parseExpressionRoot()
	elementType := p.GetExpressionType(expression)

	// Set element type to list subtype (if type was derived)
	if elementType.DType != data.DT_NoType {
		elementType = elementType.SubType.(data.DataType)
	}

	// Check if list element can be assigned to iterator
	if !iteratorType.Equals(elementType) {
		logger.Error2CodePos(typePosition, expression.Position, fmt.Sprintf("Can't assign expression of type %s to variable of type %s.", elementType, iteratorType))
		p.newErrorNoMessage()
	}

	// Assign to iterated_expression[interator_index] to iterator
	iteratorIndexVariable := &Node{iteratorPosition, NT_Variable, &VariableNode{indexIdentifier, data.DataType{data.DT_Int, nil}}}
	indexExpression := &Node{iteratorPosition, NT_ListValue, &TypedBinaryNode{expression, iteratorIndexVariable, elementType}}
	p.appendScope(&Node{iteratorPosition, NT_Assign, &BinaryNode{iteratorVariable, indexExpression}})

	// Add enumerated expression to previous length() function call
	functionCallNode.Arguments = []*Node{expression}
	functionCallNode.ArgumentTypes = []data.DataType{elementType}

	// Consume )
	p.consume()

	// Collect body
	body := p.parseScope(false, true).(*Node)

	// Generate iterator_index + 1
	oneLiteral := &Node{iteratorPosition, NT_Literal, &LiteralNode{data.DT_Int, int64(1)}}
	p.IntConstants[1] = -1 // Store one in constants

	addOne := &Node{iteratorPosition, NT_Add, &TypedBinaryNode{iteratorIndexVariable, oneLiteral, data.DataType{data.DT_Int, nil}}}

	// Insert iterator_index = iterator_index + 1
	p.appendScope(&Node{iteratorPosition, NT_Assign, &BinaryNode{iteratorVariable, addOne}})

	// Leave scope
	p.leaveScope()

	// Consume
	return &Node{startPosition.SetEndPos(p.peekPrevious().Position), NT_Loop, body}
}
