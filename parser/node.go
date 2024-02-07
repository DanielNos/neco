package parser

import (
	"neco/dataStructures"
	"neco/lexer"
)

type NodeValue interface{}

type Node struct {
	Position *dataStructures.CodePos
	NodeType NodeType
	Value    NodeValue
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
	NT_For
	NT_Drop
)

var NodeTypeToString = map[NodeType]string{
	NT_Module:          "Module",
	NT_Scope:           "Scope",
	NT_StatementList:   "StatementList",
	NT_VariableDeclare: "VariableDeclare",
	NT_Assign:          "Assign",
	NT_Literal:         "Literal",
	NT_And:             "&",
	NT_Or:              "|",
	NT_Not:             "!",
	NT_Add:             "+",
	NT_Subtract:        "-",
	NT_Multiply:        "*",
	NT_Divide:          "/",
	NT_Power:           "^",
	NT_Modulo:          "%",
	NT_Equal:           "==",
	NT_NotEqual:        "!=",
	NT_Lower:           "<",
	NT_Greater:         ">",
	NT_LowerEqual:      "<=",
	NT_GreaterEqual:    ">=",
	NT_Variable:        "Variable",
	NT_FunctionDeclare: "FunctionDeclare",
	NT_FunctionCall:    "FunctionCall",
	NT_Return:          "Return",
	NT_If:              "If",
	NT_Loop:            "Loop",
	NT_For:             "For",
	NT_Drop:            "Drop",
}

func (nt NodeType) String() string {
	return NodeTypeToString[nt]
}

type ModuleNode struct {
	FilePath   string
	Name       string
	Statements *ScopeNode
}

type ScopeNode struct {
	Id         int
	Statements []*Node
}

type VariableDeclareNode struct {
	VariableType VariableType
	Constant     bool
	Identifiers  []string
}

type AssignNode struct {
	Identifier string
	Expression *Node
}

type LiteralValue interface{}

type LiteralNode struct {
	DataType DataType
	Value    LiteralValue
}

type BinaryNode struct {
	Left     *Node
	Right    *Node
	DataType DataType
}

type VariableNode struct {
	Identifier   string
	VariableType VariableType
}

type FunctionDeclareNode struct {
	Number     int
	Identifier string
	Parameters []Parameter
	ReturnType VariableType
	Body       *Node
}

type Parameter struct {
	DataType     VariableType
	Identifier   string
	DefaultValue *Node
}

type FunctionCallNode struct {
	Number        int
	Identifier    string
	Arguments     []*Node
	ArgumentTypes []VariableType
	ReturnType    *VariableType
}

type IfNode struct {
	IfStatements []*IfStatement
	ElseBody     *Node
}

type IfStatement struct {
	Condition *Node
	Body      *Node
}

type ForNode struct {
	InitStatement *Node
	Condition     *Node
	StepStatement *Node
	Body          *Node
}

var TokenTypeToNodeType = map[lexer.TokenType]NodeType{
	lexer.TT_OP_And: NT_And,
	lexer.TT_OP_Or:  NT_Or,
	lexer.TT_OP_Not: NT_Not,

	lexer.TT_OP_Add:      NT_Add,
	lexer.TT_OP_Subtract: NT_Subtract,
	lexer.TT_OP_Multiply: NT_Multiply,
	lexer.TT_OP_Divide:   NT_Divide,
	lexer.TT_OP_Power:    NT_Power,
	lexer.TT_OP_Modulo:   NT_Modulo,

	lexer.TT_OP_Equal:        NT_Equal,
	lexer.TT_OP_NotEqual:     NT_NotEqual,
	lexer.TT_OP_Lower:        NT_Lower,
	lexer.TT_OP_Greater:      NT_Greater,
	lexer.TT_OP_LowerEqual:   NT_LowerEqual,
	lexer.TT_OP_GreaterEqual: NT_GreaterEqual,
}

var OperationAssignTokenToNodeType = map[lexer.TokenType]NodeType{
	lexer.TT_KW_AddAssign:      NT_Add,
	lexer.TT_KW_SubtractAssign: NT_Subtract,
	lexer.TT_KW_MultiplyAssign: NT_Multiply,
	lexer.TT_KW_DivideAssign:   NT_Divide,
	lexer.TT_KW_PowerAssign:    NT_Power,
	lexer.TT_KW_ModuloAssign:   NT_Modulo,
}

func (n *Node) IsBinaryNode() bool {
	return n.NodeType >= NT_Add && n.NodeType <= NT_GreaterEqual && n.NodeType != NT_Not && n.Value.(*BinaryNode).Left != nil
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

var operatorNodePrecedence = map[NodeType]int{
	NT_And: 0,
	NT_Or:  0,

	NT_Equal:        1,
	NT_NotEqual:     1,
	NT_Lower:        1,
	NT_Greater:      1,
	NT_LowerEqual:   1,
	NT_GreaterEqual: 1,

	NT_Add:      2,
	NT_Subtract: 2,

	NT_Multiply: 3,
	NT_Divide:   3,
	NT_Power:    4,
	NT_Modulo:   4,
}
