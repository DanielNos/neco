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
)

var NodeTypeToString = map[NodeType]string {
	NT_Module: "Module",
	NT_Scope: "Scope",
	NT_StatementList: "StatementList",
	NT_VariableDeclare: "VariableDeclare",
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
