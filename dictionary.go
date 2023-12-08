package main

var KEYWORDS = map[string]TokenType {
	"const": TT_KW_const,
	"var": TT_KW_var,
	"bool": TT_KW_bool,
	"int": TT_KW_int,
	"flt": TT_KW_flt,
	"str": TT_KW_str,

	"=": TT_KW_Assign,

	"loop": TT_KW_loop,
	"while": TT_KW_while,
	"for": TT_KW_for,
	"forEach": TT_KW_forEach,
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
