package main

var KEYWORDS = map[string]TokenType {
	"global": TT_KW_global,
	"const": TT_KW_const,
	"var": TT_KW_var,
	"bool": TT_KW_bool,
	"int": TT_KW_int,
	"flt": TT_KW_flt,
	"str": TT_KW_str,

	"loop": TT_KW_loop,
	"while": TT_KW_while,
	"for": TT_KW_for,
	"forEach": TT_KW_forEach,
	"continue": TT_KW_continue,
	"break": TT_KW_break,

	"pub": TT_KW_pub,
	"fun": TT_KW_fun,
	"return": TT_KW_return,

	"struct": TT_KW_struct,
	"enum": TT_KW_enum,
	"class": TT_KW_class,

	"if": TT_KW_if,
	"else": TT_KW_else,
}

var DELIMITERS = map[rune]TokenType {
	'(': TT_DL_ParenthesisOpen,
	')': TT_DL_ParenthesisClose,
	'[': TT_DL_BracketOpen,
	']': TT_DL_BracketClose,
	'{': TT_DL_BraceOpen,
	'}': TT_DL_BraceClose,
	',': TT_DL_Comma,
}
