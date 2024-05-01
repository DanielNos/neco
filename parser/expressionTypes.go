package parser

import (
	"fmt"
	data "neco/dataStructures"
	"neco/logger"
)

func (p *Parser) deriveType(expression *Node) *data.DataType {
	// Unwrap option
	if expression.NodeType == NT_Unwrap {
		unwrappedNode := expression.Value.(*Node)
		unwrappedNodeType := GetExpressionType(unwrappedNode)

		// Expression can't be unwrapped
		if unwrappedNodeType.Type != data.DT_Option {
			p.newError(GetExpressionPosition(unwrappedNode), "Can't unwrap an expression with type "+unwrappedNodeType.String()+".")
		}

		return unwrappedNodeType.SubType.(*data.DataType)
	}

	// Match statement
	if expression.NodeType == NT_Match {
		return expression.Value.(*MatchNode).DataType
	}

	// Operators
	if expression.NodeType.IsOperator() {
		return p.deriveOperatorType(expression)
	}

	return GetExpressionType(expression)
}

func (p *Parser) deriveOperatorType(expression *Node) *data.DataType {
	binaryNode := expression.Value.(*TypedBinaryNode)

	// Node has it's type stored already
	if binaryNode.DataType.Type != data.DT_Unknown {
		return binaryNode.DataType
	}

	// Unary operator
	if binaryNode.Left == nil {
		return GetExpressionType(binaryNode.Right)
	}

	// Ternary expression
	if expression.NodeType == NT_Ternary {
		leftType := p.deriveType(binaryNode.Left)

		// Left side isn't bool
		if leftType.Type != data.DT_Bool {
			p.newError(GetExpressionPosition(binaryNode.Left), "Left side of ternary operator ?? has to be of type bool.")
		}

		ternaryBranches := binaryNode.Right.Value.(*TypedBinaryNode)
		leftBranchType := p.deriveType(ternaryBranches.Left)
		rightBranchType := p.deriveType(ternaryBranches.Right)

		// Branches have different types
		if !leftBranchType.Equals(rightBranchType) {
			p.newError(GetExpressionPosition(binaryNode.Right), "Both branches of ternary operator ?? have to be of the same type. Left is "+leftBranchType.String()+", right is "+rightBranchType.String()+".")
		}

		ternaryBranches.DataType = leftBranchType
		binaryNode.DataType = leftBranchType
		return leftBranchType
	}

	// Type of tranches of ternary operator can't be derived
	if expression.NodeType == NT_TernaryBranches {
		p.newError(GetExpressionPosition(expression), "Unexpected expression. Are you missing an \"??\" operator?")
		return &data.DataType{data.DT_Unknown, nil}
	}

	// Collect left and right node data types
	leftType := p.deriveType(binaryNode.Left)
	rightType := p.deriveType(binaryNode.Right)

	// Error in one of types
	if leftType.Type == data.DT_Unknown || rightType.Type == data.DT_Unknown {
		return &data.DataType{data.DT_Unknown, nil}
	}

	// In operator has to be used on set with correct type
	if expression.NodeType == NT_In {
		// Right type isn't a set
		if rightType.Type != data.DT_Set {
			p.newError(GetExpressionPosition(binaryNode.Right), "Right side of operator \"in\" has to be a set.")
			// Left type isn't set's sub-type
		} else if !rightType.SubType.(*data.DataType).CanBeAssigned(leftType) {
			p.newErrorNoMessage()
			logger.Error2CodePos(GetExpressionPosition(binaryNode.Left), GetExpressionPosition(binaryNode.Right), "Left expression type ("+leftType.String()+") doesn't match the set element type ("+rightType.SubType.(*data.DataType).String()+").")
		}
		binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
		return binaryNode.DataType
	}

	if expression.NodeType == NT_UnpackOrDefault {
		// Right side of ?! can't be none
		if rightType.Type == data.DT_None {
			p.newError(GetExpressionPosition(binaryNode.Right), "Expression on the right side of ?! operator can't be none.")
		} else if rightType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(binaryNode.Right), "Expression on the right side of ?! operator can't be possibly none.")
		}

		// Check if left and right type is compatible
		if leftType.Type == data.DT_Option {
			if !leftType.CanBeAssigned(rightType) {
				p.newError(GetExpressionPosition(binaryNode.Left), "Both sides of operator ?! have to have the same type. Left is "+leftType.String()+", right is "+rightType.String()+".")
			}
			// Left has to be option or none
		} else if leftType.Type != data.DT_None {
			p.newError(GetExpressionPosition(binaryNode.Left), "Left side of operator ?! has to be an option type.")
		}

		binaryNode.DataType = rightType
		return rightType
	}

	// Compatible data types, check if operator is allowed
	if leftType.CanBeAssigned(rightType) {
		// Can't do any operations on options without unwrapping
		if leftType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left), "Options need to be unwrapped or matched to access their values.")
		}
		if rightType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right), "Options need to be unwrapped or matched to access their values.")
		}

		// Logic operators can be used only on booleans
		if expression.NodeType.IsLogicOperator() && (leftType.Type != data.DT_Bool || rightType.Type != data.DT_Bool) {
			p.newError(expression.Position, "Operator "+expression.NodeType.String()+" can be only used on expressions of type bool.")
			binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
			return binaryNode.DataType
		}

		// Comparison operators return boolean
		if expression.NodeType.IsComparisonOperator() {
			binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
			return binaryNode.DataType
		}

		// Can't do non-comparison operations on enums
		if leftType.Type == data.DT_Enum || rightType.Type == data.DT_Enum {
			p.newError(expression.Position, "Operator "+expression.NodeType.String()+" can't be used on enum constants.")
			return &data.DataType{data.DT_Unknown, nil}
		}

		// Only + can be used on strings and lists
		if (leftType.Type == data.DT_String || leftType.Type == data.DT_List) && expression.NodeType != NT_Add {
			p.newError(expression.Position, "Can't use operator "+expression.NodeType.String()+" on data types "+leftType.String()+" and "+rightType.String()+".")
			return &data.DataType{data.DT_Unknown, nil}
		}

		// Return left type
		if leftType.Type != data.DT_Unknown {
			binaryNode.DataType = leftType
			return leftType
		}

		// Return right type
		if rightType.Type != data.DT_Unknown {
			binaryNode.DataType = rightType
			return rightType
		}

		// Neither have type
		return leftType
	}

	// Left or right doesn't have a type
	if leftType.Type == data.DT_Unknown || rightType.Type == data.DT_Unknown {
		return leftType
	}

	// Failed to determine data type
	p.newError(expression.Position, "Operator "+expression.NodeType.String()+" is used on incompatible data types "+leftType.String()+" and "+rightType.String()+".")
	return &data.DataType{data.DT_Unknown, nil}
}

