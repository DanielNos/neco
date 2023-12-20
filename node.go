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
)

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
