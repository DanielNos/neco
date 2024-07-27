package parser

import (
	data "github.com/DanielNos/neco/dataStructures"
	"github.com/DanielNos/neco/lexer"
	"github.com/DanielNos/neco/logger"
)

func (p *Parser) parseMatch(isExpression bool) *Node {
	startPosition := p.consume().Position

	expression := p.parseExpressionRoot()
	cases := []*Node{}

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	p.consume() // {
	var defaultCase *Node = nil
	caseCount := 0

	// Collect cases
	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Skip empty lines
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
			continue
		}

		if p.peek().TokenType == lexer.TT_KW_default {
			casePosition := p.consume().Position // default
			p.consume()                          // =>

			if isExpression {
				defaultCase = p.parseExpressionRoot()
			} else {
				defaultCase = p.parseStatement(false)
			}

			defaultCase = &Node{casePosition, NT_Case, &CaseNode{[]*Node{}, defaultCase}}
			caseCount++

			// Parse case
		} else {
			// Default case has to be the last case
			if defaultCase != nil {
				p.newError(defaultCase.Position, "Default case has to be the last case.")
			}

			caseCount++
			casePosition := p.peek().Position

			// Collect case expressions
			expressions := []*Node{}
			expressions = append(expressions, p.parseExpressionRoot())

			for p.peek().TokenType == lexer.TT_DL_Comma {
				p.consume()
				expressions = append(expressions, p.parseExpressionRoot())
				caseCount++
			}

			p.consume() // =>

			var caseContent *Node

			if isExpression {
				caseContent = p.parseExpressionRoot()
			} else {
				caseContent = p.parseStatement(false)
			}

			cases = append(cases, &Node{casePosition, NT_Case, &CaseNode{expressions, caseContent}})
		}
	}

	p.consume() // }

	return &Node{startPosition, NT_Match, &MatchNode{expression, cases, caseCount, defaultCase, nil}}
}

func (p *Parser) parseMatchExpression() *Node {
	// Parse match statement
	matchNode := p.parseMatch(true)
	match := matchNode.Value.(*MatchNode)

	// Add default case to cases
	if match.Default != nil {
		match.Cases = append(match.Cases, match.Default)
	}

	// Collect expression data types and their counts
	expressionTypes := map[string]*dataTypeCount{}
	elementType := &data.DataType{data.DT_Unknown, nil}

	for _, caseExpression := range match.Cases {
		elementType = GetExpressionType(caseExpression.Value.(*CaseNode).Statement)
		signature := elementType.Signature()

		typeAndCount, exists := expressionTypes[signature]
		if !exists {
			expressionTypes[signature] = &dataTypeCount{elementType, 1}
		} else {
			typeAndCount.Count++
		}
	}

	// Check if all list elements have the same type and set element type to the most common type
	if len(expressionTypes) > 1 {
		elementType = p.checkElementTypes(expressionTypes, match.Cases, "match")
	}

	// Remove default case from cases
	if match.Default != nil {
		match.Cases = match.Cases[:len(match.Cases)-1]
	}

	match.DataType = elementType

	// Check if all values have cases
	p.checkCaseValueCoverage(matchNode)

	return matchNode
}

func (p *Parser) checkCaseValueCoverage(matchNode *Node) {
	match := matchNode.Value.(*MatchNode)
	matchedExpressionType := GetExpressionType(match.Expression)

	if match.Default != nil && matchedExpressionType.Type != data.DT_Bool {
		return
	}

	// Boolean
	if matchedExpressionType.Type == data.DT_Bool {
		isCovered := checkCoverage(match, data.DT_Bool, []any{false, true})

		// All values aren't covered
		if !isCovered {
			p.newError(matchNode.Position, "Not all possible matched values are covered. Add cases for all possible values or a default case.")
			// All values are covered, but default case exists
		} else if isCovered && match.Default != nil {
			logger.WarningCodePos(match.Default.Position, "Unnecessary default case. All possible expression types are covered.")

			// Remove redundant default case
			if p.optimize {
				match.Default = nil
			}
		}

		return
	}

	// Default isn't necessary for options if "none" case is covered
	if matchedExpressionType.Type == data.DT_Option {
		foundNone := checkCoverage(match, data.DT_None, []any{nil})

		if !foundNone {
			p.newError(matchNode.Position, "Not all possible matched values are covered. Add cases for all possible values or a default case.")
		}
		return
	}

	p.newError(matchNode.Position, "Not all possible matched values are covered. Add cases for all possible values or a default case.")
}

func checkCoverage(matchNode *MatchNode, dataType data.PrimitiveType, values []any) bool {
	foundValues := make([]bool, len(values))

	// Check all case expressions for values
	for _, matchCase := range matchNode.Cases {
		for _, expression := range matchCase.Value.(*CaseNode).Expressions {

			if expression.NodeType == NT_Literal && expression.Value.(*LiteralNode).PrimitiveType == dataType {
				// Check if expression is any of the searched values
				for i, value := range values {
					if value == expression.Value.(*LiteralNode).Value {
						foundValues[i] = true
					}
				}
			}
		}
	}

	// Check if all values were found
	for _, found := range foundValues {
		if !found {
			return false
		}
	}

	return true
}