func GetExpressionType(expression *Node) *data.DataType {
	// Binary nodes store their type
	if expression.NodeType.IsOperator() {
		return expression.Value.(*TypedBinaryNode).DataType
	}

	switch expression.NodeType {
	case NT_Literal:
		return &data.DataType{expression.Value.(*LiteralNode).PrimitiveType, nil}
	case NT_Variable:
		return expression.Value.(*VariableNode).DataType
	case NT_FunctionCall:
		return expression.Value.(*FunctionCallNode).ReturnType
	case NT_List:
		return expression.Value.(*ListNode).DataType
	case NT_ListValue:
		return GetExpressionType(expression.Value.(*TypedBinaryNode).Left).SubType.(*data.DataType)
	case NT_Enum:
		return &data.DataType{data.DT_Enum, expression.Value.(*EnumNode).Identifier}
	case NT_Object:
		return &data.DataType{data.DT_Object, expression.Value.(*ObjectNode).Identifier}
	case NT_ObjectField:
		return expression.Value.(*ObjectFieldNode).DataType
	case NT_Set:
		return expression.Value.(*ListNode).DataType
	case NT_Unwrap:
		return GetExpressionType(expression.Value.(*Node)).SubType.(*data.DataType)
	case NT_IsNone:
		return &data.DataType{data.DT_Bool, nil}
	case NT_Match:
		return expression.Value.(*MatchNode).DataType
	case NT_Ternary:
		return GetExpressionType(expression.Value.(*TypedBinaryNode).Right)
	case NT_TernaryBranches:
		return expression.Value.(*TypedBinaryNode).DataType
	}

	panic("Can't determine expression data type from " + NodeTypeToString[expression.NodeType] + fmt.Sprintf(" (%d)", expression.NodeType) + ".")
}
