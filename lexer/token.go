package lexer

import (
	"fmt"
	"neco/dataStructures"
)

type TokenType int8

const (
	TT_EndOfCommand TokenType = iota
	TT_StartOfFile
	TT_EndOfFile

	TT_Identifier

	TT_DL_ParenthesisOpen
	TT_DL_ParenthesisClose
	TT_DL_BracketOpen
	TT_DL_BracketClose
	TT_DL_BraceOpen
	TT_DL_BraceClose
	TT_DL_Comma

	TT_OP_Not
	TT_OP_And
	TT_OP_Or

	TT_OP_Dot

	TT_OP_Add
	TT_OP_Subtract
	TT_OP_Multiply
	TT_OP_Divide
	TT_OP_Power
	TT_OP_Modulo

	TT_OP_Equal
	TT_OP_NotEqual
	TT_OP_Lower
	TT_OP_LowerEqual
	TT_OP_Greater
	TT_OP_GreaterEqual

	TT_LT_Bool
	TT_LT_Int
	TT_LT_Float
	TT_LT_String

	TT_KW_global
	TT_KW_const
	TT_KW_var
	TT_KW_bool
	TT_KW_int
	TT_KW_flt
	TT_KW_str
	TT_KW_list

	TT_KW_Assign
	TT_KW_AddAssign
	TT_KW_SubtractAssign
	TT_KW_MultiplyAssign
	TT_KW_DivideAssign
	TT_KW_PowerAssign
	TT_KW_ModuloAssign

	TT_KW_loop
	TT_KW_while
	TT_KW_for
	TT_KW_forEach
	TT_KW_continue
	TT_KW_break
	TT_KW_in

	TT_KW_pub
	TT_KW_fun
	TT_KW_returns
	TT_KW_return

	TT_KW_struct
	TT_KW_enum
	TT_KW_class

	TT_KW_if
	TT_KW_elif
	TT_KW_else
)

var TokenTypeToString = map[TokenType]string{
	TT_EndOfCommand: "EndOfCommand",
	TT_StartOfFile:  "StartOfFile",
	TT_EndOfFile:    "EndOfFile",

	TT_Identifier: "Identifier",

	TT_DL_ParenthesisOpen:  "(",
	TT_DL_ParenthesisClose: ")",
	TT_DL_BracketOpen:      "[",
	TT_DL_BracketClose:     "]",
	TT_DL_BraceOpen:        "{",
	TT_DL_BraceClose:       "}",
	TT_DL_Comma:            ",",

	TT_OP_And: "&",
	TT_OP_Or:  "|",
	TT_OP_Not: "!",

	TT_OP_Dot: ".",

	TT_OP_Add:      "+",
	TT_OP_Subtract: "-",
	TT_OP_Multiply: "*",
	TT_OP_Divide:   "/",
	TT_OP_Power:    "^",
	TT_OP_Modulo:   "%",

	TT_OP_Equal:        "==",
	TT_OP_NotEqual:     "!=",
	TT_OP_Lower:        "<",
	TT_OP_LowerEqual:   "<=",
	TT_OP_Greater:      ">",
	TT_OP_GreaterEqual: ">=",

	TT_LT_Bool:   "Bool",
	TT_LT_Int:    "Int",
	TT_LT_Float:  "Float",
	TT_LT_String: "String",

	TT_KW_global: "global",
	TT_KW_const:  "const",
	TT_KW_var:    "var",
	TT_KW_bool:   "bool",
	TT_KW_int:    "int",
	TT_KW_flt:    "flt",
	TT_KW_str:    "str",
	TT_KW_list:   "list",

	TT_KW_Assign:         "=",
	TT_KW_AddAssign:      "+=",
	TT_KW_SubtractAssign: "-=",
	TT_KW_MultiplyAssign: "*=",
	TT_KW_DivideAssign:   "/=",
	TT_KW_PowerAssign:    "^=",
	TT_KW_ModuloAssign:   "%=",

	TT_KW_loop:     "loop",
	TT_KW_while:    "while",
	TT_KW_for:      "for",
	TT_KW_forEach:  "forEach",
	TT_KW_continue: "continue",
	TT_KW_break:    "break",
	TT_KW_in:       "in",

	TT_KW_pub:     "pub",
	TT_KW_fun:     "fun",
	TT_KW_returns: "->",
	TT_KW_return:  "return",

	TT_KW_struct: "struct",
	TT_KW_enum:   "enum",
	TT_KW_class:  "class",

	TT_KW_if:   "if",
	TT_KW_else: "else",
	TT_KW_elif: "elif",
}

func (tt TokenType) String() string {
	return TokenTypeToString[tt]
}

func (tt TokenType) IsVariableType() bool {
	return tt >= TT_KW_var && tt <= TT_KW_list
}

func (tt TokenType) IsLiteral() bool {
	return tt >= TT_LT_Bool && tt <= TT_LT_String
}

func (tt TokenType) IsOperator() bool {
	return tt >= TT_OP_And && tt <= TT_OP_GreaterEqual
}

func (tt TokenType) IsBinaryOperator() bool {
	return tt >= TT_OP_And && tt <= TT_OP_GreaterEqual
}

func (tt TokenType) IsUnaryOperator() bool {
	return tt == TT_OP_Not || tt == TT_OP_Subtract
}

func (tt TokenType) IsAssignKeyword() bool {
	return tt >= TT_KW_Assign && tt <= TT_KW_ModuloAssign
}

func (tt TokenType) IsDelimiter() bool {
	return tt >= TT_DL_ParenthesisOpen && tt <= TT_DL_Comma
}

type Token struct {
	Position  *dataStructures.CodePos
	TokenType TokenType
	Value     string
}

func (t *Token) String() string {
	switch t.TokenType {
	case TT_EndOfCommand:
		if len(t.Value) == 0 {
			return "\\n"
		}
		return ";"
	case TT_StartOfFile:
		return "SOF"
	case TT_EndOfFile:
		return "EOF"
	case TT_Identifier:
		return t.Value
	case TT_LT_Bool:
		if t.Value == "0" {
			return "false"
		}
		return "true"
	default:
		if t.TokenType.IsLiteral() {
			return t.Value
		}
		return TokenTypeToString[t.TokenType]
	}
}

func (t *Token) TableString() string {
	// Print file
	message := fmt.Sprintf("%s  ", *t.Position.File)

	if t.Position.StartLine < 10 {
		message = fmt.Sprintf(" %s", message)
	}

	// Print start line and char
	message = fmt.Sprintf("%s%d:%d  ", message, t.Position.StartLine, t.Position.StartChar)

	if t.Position.StartChar < 10 {
		message = fmt.Sprintf("%s ", message)
	}

	// Print end line and char
	message = fmt.Sprintf("%s%d:%d", message, t.Position.EndLine, t.Position.EndChar)

	if t.Position.EndChar < 10 {
		message = fmt.Sprintf("%s ", message)
	}

	// Print token type
	message = fmt.Sprintf("%s  %s", message, t.TokenType)

	for len(message) < 39 {
		message = fmt.Sprintf("%s ", message)
	}

	// Print token value
	return fmt.Sprintf("%s%s", message, t.Value)
}
