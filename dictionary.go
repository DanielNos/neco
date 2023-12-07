package main

var KEYWORDS = map[string]TokenType {
	"const": TT_KW_const,
	"var": TT_KW_var,
	"bool": TT_KW_bool,
	"int": TT_KW_int,
	"flt": TT_KW_flt,
	"str": TT_KW_str,

	"=": TT_KW_Assign,
}

var DELIMITERS = map[rune]bool {
	'(': true,
	')': true,
	'[': true,
	']': true,
	'{': true,
	'}': true,
}
