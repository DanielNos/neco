package main

type NodeValue interface{}

type Node struct {
	position *CodePos
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
}

type ModuleNode struct {
	name string
	statements *ScopeNode
}

type ScopeNode struct {
	id int
	statements []*Node
}

type VariableDeclareNode struct {
	dataType DataType
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
