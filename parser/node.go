package parser

import (
	"neko/dataStructures"
	"neko/lexer"
)

type NodeValue interface{}

type Node struct {
	position *dataStructures.CodePos
	nodeType NodeType
	value NodeValue
}

type NodeType uint8

const (
	NT_Module NodeType = iota
	NT_Scope
	NT_StatementList
	NT_VariableDeclare
	NT_Assign
	NT_Literal
	NT_Add
	NT_Subtract
	NT_Multiply
	NT_Divide
	NT_Power
	NT_Modulo
	NT_Equal
	NT_NotEqual
	NT_Lower
	NT_Greater
	NT_LowerEqual
	NT_GreaterEqual
	NT_Variable
	NT_FunctionDeclare
	NT_FunctionCall
	NT_Return
)

var NodeTypeToString = map[NodeType]string {
	NT_Module: "Module",
	NT_Scope: "Scope",
	NT_StatementList: "StatementList",
	NT_VariableDeclare: "VariableDeclare",
	NT_Assign: "Assign",
	NT_Literal: "Literal",
	NT_Add: "+",
	NT_Subtract: "-",
	NT_Multiply: "*",
	NT_Divide: "/",
	NT_Power: "^",
	NT_Modulo: "%",
	NT_Equal: "==",
	NT_NotEqual: "!=",
	NT_Lower: "<",
	NT_Greater: ">",
	NT_LowerEqual: "<=",
	NT_GreaterEqual: ">=",
	NT_Variable: "Variable",
	NT_FunctionDeclare: "FunctionDeclare",
	NT_FunctionCall: "FunctionCall",
	NT_Return: "Return",
}

func (nt NodeType) String() string {
	return NodeTypeToString[nt]
}

type ModuleNode struct {
	filePath string
	name string
	statements *ScopeNode
}

type ScopeNode struct {
	id int
	statements []*Node
}

type VariableDeclareNode struct {
	variableType VariableType
	identifiers []string
}

type AssignNode struct {
	identifiers []string
	expression *Node
}

type LiteralNode struct {
	dataType DataType
	value string
}

type BinaryNode struct {
	left *Node
	right *Node
}

type VariableNode struct {
	identifier string
	variableType VariableType
}

type FunctionDeclareNode struct {
	identifier string
	parameters []Parameter
	returnType VariableType
	body *Node
}

type Parameter struct {
	dataType DataType
	identifier string
	defaultValue *Node
}

type FunctionCallNode struct {
	identifier string
	arguments []*Node
	returnType *VariableType
}

var TokenTypeToNodeType = map[lexer.TokenType]NodeType {
	lexer.TT_OP_Add: NT_Add,
	lexer.TT_OP_Subtract: NT_Subtract,
	lexer.TT_OP_Multiply: NT_Multiply,
	lexer.TT_OP_Divide: NT_Divide,
	lexer.TT_OP_Power: NT_Power,
	lexer.TT_OP_Modulo: NT_Modulo,
	lexer.TT_OP_Equal: NT_Equal,
	lexer.TT_OP_NotEqual: NT_NotEqual,
	lexer.TT_OP_Lower: NT_Lower,
	lexer.TT_OP_Greater: NT_Greater,
	lexer.TT_OP_LowerEqual: NT_LowerEqual,
	lexer.TT_OP_GreaterEqual: NT_GreaterEqual,
}

func (nt NodeType) IsOperator() bool {
	return nt >= NT_Add && nt <= NT_GreaterEqual
}
