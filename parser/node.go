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
	NT_And
	NT_Or
	NT_Not
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
	NT_If
	NT_Loop
)

var NodeTypeToString = map[NodeType]string {
	NT_Module: "Module",
	NT_Scope: "Scope",
	NT_StatementList: "StatementList",
	NT_VariableDeclare: "VariableDeclare",
	NT_Assign: "Assign",
	NT_Literal: "Literal",
	NT_And: "&",
	NT_Or: "|",
	NT_Not: "!",
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
	NT_If: "If",
	NT_Loop: "Loop",
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
	constant bool
	identifiers []string
}

type AssignNode struct {
	identifier string
	expression *Node
}

type LiteralValue interface{}

type LiteralNode struct {
	dataType DataType
	value LiteralValue
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

type IfNode struct {
	condition *Node
	body *Node
	elseIfs []*Node
	elseBody *Node
}

var TokenTypeToNodeType = map[lexer.TokenType]NodeType {
	lexer.TT_OP_And: NT_And,
	lexer.TT_OP_Or: NT_Or,
	lexer.TT_OP_Not: NT_Not,

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

var OperationAssignTokenToNodeType = map[lexer.TokenType]NodeType {
	lexer.TT_KW_AddAssign: NT_Add,
	lexer.TT_KW_SubtractAssign: NT_Subtract,
	lexer.TT_KW_MultiplyAssign: NT_Multiply,
	lexer.TT_KW_DivideAssign: NT_Divide,
	lexer.TT_KW_PowerAssign: NT_Power,
	lexer.TT_KW_ModuloAssign: NT_Modulo,
}

func (nt NodeType) IsOperator() bool {
	return nt >= NT_And && nt <= NT_GreaterEqual
}

func (nt NodeType) IsComparisonOperator() bool {
	return nt >= NT_Equal && nt <= NT_GreaterEqual
}

func (nt NodeType) IsLogicOperator() bool {
	return nt == NT_And || nt == NT_Or
}
