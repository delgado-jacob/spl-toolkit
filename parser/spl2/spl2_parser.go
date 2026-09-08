// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

import "strings"

// ANTLR also emits the header into the visitor interface file.
var _ = strings.EqualFold

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type SPL2Parser struct {
	*antlr.BaseParser
}

var SPL2ParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func spl2parserParserInit() {
	staticData := &SPL2ParserParserStaticData
	staticData.LiteralNames = []string{
		"", "", "", "'search'", "'index'", "'eval'", "", "'fields'", "'table'",
		"'AS'", "", "'rex'", "'makeresults'", "'spl1'", "'import'", "'export'",
		"'function'", "'return'", "'AND'", "'OR'", "'XOR'", "'NOT'", "'BETWEEN'",
		"'IN'", "'LIKE'", "'IS'", "'null'", "'NULL'", "", "", "'->'", "'<='",
		"'>='", "'!='", "'=='", "", "'<'", "'>'", "'+'", "'-'", "'*'", "", "",
		"'/'", "'%'", "", "','", "':'", "'.'", "'('", "')'", "'['", "']'", "'{'",
		"'}'", "';'",
	}
	staticData.SymbolicNames = []string{
		"", "FROM", "SELECT", "SEARCH", "INDEX", "EVAL", "WHERE", "FIELDS",
		"TABLE", "AS", "GROUP", "REX", "MAKERESULTS", "SPL1", "IMPORT", "EXPORT",
		"FUNCTION", "RETURN", "AND", "OR", "XOR", "NOT", "BETWEEN", "IN", "LIKE",
		"IS", "NULL", "NULL_TEST", "BOOLEAN", "TYPE", "ARROW", "LE", "GE", "NE",
		"EQ", "ASSIGN", "LT", "GT", "PLUS", "MINUS", "STAR", "LINE_COMMENT",
		"BLOCK_COMMENT", "SLASH", "MOD", "PIPE", "COMMA", "COLON", "DOT", "LPAREN",
		"RPAREN", "LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "SEMI", "RAW_STRING",
		"DQUOTE", "SQUOTE", "BACKTICK", "NUMBER", "LOCAL", "IDENTIFIER", "NL",
		"WS", "STRING_END", "STRING_INTERPOLATION", "STRING_TEXT", "STRING_DOLLAR",
		"NAME_END", "NAME_INTERPOLATION", "NAME_TEXT", "NAME_DOLLAR", "EMBEDDED_END",
		"EMBEDDED_TEXT", "REGEX", "RX_WS",
	}
	staticData.RuleNames = []string{
		"query", "pipeline", "start", "command", "fromCommand", "selectCommand",
		"projection", "dataset", "generator", "evalCommand", "assignment", "whereCommand",
		"fieldsCommand", "rexCommand", "rexOption", "embeddedCommand", "embeddedText",
		"moduleSuffix", "moduleDeclaration", "searchCommand", "implicitSearch",
		"searchExpression", "searchXor", "searchAnd", "searchOr", "searchNot",
		"searchAtom", "searchValue", "expression", "xorExpression", "orExpression",
		"andExpression", "notExpression", "predicate", "logicalAnd", "logicalOr",
		"logicalXor", "logicalNot", "betweenOperator", "betweenConjunction",
		"comparison", "additive", "multiplicative", "unary", "access", "accessPart",
		"primary", "call", "arguments", "namedArgument", "literal", "stringLiteral",
		"quotedName", "fieldTemplate", "fieldName", "identifier", "array", "object",
		"objectEntry", "objectKey", "searchLiteral", "lambdaExpression", "lambdaParameter",
		"lambdaBlock",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 76, 741, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36,
		2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7, 41, 2,
		42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46, 2, 47,
		7, 47, 2, 48, 7, 48, 2, 49, 7, 49, 2, 50, 7, 50, 2, 51, 7, 51, 2, 52, 7,
		52, 2, 53, 7, 53, 2, 54, 7, 54, 2, 55, 7, 55, 2, 56, 7, 56, 2, 57, 7, 57,
		2, 58, 7, 58, 2, 59, 7, 59, 2, 60, 7, 60, 2, 61, 7, 61, 2, 62, 7, 62, 2,
		63, 7, 63, 1, 0, 5, 0, 130, 8, 0, 10, 0, 12, 0, 133, 9, 0, 1, 0, 1, 0,
		3, 0, 137, 8, 0, 1, 0, 3, 0, 140, 8, 0, 1, 0, 5, 0, 143, 8, 0, 10, 0, 12,
		0, 146, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 5, 1, 152, 8, 1, 10, 1, 12, 1, 155,
		9, 1, 1, 1, 1, 1, 5, 1, 159, 8, 1, 10, 1, 12, 1, 162, 9, 1, 1, 1, 5, 1,
		165, 8, 1, 10, 1, 12, 1, 168, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		3, 2, 176, 8, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3,
		186, 8, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 5, 5, 195, 8, 5, 10,
		5, 12, 5, 198, 9, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 3, 6, 206, 8,
		6, 1, 7, 1, 7, 3, 7, 210, 8, 7, 1, 8, 1, 8, 3, 8, 214, 8, 8, 1, 9, 1, 9,
		1, 9, 1, 9, 5, 9, 220, 8, 9, 10, 9, 12, 9, 223, 9, 9, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 3, 12, 234, 8, 12, 1, 12,
		1, 12, 1, 12, 5, 12, 239, 8, 12, 10, 12, 12, 12, 242, 9, 12, 1, 13, 1,
		13, 5, 13, 246, 8, 13, 10, 13, 12, 13, 249, 9, 13, 1, 13, 1, 13, 1, 13,
		3, 13, 254, 8, 13, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 260, 8, 14, 1, 15,
		3, 15, 263, 8, 15, 1, 15, 1, 15, 1, 16, 1, 16, 3, 16, 269, 8, 16, 1, 16,
		1, 16, 1, 17, 1, 17, 5, 17, 275, 8, 17, 10, 17, 12, 17, 278, 9, 17, 1,
		18, 1, 18, 1, 18, 1, 18, 1, 18, 3, 18, 285, 8, 18, 1, 18, 5, 18, 288, 8,
		18, 10, 18, 12, 18, 291, 9, 18, 1, 19, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20,
		1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 5, 22, 304, 8, 22, 10, 22, 12, 22, 307,
		9, 22, 1, 23, 1, 23, 3, 23, 311, 8, 23, 1, 23, 5, 23, 314, 8, 23, 10, 23,
		12, 23, 317, 9, 23, 1, 24, 1, 24, 1, 24, 5, 24, 322, 8, 24, 10, 24, 12,
		24, 325, 9, 24, 1, 25, 1, 25, 1, 25, 3, 25, 330, 8, 25, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 4, 26, 346, 8, 26, 11, 26, 12, 26, 347, 1, 26, 1, 26, 1, 26, 3,
		26, 353, 8, 26, 1, 27, 1, 27, 1, 27, 1, 27, 1, 27, 3, 27, 360, 8, 27, 1,
		28, 1, 28, 3, 28, 364, 8, 28, 1, 29, 1, 29, 1, 29, 1, 29, 5, 29, 370, 8,
		29, 10, 29, 12, 29, 373, 9, 29, 1, 30, 1, 30, 1, 30, 1, 30, 5, 30, 379,
		8, 30, 10, 30, 12, 30, 382, 9, 30, 1, 31, 1, 31, 1, 31, 1, 31, 5, 31, 388,
		8, 31, 10, 31, 12, 31, 391, 9, 31, 1, 32, 1, 32, 1, 32, 1, 32, 3, 32, 397,
		8, 32, 1, 33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 404, 8, 33, 1, 33, 1,
		33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 412, 8, 33, 1, 33, 1, 33, 1, 33,
		1, 33, 1, 33, 5, 33, 419, 8, 33, 10, 33, 12, 33, 422, 9, 33, 1, 33, 1,
		33, 1, 33, 3, 33, 427, 8, 33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 433, 8,
		33, 1, 33, 1, 33, 3, 33, 437, 8, 33, 1, 33, 3, 33, 440, 8, 33, 3, 33, 442,
		8, 33, 1, 34, 1, 34, 1, 34, 3, 34, 447, 8, 34, 1, 35, 1, 35, 1, 35, 3,
		35, 452, 8, 35, 1, 36, 1, 36, 1, 36, 3, 36, 457, 8, 36, 1, 37, 1, 37, 1,
		37, 3, 37, 462, 8, 37, 1, 38, 1, 38, 1, 38, 3, 38, 467, 8, 38, 1, 39, 1,
		39, 1, 39, 3, 39, 472, 8, 39, 1, 40, 1, 40, 1, 41, 1, 41, 1, 41, 5, 41,
		479, 8, 41, 10, 41, 12, 41, 482, 9, 41, 1, 42, 1, 42, 1, 42, 5, 42, 487,
		8, 42, 10, 42, 12, 42, 490, 9, 42, 1, 43, 1, 43, 1, 43, 3, 43, 495, 8,
		43, 1, 44, 1, 44, 5, 44, 499, 8, 44, 10, 44, 12, 44, 502, 9, 44, 1, 45,
		1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 3, 45, 510, 8, 45, 1, 46, 1, 46, 1,
		46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 3, 46, 523,
		8, 46, 1, 47, 1, 47, 1, 47, 3, 47, 528, 8, 47, 1, 47, 1, 47, 1, 48, 1,
		48, 1, 48, 5, 48, 535, 8, 48, 10, 48, 12, 48, 538, 9, 48, 1, 48, 1, 48,
		1, 48, 5, 48, 543, 8, 48, 10, 48, 12, 48, 546, 9, 48, 1, 48, 1, 48, 5,
		48, 550, 8, 48, 10, 48, 12, 48, 553, 9, 48, 3, 48, 555, 8, 48, 1, 49, 1,
		49, 1, 49, 1, 49, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 3, 50, 566, 8, 50,
		1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 1, 51, 5, 51, 575, 8, 51, 10,
		51, 12, 51, 578, 9, 51, 1, 51, 1, 51, 1, 52, 1, 52, 4, 52, 584, 8, 52,
		11, 52, 12, 52, 585, 1, 52, 1, 52, 1, 53, 1, 53, 5, 53, 592, 8, 53, 10,
		53, 12, 53, 595, 9, 53, 1, 53, 1, 53, 1, 53, 1, 53, 1, 53, 1, 53, 1, 53,
		1, 53, 1, 53, 5, 53, 606, 8, 53, 10, 53, 12, 53, 609, 9, 53, 1, 53, 1,
		53, 1, 54, 1, 54, 3, 54, 615, 8, 54, 1, 55, 1, 55, 1, 55, 3, 55, 620, 8,
		55, 1, 56, 1, 56, 1, 56, 1, 56, 5, 56, 626, 8, 56, 10, 56, 12, 56, 629,
		9, 56, 1, 56, 3, 56, 632, 8, 56, 3, 56, 634, 8, 56, 1, 56, 1, 56, 1, 57,
		1, 57, 1, 57, 1, 57, 5, 57, 642, 8, 57, 10, 57, 12, 57, 645, 9, 57, 1,
		57, 3, 57, 648, 8, 57, 3, 57, 650, 8, 57, 1, 57, 1, 57, 1, 58, 1, 58, 1,
		58, 1, 58, 1, 59, 1, 59, 3, 59, 660, 8, 59, 1, 60, 1, 60, 1, 61, 1, 61,
		1, 61, 1, 61, 1, 61, 5, 61, 669, 8, 61, 10, 61, 12, 61, 672, 9, 61, 3,
		61, 674, 8, 61, 1, 61, 3, 61, 677, 8, 61, 1, 61, 1, 61, 1, 61, 3, 61, 682,
		8, 61, 1, 62, 1, 62, 1, 62, 3, 62, 687, 8, 62, 1, 62, 1, 62, 3, 62, 691,
		8, 62, 1, 62, 1, 62, 3, 62, 695, 8, 62, 3, 62, 697, 8, 62, 1, 63, 1, 63,
		5, 63, 701, 8, 63, 10, 63, 12, 63, 704, 9, 63, 1, 63, 1, 63, 1, 63, 1,
		63, 1, 63, 5, 63, 711, 8, 63, 10, 63, 12, 63, 714, 9, 63, 1, 63, 4, 63,
		717, 8, 63, 11, 63, 12, 63, 718, 3, 63, 721, 8, 63, 5, 63, 723, 8, 63,
		10, 63, 12, 63, 726, 9, 63, 1, 63, 1, 63, 1, 63, 3, 63, 731, 8, 63, 1,
		63, 5, 63, 734, 8, 63, 10, 63, 12, 63, 737, 9, 63, 1, 63, 1, 63, 1, 63,
		2, 276, 289, 0, 64, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26,
		28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62,
		64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 90, 92, 94, 96, 98,
		100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124, 126, 0,
		6, 1, 0, 7, 8, 1, 0, 38, 39, 1, 0, 26, 27, 1, 0, 31, 37, 2, 0, 40, 40,
		43, 44, 1, 0, 71, 72, 803, 0, 131, 1, 0, 0, 0, 2, 149, 1, 0, 0, 0, 4, 175,
		1, 0, 0, 0, 6, 185, 1, 0, 0, 0, 8, 187, 1, 0, 0, 0, 10, 190, 1, 0, 0, 0,
		12, 202, 1, 0, 0, 0, 14, 209, 1, 0, 0, 0, 16, 211, 1, 0, 0, 0, 18, 215,
		1, 0, 0, 0, 20, 224, 1, 0, 0, 0, 22, 228, 1, 0, 0, 0, 24, 231, 1, 0, 0,
		0, 26, 243, 1, 0, 0, 0, 28, 255, 1, 0, 0, 0, 30, 262, 1, 0, 0, 0, 32, 266,
		1, 0, 0, 0, 34, 272, 1, 0, 0, 0, 36, 284, 1, 0, 0, 0, 38, 292, 1, 0, 0,
		0, 40, 295, 1, 0, 0, 0, 42, 298, 1, 0, 0, 0, 44, 300, 1, 0, 0, 0, 46, 308,
		1, 0, 0, 0, 48, 318, 1, 0, 0, 0, 50, 329, 1, 0, 0, 0, 52, 352, 1, 0, 0,
		0, 54, 359, 1, 0, 0, 0, 56, 363, 1, 0, 0, 0, 58, 365, 1, 0, 0, 0, 60, 374,
		1, 0, 0, 0, 62, 383, 1, 0, 0, 0, 64, 396, 1, 0, 0, 0, 66, 398, 1, 0, 0,
		0, 68, 446, 1, 0, 0, 0, 70, 451, 1, 0, 0, 0, 72, 456, 1, 0, 0, 0, 74, 461,
		1, 0, 0, 0, 76, 466, 1, 0, 0, 0, 78, 471, 1, 0, 0, 0, 80, 473, 1, 0, 0,
		0, 82, 475, 1, 0, 0, 0, 84, 483, 1, 0, 0, 0, 86, 494, 1, 0, 0, 0, 88, 496,
		1, 0, 0, 0, 90, 509, 1, 0, 0, 0, 92, 522, 1, 0, 0, 0, 94, 524, 1, 0, 0,
		0, 96, 554, 1, 0, 0, 0, 98, 556, 1, 0, 0, 0, 100, 565, 1, 0, 0, 0, 102,
		567, 1, 0, 0, 0, 104, 581, 1, 0, 0, 0, 106, 589, 1, 0, 0, 0, 108, 614,
		1, 0, 0, 0, 110, 619, 1, 0, 0, 0, 112, 621, 1, 0, 0, 0, 114, 637, 1, 0,
		0, 0, 116, 653, 1, 0, 0, 0, 118, 659, 1, 0, 0, 0, 120, 661, 1, 0, 0, 0,
		122, 676, 1, 0, 0, 0, 124, 683, 1, 0, 0, 0, 126, 698, 1, 0, 0, 0, 128,
		130, 5, 63, 0, 0, 129, 128, 1, 0, 0, 0, 130, 133, 1, 0, 0, 0, 131, 129,
		1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132, 139, 1, 0, 0, 0, 133, 131, 1, 0,
		0, 0, 134, 136, 3, 2, 1, 0, 135, 137, 3, 34, 17, 0, 136, 135, 1, 0, 0,
		0, 136, 137, 1, 0, 0, 0, 137, 140, 1, 0, 0, 0, 138, 140, 3, 36, 18, 0,
		139, 134, 1, 0, 0, 0, 139, 138, 1, 0, 0, 0, 140, 144, 1, 0, 0, 0, 141,
		143, 5, 63, 0, 0, 142, 141, 1, 0, 0, 0, 143, 146, 1, 0, 0, 0, 144, 142,
		1, 0, 0, 0, 144, 145, 1, 0, 0, 0, 145, 147, 1, 0, 0, 0, 146, 144, 1, 0,
		0, 0, 147, 148, 5, 0, 0, 1, 148, 1, 1, 0, 0, 0, 149, 166, 3, 4, 2, 0, 150,
		152, 5, 63, 0, 0, 151, 150, 1, 0, 0, 0, 152, 155, 1, 0, 0, 0, 153, 151,
		1, 0, 0, 0, 153, 154, 1, 0, 0, 0, 154, 156, 1, 0, 0, 0, 155, 153, 1, 0,
		0, 0, 156, 160, 5, 45, 0, 0, 157, 159, 5, 63, 0, 0, 158, 157, 1, 0, 0,
		0, 159, 162, 1, 0, 0, 0, 160, 158, 1, 0, 0, 0, 160, 161, 1, 0, 0, 0, 161,
		163, 1, 0, 0, 0, 162, 160, 1, 0, 0, 0, 163, 165, 3, 6, 3, 0, 164, 153,
		1, 0, 0, 0, 165, 168, 1, 0, 0, 0, 166, 164, 1, 0, 0, 0, 166, 167, 1, 0,
		0, 0, 167, 3, 1, 0, 0, 0, 168, 166, 1, 0, 0, 0, 169, 176, 3, 8, 4, 0, 170,
		176, 3, 10, 5, 0, 171, 176, 3, 38, 19, 0, 172, 176, 3, 40, 20, 0, 173,
		176, 3, 16, 8, 0, 174, 176, 3, 30, 15, 0, 175, 169, 1, 0, 0, 0, 175, 170,
		1, 0, 0, 0, 175, 171, 1, 0, 0, 0, 175, 172, 1, 0, 0, 0, 175, 173, 1, 0,
		0, 0, 175, 174, 1, 0, 0, 0, 176, 5, 1, 0, 0, 0, 177, 186, 3, 18, 9, 0,
		178, 186, 3, 22, 11, 0, 179, 186, 3, 24, 12, 0, 180, 186, 3, 8, 4, 0, 181,
		186, 3, 10, 5, 0, 182, 186, 3, 38, 19, 0, 183, 186, 3, 26, 13, 0, 184,
		186, 3, 30, 15, 0, 185, 177, 1, 0, 0, 0, 185, 178, 1, 0, 0, 0, 185, 179,
		1, 0, 0, 0, 185, 180, 1, 0, 0, 0, 185, 181, 1, 0, 0, 0, 185, 182, 1, 0,
		0, 0, 185, 183, 1, 0, 0, 0, 185, 184, 1, 0, 0, 0, 186, 7, 1, 0, 0, 0, 187,
		188, 5, 1, 0, 0, 188, 189, 3, 14, 7, 0, 189, 9, 1, 0, 0, 0, 190, 191, 5,
		2, 0, 0, 191, 196, 3, 12, 6, 0, 192, 193, 5, 46, 0, 0, 193, 195, 3, 12,
		6, 0, 194, 192, 1, 0, 0, 0, 195, 198, 1, 0, 0, 0, 196, 194, 1, 0, 0, 0,
		196, 197, 1, 0, 0, 0, 197, 199, 1, 0, 0, 0, 198, 196, 1, 0, 0, 0, 199,
		200, 5, 1, 0, 0, 200, 201, 3, 14, 7, 0, 201, 11, 1, 0, 0, 0, 202, 205,
		3, 56, 28, 0, 203, 204, 5, 9, 0, 0, 204, 206, 3, 110, 55, 0, 205, 203,
		1, 0, 0, 0, 205, 206, 1, 0, 0, 0, 206, 13, 1, 0, 0, 0, 207, 210, 3, 110,
		55, 0, 208, 210, 3, 112, 56, 0, 209, 207, 1, 0, 0, 0, 209, 208, 1, 0, 0,
		0, 210, 15, 1, 0, 0, 0, 211, 213, 5, 12, 0, 0, 212, 214, 5, 60, 0, 0, 213,
		212, 1, 0, 0, 0, 213, 214, 1, 0, 0, 0, 214, 17, 1, 0, 0, 0, 215, 216, 5,
		5, 0, 0, 216, 221, 3, 20, 10, 0, 217, 218, 5, 46, 0, 0, 218, 220, 3, 20,
		10, 0, 219, 217, 1, 0, 0, 0, 220, 223, 1, 0, 0, 0, 221, 219, 1, 0, 0, 0,
		221, 222, 1, 0, 0, 0, 222, 19, 1, 0, 0, 0, 223, 221, 1, 0, 0, 0, 224, 225,
		3, 108, 54, 0, 225, 226, 5, 35, 0, 0, 226, 227, 3, 56, 28, 0, 227, 21,
		1, 0, 0, 0, 228, 229, 5, 6, 0, 0, 229, 230, 3, 56, 28, 0, 230, 23, 1, 0,
		0, 0, 231, 233, 7, 0, 0, 0, 232, 234, 7, 1, 0, 0, 233, 232, 1, 0, 0, 0,
		233, 234, 1, 0, 0, 0, 234, 235, 1, 0, 0, 0, 235, 240, 3, 110, 55, 0, 236,
		237, 5, 46, 0, 0, 237, 239, 3, 110, 55, 0, 238, 236, 1, 0, 0, 0, 239, 242,
		1, 0, 0, 0, 240, 238, 1, 0, 0, 0, 240, 241, 1, 0, 0, 0, 241, 25, 1, 0,
		0, 0, 242, 240, 1, 0, 0, 0, 243, 247, 5, 11, 0, 0, 244, 246, 3, 28, 14,
		0, 245, 244, 1, 0, 0, 0, 246, 249, 1, 0, 0, 0, 247, 245, 1, 0, 0, 0, 247,
		248, 1, 0, 0, 0, 248, 253, 1, 0, 0, 0, 249, 247, 1, 0, 0, 0, 250, 254,
		5, 75, 0, 0, 251, 254, 3, 102, 51, 0, 252, 254, 5, 56, 0, 0, 253, 250,
		1, 0, 0, 0, 253, 251, 1, 0, 0, 0, 253, 252, 1, 0, 0, 0, 254, 27, 1, 0,
		0, 0, 255, 256, 5, 62, 0, 0, 256, 259, 5, 35, 0, 0, 257, 260, 3, 110, 55,
		0, 258, 260, 5, 60, 0, 0, 259, 257, 1, 0, 0, 0, 259, 258, 1, 0, 0, 0, 260,
		29, 1, 0, 0, 0, 261, 263, 5, 13, 0, 0, 262, 261, 1, 0, 0, 0, 262, 263,
		1, 0, 0, 0, 263, 264, 1, 0, 0, 0, 264, 265, 3, 32, 16, 0, 265, 31, 1, 0,
		0, 0, 266, 268, 5, 59, 0, 0, 267, 269, 5, 74, 0, 0, 268, 267, 1, 0, 0,
		0, 268, 269, 1, 0, 0, 0, 269, 270, 1, 0, 0, 0, 270, 271, 5, 73, 0, 0, 271,
		33, 1, 0, 0, 0, 272, 276, 5, 55, 0, 0, 273, 275, 9, 0, 0, 0, 274, 273,
		1, 0, 0, 0, 275, 278, 1, 0, 0, 0, 276, 277, 1, 0, 0, 0, 276, 274, 1, 0,
		0, 0, 277, 35, 1, 0, 0, 0, 278, 276, 1, 0, 0, 0, 279, 285, 5, 14, 0, 0,
		280, 285, 5, 15, 0, 0, 281, 285, 5, 16, 0, 0, 282, 283, 5, 61, 0, 0, 283,
		285, 5, 35, 0, 0, 284, 279, 1, 0, 0, 0, 284, 280, 1, 0, 0, 0, 284, 281,
		1, 0, 0, 0, 284, 282, 1, 0, 0, 0, 285, 289, 1, 0, 0, 0, 286, 288, 9, 0,
		0, 0, 287, 286, 1, 0, 0, 0, 288, 291, 1, 0, 0, 0, 289, 290, 1, 0, 0, 0,
		289, 287, 1, 0, 0, 0, 290, 37, 1, 0, 0, 0, 291, 289, 1, 0, 0, 0, 292, 293,
		5, 3, 0, 0, 293, 294, 3, 42, 21, 0, 294, 39, 1, 0, 0, 0, 295, 296, 4, 20,
		0, 0, 296, 297, 3, 42, 21, 0, 297, 41, 1, 0, 0, 0, 298, 299, 3, 44, 22,
		0, 299, 43, 1, 0, 0, 0, 300, 305, 3, 46, 23, 0, 301, 302, 5, 20, 0, 0,
		302, 304, 3, 46, 23, 0, 303, 301, 1, 0, 0, 0, 304, 307, 1, 0, 0, 0, 305,
		303, 1, 0, 0, 0, 305, 306, 1, 0, 0, 0, 306, 45, 1, 0, 0, 0, 307, 305, 1,
		0, 0, 0, 308, 315, 3, 48, 24, 0, 309, 311, 5, 18, 0, 0, 310, 309, 1, 0,
		0, 0, 310, 311, 1, 0, 0, 0, 311, 312, 1, 0, 0, 0, 312, 314, 3, 48, 24,
		0, 313, 310, 1, 0, 0, 0, 314, 317, 1, 0, 0, 0, 315, 313, 1, 0, 0, 0, 315,
		316, 1, 0, 0, 0, 316, 47, 1, 0, 0, 0, 317, 315, 1, 0, 0, 0, 318, 323, 3,
		50, 25, 0, 319, 320, 5, 19, 0, 0, 320, 322, 3, 50, 25, 0, 321, 319, 1,
		0, 0, 0, 322, 325, 1, 0, 0, 0, 323, 321, 1, 0, 0, 0, 323, 324, 1, 0, 0,
		0, 324, 49, 1, 0, 0, 0, 325, 323, 1, 0, 0, 0, 326, 327, 5, 21, 0, 0, 327,
		330, 3, 50, 25, 0, 328, 330, 3, 52, 26, 0, 329, 326, 1, 0, 0, 0, 329, 328,
		1, 0, 0, 0, 330, 51, 1, 0, 0, 0, 331, 332, 5, 49, 0, 0, 332, 333, 3, 42,
		21, 0, 333, 334, 5, 50, 0, 0, 334, 353, 1, 0, 0, 0, 335, 336, 3, 110, 55,
		0, 336, 337, 3, 80, 40, 0, 337, 338, 3, 54, 27, 0, 338, 353, 1, 0, 0, 0,
		339, 340, 3, 110, 55, 0, 340, 341, 5, 23, 0, 0, 341, 342, 5, 49, 0, 0,
		342, 345, 3, 54, 27, 0, 343, 344, 5, 46, 0, 0, 344, 346, 3, 54, 27, 0,
		345, 343, 1, 0, 0, 0, 346, 347, 1, 0, 0, 0, 347, 345, 1, 0, 0, 0, 347,
		348, 1, 0, 0, 0, 348, 349, 1, 0, 0, 0, 349, 350, 5, 50, 0, 0, 350, 353,
		1, 0, 0, 0, 351, 353, 3, 54, 27, 0, 352, 331, 1, 0, 0, 0, 352, 335, 1,
		0, 0, 0, 352, 339, 1, 0, 0, 0, 352, 351, 1, 0, 0, 0, 353, 53, 1, 0, 0,
		0, 354, 360, 3, 110, 55, 0, 355, 360, 5, 60, 0, 0, 356, 360, 3, 102, 51,
		0, 357, 360, 5, 56, 0, 0, 358, 360, 5, 40, 0, 0, 359, 354, 1, 0, 0, 0,
		359, 355, 1, 0, 0, 0, 359, 356, 1, 0, 0, 0, 359, 357, 1, 0, 0, 0, 359,
		358, 1, 0, 0, 0, 360, 55, 1, 0, 0, 0, 361, 364, 3, 122, 61, 0, 362, 364,
		3, 58, 29, 0, 363, 361, 1, 0, 0, 0, 363, 362, 1, 0, 0, 0, 364, 57, 1, 0,
		0, 0, 365, 371, 3, 60, 30, 0, 366, 367, 3, 72, 36, 0, 367, 368, 3, 60,
		30, 0, 368, 370, 1, 0, 0, 0, 369, 366, 1, 0, 0, 0, 370, 373, 1, 0, 0, 0,
		371, 369, 1, 0, 0, 0, 371, 372, 1, 0, 0, 0, 372, 59, 1, 0, 0, 0, 373, 371,
		1, 0, 0, 0, 374, 380, 3, 62, 31, 0, 375, 376, 3, 70, 35, 0, 376, 377, 3,
		62, 31, 0, 377, 379, 1, 0, 0, 0, 378, 375, 1, 0, 0, 0, 379, 382, 1, 0,
		0, 0, 380, 378, 1, 0, 0, 0, 380, 381, 1, 0, 0, 0, 381, 61, 1, 0, 0, 0,
		382, 380, 1, 0, 0, 0, 383, 389, 3, 64, 32, 0, 384, 385, 3, 68, 34, 0, 385,
		386, 3, 64, 32, 0, 386, 388, 1, 0, 0, 0, 387, 384, 1, 0, 0, 0, 388, 391,
		1, 0, 0, 0, 389, 387, 1, 0, 0, 0, 389, 390, 1, 0, 0, 0, 390, 63, 1, 0,
		0, 0, 391, 389, 1, 0, 0, 0, 392, 397, 3, 66, 33, 0, 393, 394, 3, 74, 37,
		0, 394, 395, 3, 64, 32, 0, 395, 397, 1, 0, 0, 0, 396, 392, 1, 0, 0, 0,
		396, 393, 1, 0, 0, 0, 397, 65, 1, 0, 0, 0, 398, 441, 3, 82, 41, 0, 399,
		400, 3, 80, 40, 0, 400, 401, 3, 82, 41, 0, 401, 442, 1, 0, 0, 0, 402, 404,
		3, 74, 37, 0, 403, 402, 1, 0, 0, 0, 403, 404, 1, 0, 0, 0, 404, 405, 1,
		0, 0, 0, 405, 406, 3, 76, 38, 0, 406, 407, 3, 82, 41, 0, 407, 408, 3, 78,
		39, 0, 408, 409, 3, 82, 41, 0, 409, 442, 1, 0, 0, 0, 410, 412, 3, 74, 37,
		0, 411, 410, 1, 0, 0, 0, 411, 412, 1, 0, 0, 0, 412, 413, 1, 0, 0, 0, 413,
		414, 5, 23, 0, 0, 414, 415, 5, 49, 0, 0, 415, 420, 3, 56, 28, 0, 416, 417,
		5, 46, 0, 0, 417, 419, 3, 56, 28, 0, 418, 416, 1, 0, 0, 0, 419, 422, 1,
		0, 0, 0, 420, 418, 1, 0, 0, 0, 420, 421, 1, 0, 0, 0, 421, 423, 1, 0, 0,
		0, 422, 420, 1, 0, 0, 0, 423, 424, 5, 50, 0, 0, 424, 442, 1, 0, 0, 0, 425,
		427, 3, 74, 37, 0, 426, 425, 1, 0, 0, 0, 426, 427, 1, 0, 0, 0, 427, 428,
		1, 0, 0, 0, 428, 429, 5, 24, 0, 0, 429, 442, 3, 82, 41, 0, 430, 439, 5,
		25, 0, 0, 431, 433, 3, 74, 37, 0, 432, 431, 1, 0, 0, 0, 432, 433, 1, 0,
		0, 0, 433, 434, 1, 0, 0, 0, 434, 440, 7, 2, 0, 0, 435, 437, 3, 74, 37,
		0, 436, 435, 1, 0, 0, 0, 436, 437, 1, 0, 0, 0, 437, 438, 1, 0, 0, 0, 438,
		440, 5, 29, 0, 0, 439, 432, 1, 0, 0, 0, 439, 436, 1, 0, 0, 0, 440, 442,
		1, 0, 0, 0, 441, 399, 1, 0, 0, 0, 441, 403, 1, 0, 0, 0, 441, 411, 1, 0,
		0, 0, 441, 426, 1, 0, 0, 0, 441, 430, 1, 0, 0, 0, 441, 442, 1, 0, 0, 0,
		442, 67, 1, 0, 0, 0, 443, 447, 5, 18, 0, 0, 444, 445, 4, 34, 1, 0, 445,
		447, 5, 62, 0, 0, 446, 443, 1, 0, 0, 0, 446, 444, 1, 0, 0, 0, 447, 69,
		1, 0, 0, 0, 448, 452, 5, 19, 0, 0, 449, 450, 4, 35, 2, 0, 450, 452, 5,
		62, 0, 0, 451, 448, 1, 0, 0, 0, 451, 449, 1, 0, 0, 0, 452, 71, 1, 0, 0,
		0, 453, 457, 5, 20, 0, 0, 454, 455, 4, 36, 3, 0, 455, 457, 5, 62, 0, 0,
		456, 453, 1, 0, 0, 0, 456, 454, 1, 0, 0, 0, 457, 73, 1, 0, 0, 0, 458, 462,
		5, 21, 0, 0, 459, 460, 4, 37, 4, 0, 460, 462, 5, 62, 0, 0, 461, 458, 1,
		0, 0, 0, 461, 459, 1, 0, 0, 0, 462, 75, 1, 0, 0, 0, 463, 467, 5, 22, 0,
		0, 464, 465, 4, 38, 5, 0, 465, 467, 5, 62, 0, 0, 466, 463, 1, 0, 0, 0,
		466, 464, 1, 0, 0, 0, 467, 77, 1, 0, 0, 0, 468, 472, 5, 18, 0, 0, 469,
		470, 4, 39, 6, 0, 470, 472, 5, 62, 0, 0, 471, 468, 1, 0, 0, 0, 471, 469,
		1, 0, 0, 0, 472, 79, 1, 0, 0, 0, 473, 474, 7, 3, 0, 0, 474, 81, 1, 0, 0,
		0, 475, 480, 3, 84, 42, 0, 476, 477, 7, 1, 0, 0, 477, 479, 3, 84, 42, 0,
		478, 476, 1, 0, 0, 0, 479, 482, 1, 0, 0, 0, 480, 478, 1, 0, 0, 0, 480,
		481, 1, 0, 0, 0, 481, 83, 1, 0, 0, 0, 482, 480, 1, 0, 0, 0, 483, 488, 3,
		86, 43, 0, 484, 485, 7, 4, 0, 0, 485, 487, 3, 86, 43, 0, 486, 484, 1, 0,
		0, 0, 487, 490, 1, 0, 0, 0, 488, 486, 1, 0, 0, 0, 488, 489, 1, 0, 0, 0,
		489, 85, 1, 0, 0, 0, 490, 488, 1, 0, 0, 0, 491, 492, 7, 1, 0, 0, 492, 495,
		3, 86, 43, 0, 493, 495, 3, 88, 44, 0, 494, 491, 1, 0, 0, 0, 494, 493, 1,
		0, 0, 0, 495, 87, 1, 0, 0, 0, 496, 500, 3, 92, 46, 0, 497, 499, 3, 90,
		45, 0, 498, 497, 1, 0, 0, 0, 499, 502, 1, 0, 0, 0, 500, 498, 1, 0, 0, 0,
		500, 501, 1, 0, 0, 0, 501, 89, 1, 0, 0, 0, 502, 500, 1, 0, 0, 0, 503, 504,
		5, 48, 0, 0, 504, 510, 3, 110, 55, 0, 505, 506, 5, 51, 0, 0, 506, 507,
		3, 56, 28, 0, 507, 508, 5, 52, 0, 0, 508, 510, 1, 0, 0, 0, 509, 503, 1,
		0, 0, 0, 509, 505, 1, 0, 0, 0, 510, 91, 1, 0, 0, 0, 511, 523, 3, 94, 47,
		0, 512, 523, 3, 108, 54, 0, 513, 523, 5, 61, 0, 0, 514, 523, 3, 100, 50,
		0, 515, 523, 3, 112, 56, 0, 516, 523, 3, 114, 57, 0, 517, 518, 5, 49, 0,
		0, 518, 519, 3, 56, 28, 0, 519, 520, 5, 50, 0, 0, 520, 523, 1, 0, 0, 0,
		521, 523, 3, 120, 60, 0, 522, 511, 1, 0, 0, 0, 522, 512, 1, 0, 0, 0, 522,
		513, 1, 0, 0, 0, 522, 514, 1, 0, 0, 0, 522, 515, 1, 0, 0, 0, 522, 516,
		1, 0, 0, 0, 522, 517, 1, 0, 0, 0, 522, 521, 1, 0, 0, 0, 523, 93, 1, 0,
		0, 0, 524, 525, 3, 110, 55, 0, 525, 527, 5, 49, 0, 0, 526, 528, 3, 96,
		48, 0, 527, 526, 1, 0, 0, 0, 527, 528, 1, 0, 0, 0, 528, 529, 1, 0, 0, 0,
		529, 530, 5, 50, 0, 0, 530, 95, 1, 0, 0, 0, 531, 536, 3, 98, 49, 0, 532,
		533, 5, 46, 0, 0, 533, 535, 3, 98, 49, 0, 534, 532, 1, 0, 0, 0, 535, 538,
		1, 0, 0, 0, 536, 534, 1, 0, 0, 0, 536, 537, 1, 0, 0, 0, 537, 555, 1, 0,
		0, 0, 538, 536, 1, 0, 0, 0, 539, 544, 3, 56, 28, 0, 540, 541, 5, 46, 0,
		0, 541, 543, 3, 56, 28, 0, 542, 540, 1, 0, 0, 0, 543, 546, 1, 0, 0, 0,
		544, 542, 1, 0, 0, 0, 544, 545, 1, 0, 0, 0, 545, 551, 1, 0, 0, 0, 546,
		544, 1, 0, 0, 0, 547, 548, 5, 46, 0, 0, 548, 550, 3, 98, 49, 0, 549, 547,
		1, 0, 0, 0, 550, 553, 1, 0, 0, 0, 551, 549, 1, 0, 0, 0, 551, 552, 1, 0,
		0, 0, 552, 555, 1, 0, 0, 0, 553, 551, 1, 0, 0, 0, 554, 531, 1, 0, 0, 0,
		554, 539, 1, 0, 0, 0, 555, 97, 1, 0, 0, 0, 556, 557, 3, 110, 55, 0, 557,
		558, 5, 47, 0, 0, 558, 559, 3, 56, 28, 0, 559, 99, 1, 0, 0, 0, 560, 566,
		5, 60, 0, 0, 561, 566, 5, 28, 0, 0, 562, 566, 5, 26, 0, 0, 563, 566, 5,
		56, 0, 0, 564, 566, 3, 102, 51, 0, 565, 560, 1, 0, 0, 0, 565, 561, 1, 0,
		0, 0, 565, 562, 1, 0, 0, 0, 565, 563, 1, 0, 0, 0, 565, 564, 1, 0, 0, 0,
		566, 101, 1, 0, 0, 0, 567, 576, 5, 57, 0, 0, 568, 575, 5, 67, 0, 0, 569,
		575, 5, 68, 0, 0, 570, 571, 5, 66, 0, 0, 571, 572, 3, 56, 28, 0, 572, 573,
		5, 54, 0, 0, 573, 575, 1, 0, 0, 0, 574, 568, 1, 0, 0, 0, 574, 569, 1, 0,
		0, 0, 574, 570, 1, 0, 0, 0, 575, 578, 1, 0, 0, 0, 576, 574, 1, 0, 0, 0,
		576, 577, 1, 0, 0, 0, 577, 579, 1, 0, 0, 0, 578, 576, 1, 0, 0, 0, 579,
		580, 5, 65, 0, 0, 580, 103, 1, 0, 0, 0, 581, 583, 5, 58, 0, 0, 582, 584,
		7, 5, 0, 0, 583, 582, 1, 0, 0, 0, 584, 585, 1, 0, 0, 0, 585, 583, 1, 0,
		0, 0, 585, 586, 1, 0, 0, 0, 586, 587, 1, 0, 0, 0, 587, 588, 5, 69, 0, 0,
		588, 105, 1, 0, 0, 0, 589, 593, 5, 58, 0, 0, 590, 592, 7, 5, 0, 0, 591,
		590, 1, 0, 0, 0, 592, 595, 1, 0, 0, 0, 593, 591, 1, 0, 0, 0, 593, 594,
		1, 0, 0, 0, 594, 596, 1, 0, 0, 0, 595, 593, 1, 0, 0, 0, 596, 597, 5, 70,
		0, 0, 597, 598, 3, 56, 28, 0, 598, 607, 5, 54, 0, 0, 599, 606, 5, 71, 0,
		0, 600, 606, 5, 72, 0, 0, 601, 602, 5, 70, 0, 0, 602, 603, 3, 56, 28, 0,
		603, 604, 5, 54, 0, 0, 604, 606, 1, 0, 0, 0, 605, 599, 1, 0, 0, 0, 605,
		600, 1, 0, 0, 0, 605, 601, 1, 0, 0, 0, 606, 609, 1, 0, 0, 0, 607, 605,
		1, 0, 0, 0, 607, 608, 1, 0, 0, 0, 608, 610, 1, 0, 0, 0, 609, 607, 1, 0,
		0, 0, 610, 611, 5, 69, 0, 0, 611, 107, 1, 0, 0, 0, 612, 615, 3, 110, 55,
		0, 613, 615, 3, 106, 53, 0, 614, 612, 1, 0, 0, 0, 614, 613, 1, 0, 0, 0,
		615, 109, 1, 0, 0, 0, 616, 620, 5, 62, 0, 0, 617, 620, 5, 4, 0, 0, 618,
		620, 3, 104, 52, 0, 619, 616, 1, 0, 0, 0, 619, 617, 1, 0, 0, 0, 619, 618,
		1, 0, 0, 0, 620, 111, 1, 0, 0, 0, 621, 633, 5, 51, 0, 0, 622, 627, 3, 56,
		28, 0, 623, 624, 5, 46, 0, 0, 624, 626, 3, 56, 28, 0, 625, 623, 1, 0, 0,
		0, 626, 629, 1, 0, 0, 0, 627, 625, 1, 0, 0, 0, 627, 628, 1, 0, 0, 0, 628,
		631, 1, 0, 0, 0, 629, 627, 1, 0, 0, 0, 630, 632, 5, 46, 0, 0, 631, 630,
		1, 0, 0, 0, 631, 632, 1, 0, 0, 0, 632, 634, 1, 0, 0, 0, 633, 622, 1, 0,
		0, 0, 633, 634, 1, 0, 0, 0, 634, 635, 1, 0, 0, 0, 635, 636, 5, 52, 0, 0,
		636, 113, 1, 0, 0, 0, 637, 649, 5, 53, 0, 0, 638, 643, 3, 116, 58, 0, 639,
		640, 5, 46, 0, 0, 640, 642, 3, 116, 58, 0, 641, 639, 1, 0, 0, 0, 642, 645,
		1, 0, 0, 0, 643, 641, 1, 0, 0, 0, 643, 644, 1, 0, 0, 0, 644, 647, 1, 0,
		0, 0, 645, 643, 1, 0, 0, 0, 646, 648, 5, 46, 0, 0, 647, 646, 1, 0, 0, 0,
		647, 648, 1, 0, 0, 0, 648, 650, 1, 0, 0, 0, 649, 638, 1, 0, 0, 0, 649,
		650, 1, 0, 0, 0, 650, 651, 1, 0, 0, 0, 651, 652, 5, 54, 0, 0, 652, 115,
		1, 0, 0, 0, 653, 654, 3, 118, 59, 0, 654, 655, 5, 47, 0, 0, 655, 656, 3,
		56, 28, 0, 656, 117, 1, 0, 0, 0, 657, 660, 3, 110, 55, 0, 658, 660, 3,
		102, 51, 0, 659, 657, 1, 0, 0, 0, 659, 658, 1, 0, 0, 0, 660, 119, 1, 0,
		0, 0, 661, 662, 3, 32, 16, 0, 662, 121, 1, 0, 0, 0, 663, 677, 3, 124, 62,
		0, 664, 673, 5, 49, 0, 0, 665, 670, 3, 124, 62, 0, 666, 667, 5, 46, 0,
		0, 667, 669, 3, 124, 62, 0, 668, 666, 1, 0, 0, 0, 669, 672, 1, 0, 0, 0,
		670, 668, 1, 0, 0, 0, 670, 671, 1, 0, 0, 0, 671, 674, 1, 0, 0, 0, 672,
		670, 1, 0, 0, 0, 673, 665, 1, 0, 0, 0, 673, 674, 1, 0, 0, 0, 674, 675,
		1, 0, 0, 0, 675, 677, 5, 50, 0, 0, 676, 663, 1, 0, 0, 0, 676, 664, 1, 0,
		0, 0, 677, 678, 1, 0, 0, 0, 678, 681, 5, 30, 0, 0, 679, 682, 3, 56, 28,
		0, 680, 682, 3, 126, 63, 0, 681, 679, 1, 0, 0, 0, 681, 680, 1, 0, 0, 0,
		682, 123, 1, 0, 0, 0, 683, 686, 5, 61, 0, 0, 684, 685, 5, 47, 0, 0, 685,
		687, 5, 29, 0, 0, 686, 684, 1, 0, 0, 0, 686, 687, 1, 0, 0, 0, 687, 696,
		1, 0, 0, 0, 688, 694, 5, 35, 0, 0, 689, 691, 7, 1, 0, 0, 690, 689, 1, 0,
		0, 0, 690, 691, 1, 0, 0, 0, 691, 692, 1, 0, 0, 0, 692, 695, 5, 60, 0, 0,
		693, 695, 3, 102, 51, 0, 694, 690, 1, 0, 0, 0, 694, 693, 1, 0, 0, 0, 695,
		697, 1, 0, 0, 0, 696, 688, 1, 0, 0, 0, 696, 697, 1, 0, 0, 0, 697, 125,
		1, 0, 0, 0, 698, 702, 5, 53, 0, 0, 699, 701, 5, 63, 0, 0, 700, 699, 1,
		0, 0, 0, 701, 704, 1, 0, 0, 0, 702, 700, 1, 0, 0, 0, 702, 703, 1, 0, 0,
		0, 703, 724, 1, 0, 0, 0, 704, 702, 1, 0, 0, 0, 705, 706, 5, 61, 0, 0, 706,
		707, 5, 35, 0, 0, 707, 720, 3, 56, 28, 0, 708, 712, 5, 55, 0, 0, 709, 711,
		5, 63, 0, 0, 710, 709, 1, 0, 0, 0, 711, 714, 1, 0, 0, 0, 712, 710, 1, 0,
		0, 0, 712, 713, 1, 0, 0, 0, 713, 721, 1, 0, 0, 0, 714, 712, 1, 0, 0, 0,
		715, 717, 5, 63, 0, 0, 716, 715, 1, 0, 0, 0, 717, 718, 1, 0, 0, 0, 718,
		716, 1, 0, 0, 0, 718, 719, 1, 0, 0, 0, 719, 721, 1, 0, 0, 0, 720, 708,
		1, 0, 0, 0, 720, 716, 1, 0, 0, 0, 721, 723, 1, 0, 0, 0, 722, 705, 1, 0,
		0, 0, 723, 726, 1, 0, 0, 0, 724, 722, 1, 0, 0, 0, 724, 725, 1, 0, 0, 0,
		725, 727, 1, 0, 0, 0, 726, 724, 1, 0, 0, 0, 727, 728, 5, 17, 0, 0, 728,
		730, 3, 56, 28, 0, 729, 731, 5, 55, 0, 0, 730, 729, 1, 0, 0, 0, 730, 731,
		1, 0, 0, 0, 731, 735, 1, 0, 0, 0, 732, 734, 5, 63, 0, 0, 733, 732, 1, 0,
		0, 0, 734, 737, 1, 0, 0, 0, 735, 733, 1, 0, 0, 0, 735, 736, 1, 0, 0, 0,
		736, 738, 1, 0, 0, 0, 737, 735, 1, 0, 0, 0, 738, 739, 5, 54, 0, 0, 739,
		127, 1, 0, 0, 0, 93, 131, 136, 139, 144, 153, 160, 166, 175, 185, 196,
		205, 209, 213, 221, 233, 240, 247, 253, 259, 262, 268, 276, 284, 289, 305,
		310, 315, 323, 329, 347, 352, 359, 363, 371, 380, 389, 396, 403, 411, 420,
		426, 432, 436, 439, 441, 446, 451, 456, 461, 466, 471, 480, 488, 494, 500,
		509, 522, 527, 536, 544, 551, 554, 565, 574, 576, 585, 593, 605, 607, 614,
		619, 627, 631, 633, 643, 647, 649, 659, 670, 673, 676, 681, 686, 690, 694,
		696, 702, 712, 718, 720, 724, 730, 735,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// SPL2ParserInit initializes any static state used to implement SPL2Parser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewSPL2Parser(). You can call this function if you wish to initialize the static state ahead
// of time.
func SPL2ParserInit() {
	staticData := &SPL2ParserParserStaticData
	staticData.once.Do(spl2parserParserInit)
}

// NewSPL2Parser produces a new parser instance for the optional input antlr.TokenStream.
func NewSPL2Parser(input antlr.TokenStream) *SPL2Parser {
	SPL2ParserInit()
	this := new(SPL2Parser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &SPL2ParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "SPL2Parser.g4"

	return this
}

// Note that '@members' cannot be changed now, but this should have been 'globals'
// If you are looking to have variables for each instance, use '@structmembers'

// Casing is recognized only in an operator grammar position. The same lower-
// case token remains an ordinary identifier or search literal elsewhere.
func (p *SPL2Parser) contextualKeyword(word string) bool {
	return strings.EqualFold(p.GetTokenStream().LT(1).GetText(), word)
}

// SPL2Parser tokens.
const (
	SPL2ParserEOF                  = antlr.TokenEOF
	SPL2ParserFROM                 = 1
	SPL2ParserSELECT               = 2
	SPL2ParserSEARCH               = 3
	SPL2ParserINDEX                = 4
	SPL2ParserEVAL                 = 5
	SPL2ParserWHERE                = 6
	SPL2ParserFIELDS               = 7
	SPL2ParserTABLE                = 8
	SPL2ParserAS                   = 9
	SPL2ParserGROUP                = 10
	SPL2ParserREX                  = 11
	SPL2ParserMAKERESULTS          = 12
	SPL2ParserSPL1                 = 13
	SPL2ParserIMPORT               = 14
	SPL2ParserEXPORT               = 15
	SPL2ParserFUNCTION             = 16
	SPL2ParserRETURN               = 17
	SPL2ParserAND                  = 18
	SPL2ParserOR                   = 19
	SPL2ParserXOR                  = 20
	SPL2ParserNOT                  = 21
	SPL2ParserBETWEEN              = 22
	SPL2ParserIN                   = 23
	SPL2ParserLIKE                 = 24
	SPL2ParserIS                   = 25
	SPL2ParserNULL                 = 26
	SPL2ParserNULL_TEST            = 27
	SPL2ParserBOOLEAN              = 28
	SPL2ParserTYPE                 = 29
	SPL2ParserARROW                = 30
	SPL2ParserLE                   = 31
	SPL2ParserGE                   = 32
	SPL2ParserNE                   = 33
	SPL2ParserEQ                   = 34
	SPL2ParserASSIGN               = 35
	SPL2ParserLT                   = 36
	SPL2ParserGT                   = 37
	SPL2ParserPLUS                 = 38
	SPL2ParserMINUS                = 39
	SPL2ParserSTAR                 = 40
	SPL2ParserLINE_COMMENT         = 41
	SPL2ParserBLOCK_COMMENT        = 42
	SPL2ParserSLASH                = 43
	SPL2ParserMOD                  = 44
	SPL2ParserPIPE                 = 45
	SPL2ParserCOMMA                = 46
	SPL2ParserCOLON                = 47
	SPL2ParserDOT                  = 48
	SPL2ParserLPAREN               = 49
	SPL2ParserRPAREN               = 50
	SPL2ParserLBRACKET             = 51
	SPL2ParserRBRACKET             = 52
	SPL2ParserLBRACE               = 53
	SPL2ParserRBRACE               = 54
	SPL2ParserSEMI                 = 55
	SPL2ParserRAW_STRING           = 56
	SPL2ParserDQUOTE               = 57
	SPL2ParserSQUOTE               = 58
	SPL2ParserBACKTICK             = 59
	SPL2ParserNUMBER               = 60
	SPL2ParserLOCAL                = 61
	SPL2ParserIDENTIFIER           = 62
	SPL2ParserNL                   = 63
	SPL2ParserWS                   = 64
	SPL2ParserSTRING_END           = 65
	SPL2ParserSTRING_INTERPOLATION = 66
	SPL2ParserSTRING_TEXT          = 67
	SPL2ParserSTRING_DOLLAR        = 68
	SPL2ParserNAME_END             = 69
	SPL2ParserNAME_INTERPOLATION   = 70
	SPL2ParserNAME_TEXT            = 71
	SPL2ParserNAME_DOLLAR          = 72
	SPL2ParserEMBEDDED_END         = 73
	SPL2ParserEMBEDDED_TEXT        = 74
	SPL2ParserREGEX                = 75
	SPL2ParserRX_WS                = 76
)

// SPL2Parser rules.
const (
	SPL2ParserRULE_query              = 0
	SPL2ParserRULE_pipeline           = 1
	SPL2ParserRULE_start              = 2
	SPL2ParserRULE_command            = 3
	SPL2ParserRULE_fromCommand        = 4
	SPL2ParserRULE_selectCommand      = 5
	SPL2ParserRULE_projection         = 6
	SPL2ParserRULE_dataset            = 7
	SPL2ParserRULE_generator          = 8
	SPL2ParserRULE_evalCommand        = 9
	SPL2ParserRULE_assignment         = 10
	SPL2ParserRULE_whereCommand       = 11
	SPL2ParserRULE_fieldsCommand      = 12
	SPL2ParserRULE_rexCommand         = 13
	SPL2ParserRULE_rexOption          = 14
	SPL2ParserRULE_embeddedCommand    = 15
	SPL2ParserRULE_embeddedText       = 16
	SPL2ParserRULE_moduleSuffix       = 17
	SPL2ParserRULE_moduleDeclaration  = 18
	SPL2ParserRULE_searchCommand      = 19
	SPL2ParserRULE_implicitSearch     = 20
	SPL2ParserRULE_searchExpression   = 21
	SPL2ParserRULE_searchXor          = 22
	SPL2ParserRULE_searchAnd          = 23
	SPL2ParserRULE_searchOr           = 24
	SPL2ParserRULE_searchNot          = 25
	SPL2ParserRULE_searchAtom         = 26
	SPL2ParserRULE_searchValue        = 27
	SPL2ParserRULE_expression         = 28
	SPL2ParserRULE_xorExpression      = 29
	SPL2ParserRULE_orExpression       = 30
	SPL2ParserRULE_andExpression      = 31
	SPL2ParserRULE_notExpression      = 32
	SPL2ParserRULE_predicate          = 33
	SPL2ParserRULE_logicalAnd         = 34
	SPL2ParserRULE_logicalOr          = 35
	SPL2ParserRULE_logicalXor         = 36
	SPL2ParserRULE_logicalNot         = 37
	SPL2ParserRULE_betweenOperator    = 38
	SPL2ParserRULE_betweenConjunction = 39
	SPL2ParserRULE_comparison         = 40
	SPL2ParserRULE_additive           = 41
	SPL2ParserRULE_multiplicative     = 42
	SPL2ParserRULE_unary              = 43
	SPL2ParserRULE_access             = 44
	SPL2ParserRULE_accessPart         = 45
	SPL2ParserRULE_primary            = 46
	SPL2ParserRULE_call               = 47
	SPL2ParserRULE_arguments          = 48
	SPL2ParserRULE_namedArgument      = 49
	SPL2ParserRULE_literal            = 50
	SPL2ParserRULE_stringLiteral      = 51
	SPL2ParserRULE_quotedName         = 52
	SPL2ParserRULE_fieldTemplate      = 53
	SPL2ParserRULE_fieldName          = 54
	SPL2ParserRULE_identifier         = 55
	SPL2ParserRULE_array              = 56
	SPL2ParserRULE_object             = 57
	SPL2ParserRULE_objectEntry        = 58
	SPL2ParserRULE_objectKey          = 59
	SPL2ParserRULE_searchLiteral      = 60
	SPL2ParserRULE_lambdaExpression   = 61
	SPL2ParserRULE_lambdaParameter    = 62
	SPL2ParserRULE_lambdaBlock        = 63
)

// IQueryContext is an interface to support dynamic dispatch.
type IQueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	Pipeline() IPipelineContext
	ModuleDeclaration() IModuleDeclarationContext
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode
	ModuleSuffix() IModuleSuffixContext

	// IsQueryContext differentiates from other interfaces.
	IsQueryContext()
}

type QueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQueryContext() *QueryContext {
	var p = new(QueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_query
	return p
}

func InitEmptyQueryContext(p *QueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_query
}

func (*QueryContext) IsQueryContext() {}

func NewQueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QueryContext {
	var p = new(QueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_query

	return p
}

func (s *QueryContext) GetParser() antlr.Parser { return s.parser }

func (s *QueryContext) EOF() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEOF, 0)
}

func (s *QueryContext) Pipeline() IPipelineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPipelineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPipelineContext)
}

func (s *QueryContext) ModuleDeclaration() IModuleDeclarationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IModuleDeclarationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IModuleDeclarationContext)
}

func (s *QueryContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNL)
}

func (s *QueryContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNL, i)
}

func (s *QueryContext) ModuleSuffix() IModuleSuffixContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IModuleSuffixContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IModuleSuffixContext)
}

func (s *QueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QueryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterQuery(s)
	}
}

func (s *QueryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitQuery(s)
	}
}

func (s *QueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitQuery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Query() (localctx IQueryContext) {
	localctx = NewQueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, SPL2ParserRULE_query)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(131)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(128)
				p.Match(SPL2ParserNL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(133)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(134)
			p.Pipeline()
		}
		p.SetState(136)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserSEMI {
			{
				p.SetState(135)
				p.ModuleSuffix()
			}

		}

	case 2:
		{
			p.SetState(138)
			p.ModuleDeclaration()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.SetState(144)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(141)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(146)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(147)
		p.Match(SPL2ParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPipelineContext is an interface to support dynamic dispatch.
type IPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Start_() IStartContext
	AllPIPE() []antlr.TerminalNode
	PIPE(i int) antlr.TerminalNode
	AllCommand() []ICommandContext
	Command(i int) ICommandContext
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode

	// IsPipelineContext differentiates from other interfaces.
	IsPipelineContext()
}

type PipelineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelineContext() *PipelineContext {
	var p = new(PipelineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_pipeline
	return p
}

func InitEmptyPipelineContext(p *PipelineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_pipeline
}

func (*PipelineContext) IsPipelineContext() {}

func NewPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelineContext {
	var p = new(PipelineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_pipeline

	return p
}

func (s *PipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelineContext) Start_() IStartContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStartContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStartContext)
}

func (s *PipelineContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserPIPE)
}

func (s *PipelineContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserPIPE, i)
}

func (s *PipelineContext) AllCommand() []ICommandContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICommandContext); ok {
			len++
		}
	}

	tst := make([]ICommandContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICommandContext); ok {
			tst[i] = t.(ICommandContext)
			i++
		}
	}

	return tst
}

func (s *PipelineContext) Command(i int) ICommandContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommandContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommandContext)
}

func (s *PipelineContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNL)
}

func (s *PipelineContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNL, i)
}

func (s *PipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterPipeline(s)
	}
}

func (s *PipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitPipeline(s)
	}
}

func (s *PipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Pipeline() (localctx IPipelineContext) {
	localctx = NewPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, SPL2ParserRULE_pipeline)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(149)
		p.Start_()
	}
	p.SetState(166)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			p.SetState(153)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(150)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(155)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(156)
				p.Match(SPL2ParserPIPE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(160)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(157)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(162)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(163)
				p.Command()
			}

		}
		p.SetState(168)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStartContext is an interface to support dynamic dispatch.
type IStartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FromCommand() IFromCommandContext
	SelectCommand() ISelectCommandContext
	SearchCommand() ISearchCommandContext
	ImplicitSearch() IImplicitSearchContext
	Generator() IGeneratorContext
	EmbeddedCommand() IEmbeddedCommandContext

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_start
	return p
}

func InitEmptyStartContext(p *StartContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_start
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) FromCommand() IFromCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFromCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFromCommandContext)
}

func (s *StartContext) SelectCommand() ISelectCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectCommandContext)
}

func (s *StartContext) SearchCommand() ISearchCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchCommandContext)
}

func (s *StartContext) ImplicitSearch() IImplicitSearchContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IImplicitSearchContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IImplicitSearchContext)
}

func (s *StartContext) Generator() IGeneratorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGeneratorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGeneratorContext)
}

func (s *StartContext) EmbeddedCommand() IEmbeddedCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEmbeddedCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEmbeddedCommandContext)
}

func (s *StartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterStart(s)
	}
}

func (s *StartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitStart(s)
	}
}

func (s *StartContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitStart(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Start_() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, SPL2ParserRULE_start)
	p.SetState(175)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(169)
			p.FromCommand()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(170)
			p.SelectCommand()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(171)
			p.SearchCommand()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(172)
			p.ImplicitSearch()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(173)
			p.Generator()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(174)
			p.EmbeddedCommand()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICommandContext is an interface to support dynamic dispatch.
type ICommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EvalCommand() IEvalCommandContext
	WhereCommand() IWhereCommandContext
	FieldsCommand() IFieldsCommandContext
	FromCommand() IFromCommandContext
	SelectCommand() ISelectCommandContext
	SearchCommand() ISearchCommandContext
	RexCommand() IRexCommandContext
	EmbeddedCommand() IEmbeddedCommandContext

	// IsCommandContext differentiates from other interfaces.
	IsCommandContext()
}

type CommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCommandContext() *CommandContext {
	var p = new(CommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_command
	return p
}

func InitEmptyCommandContext(p *CommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_command
}

func (*CommandContext) IsCommandContext() {}

func NewCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CommandContext {
	var p = new(CommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_command

	return p
}

func (s *CommandContext) GetParser() antlr.Parser { return s.parser }

func (s *CommandContext) EvalCommand() IEvalCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEvalCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEvalCommandContext)
}

func (s *CommandContext) WhereCommand() IWhereCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhereCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhereCommandContext)
}

func (s *CommandContext) FieldsCommand() IFieldsCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldsCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldsCommandContext)
}

func (s *CommandContext) FromCommand() IFromCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFromCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFromCommandContext)
}

func (s *CommandContext) SelectCommand() ISelectCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectCommandContext)
}

func (s *CommandContext) SearchCommand() ISearchCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchCommandContext)
}

func (s *CommandContext) RexCommand() IRexCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRexCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRexCommandContext)
}

func (s *CommandContext) EmbeddedCommand() IEmbeddedCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEmbeddedCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEmbeddedCommandContext)
}

func (s *CommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterCommand(s)
	}
}

func (s *CommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitCommand(s)
	}
}

func (s *CommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Command() (localctx ICommandContext) {
	localctx = NewCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, SPL2ParserRULE_command)
	p.SetState(185)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserEVAL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(177)
			p.EvalCommand()
		}

	case SPL2ParserWHERE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(178)
			p.WhereCommand()
		}

	case SPL2ParserFIELDS, SPL2ParserTABLE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(179)
			p.FieldsCommand()
		}

	case SPL2ParserFROM:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(180)
			p.FromCommand()
		}

	case SPL2ParserSELECT:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(181)
			p.SelectCommand()
		}

	case SPL2ParserSEARCH:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(182)
			p.SearchCommand()
		}

	case SPL2ParserREX:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(183)
			p.RexCommand()
		}

	case SPL2ParserSPL1, SPL2ParserBACKTICK:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(184)
			p.EmbeddedCommand()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFromCommandContext is an interface to support dynamic dispatch.
type IFromCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FROM() antlr.TerminalNode
	Dataset() IDatasetContext

	// IsFromCommandContext differentiates from other interfaces.
	IsFromCommandContext()
}

type FromCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFromCommandContext() *FromCommandContext {
	var p = new(FromCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fromCommand
	return p
}

func InitEmptyFromCommandContext(p *FromCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fromCommand
}

func (*FromCommandContext) IsFromCommandContext() {}

func NewFromCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FromCommandContext {
	var p = new(FromCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_fromCommand

	return p
}

func (s *FromCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *FromCommandContext) FROM() antlr.TerminalNode {
	return s.GetToken(SPL2ParserFROM, 0)
}

func (s *FromCommandContext) Dataset() IDatasetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDatasetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDatasetContext)
}

func (s *FromCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FromCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FromCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterFromCommand(s)
	}
}

func (s *FromCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitFromCommand(s)
	}
}

func (s *FromCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitFromCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) FromCommand() (localctx IFromCommandContext) {
	localctx = NewFromCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, SPL2ParserRULE_fromCommand)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(187)
		p.Match(SPL2ParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(188)
		p.Dataset()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectCommandContext is an interface to support dynamic dispatch.
type ISelectCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SELECT() antlr.TerminalNode
	AllProjection() []IProjectionContext
	Projection(i int) IProjectionContext
	FROM() antlr.TerminalNode
	Dataset() IDatasetContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsSelectCommandContext differentiates from other interfaces.
	IsSelectCommandContext()
}

type SelectCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectCommandContext() *SelectCommandContext {
	var p = new(SelectCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_selectCommand
	return p
}

func InitEmptySelectCommandContext(p *SelectCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_selectCommand
}

func (*SelectCommandContext) IsSelectCommandContext() {}

func NewSelectCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectCommandContext {
	var p = new(SelectCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_selectCommand

	return p
}

func (s *SelectCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectCommandContext) SELECT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSELECT, 0)
}

func (s *SelectCommandContext) AllProjection() []IProjectionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IProjectionContext); ok {
			len++
		}
	}

	tst := make([]IProjectionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IProjectionContext); ok {
			tst[i] = t.(IProjectionContext)
			i++
		}
	}

	return tst
}

func (s *SelectCommandContext) Projection(i int) IProjectionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IProjectionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IProjectionContext)
}

func (s *SelectCommandContext) FROM() antlr.TerminalNode {
	return s.GetToken(SPL2ParserFROM, 0)
}

func (s *SelectCommandContext) Dataset() IDatasetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDatasetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDatasetContext)
}

func (s *SelectCommandContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *SelectCommandContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *SelectCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSelectCommand(s)
	}
}

func (s *SelectCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSelectCommand(s)
	}
}

func (s *SelectCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSelectCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SelectCommand() (localctx ISelectCommandContext) {
	localctx = NewSelectCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, SPL2ParserRULE_selectCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(190)
		p.Match(SPL2ParserSELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(191)
		p.Projection()
	}
	p.SetState(196)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(192)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(193)
			p.Projection()
		}

		p.SetState(198)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(199)
		p.Match(SPL2ParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(200)
		p.Dataset()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IProjectionContext is an interface to support dynamic dispatch.
type IProjectionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	AS() antlr.TerminalNode
	Identifier() IIdentifierContext

	// IsProjectionContext differentiates from other interfaces.
	IsProjectionContext()
}

type ProjectionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProjectionContext() *ProjectionContext {
	var p = new(ProjectionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_projection
	return p
}

func InitEmptyProjectionContext(p *ProjectionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_projection
}

func (*ProjectionContext) IsProjectionContext() {}

func NewProjectionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProjectionContext {
	var p = new(ProjectionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_projection

	return p
}

func (s *ProjectionContext) GetParser() antlr.Parser { return s.parser }

func (s *ProjectionContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ProjectionContext) AS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserAS, 0)
}

func (s *ProjectionContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ProjectionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProjectionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProjectionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterProjection(s)
	}
}

func (s *ProjectionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitProjection(s)
	}
}

func (s *ProjectionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitProjection(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Projection() (localctx IProjectionContext) {
	localctx = NewProjectionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, SPL2ParserRULE_projection)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(202)
		p.Expression()
	}
	p.SetState(205)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserAS {
		{
			p.SetState(203)
			p.Match(SPL2ParserAS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(204)
			p.Identifier()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDatasetContext is an interface to support dynamic dispatch.
type IDatasetContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	Array() IArrayContext

	// IsDatasetContext differentiates from other interfaces.
	IsDatasetContext()
}

type DatasetContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDatasetContext() *DatasetContext {
	var p = new(DatasetContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_dataset
	return p
}

func InitEmptyDatasetContext(p *DatasetContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_dataset
}

func (*DatasetContext) IsDatasetContext() {}

func NewDatasetContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DatasetContext {
	var p = new(DatasetContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_dataset

	return p
}

func (s *DatasetContext) GetParser() antlr.Parser { return s.parser }

func (s *DatasetContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *DatasetContext) Array() IArrayContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayContext)
}

func (s *DatasetContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DatasetContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DatasetContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterDataset(s)
	}
}

func (s *DatasetContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitDataset(s)
	}
}

func (s *DatasetContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitDataset(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Dataset() (localctx IDatasetContext) {
	localctx = NewDatasetContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, SPL2ParserRULE_dataset)
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(207)
			p.Identifier()
		}

	case SPL2ParserLBRACKET:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(208)
			p.Array()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGeneratorContext is an interface to support dynamic dispatch.
type IGeneratorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MAKERESULTS() antlr.TerminalNode
	NUMBER() antlr.TerminalNode

	// IsGeneratorContext differentiates from other interfaces.
	IsGeneratorContext()
}

type GeneratorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGeneratorContext() *GeneratorContext {
	var p = new(GeneratorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_generator
	return p
}

func InitEmptyGeneratorContext(p *GeneratorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_generator
}

func (*GeneratorContext) IsGeneratorContext() {}

func NewGeneratorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GeneratorContext {
	var p = new(GeneratorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_generator

	return p
}

func (s *GeneratorContext) GetParser() antlr.Parser { return s.parser }

func (s *GeneratorContext) MAKERESULTS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserMAKERESULTS, 0)
}

func (s *GeneratorContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNUMBER, 0)
}

func (s *GeneratorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GeneratorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GeneratorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterGenerator(s)
	}
}

func (s *GeneratorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitGenerator(s)
	}
}

func (s *GeneratorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitGenerator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Generator() (localctx IGeneratorContext) {
	localctx = NewGeneratorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, SPL2ParserRULE_generator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(211)
		p.Match(SPL2ParserMAKERESULTS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(213)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserNUMBER {
		{
			p.SetState(212)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEvalCommandContext is an interface to support dynamic dispatch.
type IEvalCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EVAL() antlr.TerminalNode
	AllAssignment() []IAssignmentContext
	Assignment(i int) IAssignmentContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsEvalCommandContext differentiates from other interfaces.
	IsEvalCommandContext()
}

type EvalCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEvalCommandContext() *EvalCommandContext {
	var p = new(EvalCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_evalCommand
	return p
}

func InitEmptyEvalCommandContext(p *EvalCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_evalCommand
}

func (*EvalCommandContext) IsEvalCommandContext() {}

func NewEvalCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EvalCommandContext {
	var p = new(EvalCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_evalCommand

	return p
}

func (s *EvalCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *EvalCommandContext) EVAL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEVAL, 0)
}

func (s *EvalCommandContext) AllAssignment() []IAssignmentContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAssignmentContext); ok {
			len++
		}
	}

	tst := make([]IAssignmentContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAssignmentContext); ok {
			tst[i] = t.(IAssignmentContext)
			i++
		}
	}

	return tst
}

func (s *EvalCommandContext) Assignment(i int) IAssignmentContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAssignmentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAssignmentContext)
}

func (s *EvalCommandContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *EvalCommandContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *EvalCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EvalCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EvalCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterEvalCommand(s)
	}
}

func (s *EvalCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitEvalCommand(s)
	}
}

func (s *EvalCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitEvalCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) EvalCommand() (localctx IEvalCommandContext) {
	localctx = NewEvalCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, SPL2ParserRULE_evalCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(215)
		p.Match(SPL2ParserEVAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(216)
		p.Assignment()
	}
	p.SetState(221)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(217)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(218)
			p.Assignment()
		}

		p.SetState(223)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAssignmentContext is an interface to support dynamic dispatch.
type IAssignmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FieldName() IFieldNameContext
	ASSIGN() antlr.TerminalNode
	Expression() IExpressionContext

	// IsAssignmentContext differentiates from other interfaces.
	IsAssignmentContext()
}

type AssignmentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAssignmentContext() *AssignmentContext {
	var p = new(AssignmentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_assignment
	return p
}

func InitEmptyAssignmentContext(p *AssignmentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_assignment
}

func (*AssignmentContext) IsAssignmentContext() {}

func NewAssignmentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AssignmentContext {
	var p = new(AssignmentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_assignment

	return p
}

func (s *AssignmentContext) GetParser() antlr.Parser { return s.parser }

func (s *AssignmentContext) FieldName() IFieldNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldNameContext)
}

func (s *AssignmentContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *AssignmentContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AssignmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AssignmentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AssignmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterAssignment(s)
	}
}

func (s *AssignmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitAssignment(s)
	}
}

func (s *AssignmentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitAssignment(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Assignment() (localctx IAssignmentContext) {
	localctx = NewAssignmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, SPL2ParserRULE_assignment)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(224)
		p.FieldName()
	}
	{
		p.SetState(225)
		p.Match(SPL2ParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(226)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWhereCommandContext is an interface to support dynamic dispatch.
type IWhereCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WHERE() antlr.TerminalNode
	Expression() IExpressionContext

	// IsWhereCommandContext differentiates from other interfaces.
	IsWhereCommandContext()
}

type WhereCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhereCommandContext() *WhereCommandContext {
	var p = new(WhereCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_whereCommand
	return p
}

func InitEmptyWhereCommandContext(p *WhereCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_whereCommand
}

func (*WhereCommandContext) IsWhereCommandContext() {}

func NewWhereCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WhereCommandContext {
	var p = new(WhereCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_whereCommand

	return p
}

func (s *WhereCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *WhereCommandContext) WHERE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserWHERE, 0)
}

func (s *WhereCommandContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *WhereCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WhereCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WhereCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterWhereCommand(s)
	}
}

func (s *WhereCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitWhereCommand(s)
	}
}

func (s *WhereCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitWhereCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) WhereCommand() (localctx IWhereCommandContext) {
	localctx = NewWhereCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, SPL2ParserRULE_whereCommand)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(228)
		p.Match(SPL2ParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(229)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldsCommandContext is an interface to support dynamic dispatch.
type IFieldsCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext
	FIELDS() antlr.TerminalNode
	TABLE() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	PLUS() antlr.TerminalNode
	MINUS() antlr.TerminalNode

	// IsFieldsCommandContext differentiates from other interfaces.
	IsFieldsCommandContext()
}

type FieldsCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldsCommandContext() *FieldsCommandContext {
	var p = new(FieldsCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldsCommand
	return p
}

func InitEmptyFieldsCommandContext(p *FieldsCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldsCommand
}

func (*FieldsCommandContext) IsFieldsCommandContext() {}

func NewFieldsCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldsCommandContext {
	var p = new(FieldsCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_fieldsCommand

	return p
}

func (s *FieldsCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldsCommandContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *FieldsCommandContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *FieldsCommandContext) FIELDS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserFIELDS, 0)
}

func (s *FieldsCommandContext) TABLE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserTABLE, 0)
}

func (s *FieldsCommandContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *FieldsCommandContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *FieldsCommandContext) PLUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserPLUS, 0)
}

func (s *FieldsCommandContext) MINUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserMINUS, 0)
}

func (s *FieldsCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldsCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldsCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterFieldsCommand(s)
	}
}

func (s *FieldsCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitFieldsCommand(s)
	}
}

func (s *FieldsCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitFieldsCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) FieldsCommand() (localctx IFieldsCommandContext) {
	localctx = NewFieldsCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, SPL2ParserRULE_fieldsCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(231)
		_la = p.GetTokenStream().LA(1)

		if !(_la == SPL2ParserFIELDS || _la == SPL2ParserTABLE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	p.SetState(233)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserPLUS || _la == SPL2ParserMINUS {
		{
			p.SetState(232)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}
	{
		p.SetState(235)
		p.Identifier()
	}
	p.SetState(240)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(236)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(237)
			p.Identifier()
		}

		p.SetState(242)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRexCommandContext is an interface to support dynamic dispatch.
type IRexCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	REX() antlr.TerminalNode
	REGEX() antlr.TerminalNode
	StringLiteral() IStringLiteralContext
	RAW_STRING() antlr.TerminalNode
	AllRexOption() []IRexOptionContext
	RexOption(i int) IRexOptionContext

	// IsRexCommandContext differentiates from other interfaces.
	IsRexCommandContext()
}

type RexCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRexCommandContext() *RexCommandContext {
	var p = new(RexCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_rexCommand
	return p
}

func InitEmptyRexCommandContext(p *RexCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_rexCommand
}

func (*RexCommandContext) IsRexCommandContext() {}

func NewRexCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RexCommandContext {
	var p = new(RexCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_rexCommand

	return p
}

func (s *RexCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *RexCommandContext) REX() antlr.TerminalNode {
	return s.GetToken(SPL2ParserREX, 0)
}

func (s *RexCommandContext) REGEX() antlr.TerminalNode {
	return s.GetToken(SPL2ParserREGEX, 0)
}

func (s *RexCommandContext) StringLiteral() IStringLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStringLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStringLiteralContext)
}

func (s *RexCommandContext) RAW_STRING() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRAW_STRING, 0)
}

func (s *RexCommandContext) AllRexOption() []IRexOptionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRexOptionContext); ok {
			len++
		}
	}

	tst := make([]IRexOptionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRexOptionContext); ok {
			tst[i] = t.(IRexOptionContext)
			i++
		}
	}

	return tst
}

func (s *RexCommandContext) RexOption(i int) IRexOptionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRexOptionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRexOptionContext)
}

func (s *RexCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RexCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RexCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterRexCommand(s)
	}
}

func (s *RexCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitRexCommand(s)
	}
}

func (s *RexCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitRexCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) RexCommand() (localctx IRexCommandContext) {
	localctx = NewRexCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, SPL2ParserRULE_rexCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(243)
		p.Match(SPL2ParserREX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(247)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserIDENTIFIER {
		{
			p.SetState(244)
			p.RexOption()
		}

		p.SetState(249)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(253)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserREGEX:
		{
			p.SetState(250)
			p.Match(SPL2ParserREGEX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		{
			p.SetState(251)
			p.StringLiteral()
		}

	case SPL2ParserRAW_STRING:
		{
			p.SetState(252)
			p.Match(SPL2ParserRAW_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRexOptionContext is an interface to support dynamic dispatch.
type IRexOptionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	Identifier() IIdentifierContext
	NUMBER() antlr.TerminalNode

	// IsRexOptionContext differentiates from other interfaces.
	IsRexOptionContext()
}

type RexOptionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRexOptionContext() *RexOptionContext {
	var p = new(RexOptionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_rexOption
	return p
}

func InitEmptyRexOptionContext(p *RexOptionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_rexOption
}

func (*RexOptionContext) IsRexOptionContext() {}

func NewRexOptionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RexOptionContext {
	var p = new(RexOptionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_rexOption

	return p
}

func (s *RexOptionContext) GetParser() antlr.Parser { return s.parser }

func (s *RexOptionContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *RexOptionContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *RexOptionContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *RexOptionContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNUMBER, 0)
}

func (s *RexOptionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RexOptionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RexOptionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterRexOption(s)
	}
}

func (s *RexOptionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitRexOption(s)
	}
}

func (s *RexOptionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitRexOption(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) RexOption() (localctx IRexOptionContext) {
	localctx = NewRexOptionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, SPL2ParserRULE_rexOption)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(255)
		p.Match(SPL2ParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(256)
		p.Match(SPL2ParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(259)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		{
			p.SetState(257)
			p.Identifier()
		}

	case SPL2ParserNUMBER:
		{
			p.SetState(258)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEmbeddedCommandContext is an interface to support dynamic dispatch.
type IEmbeddedCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EmbeddedText() IEmbeddedTextContext
	SPL1() antlr.TerminalNode

	// IsEmbeddedCommandContext differentiates from other interfaces.
	IsEmbeddedCommandContext()
}

type EmbeddedCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEmbeddedCommandContext() *EmbeddedCommandContext {
	var p = new(EmbeddedCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_embeddedCommand
	return p
}

func InitEmptyEmbeddedCommandContext(p *EmbeddedCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_embeddedCommand
}

func (*EmbeddedCommandContext) IsEmbeddedCommandContext() {}

func NewEmbeddedCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EmbeddedCommandContext {
	var p = new(EmbeddedCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_embeddedCommand

	return p
}

func (s *EmbeddedCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *EmbeddedCommandContext) EmbeddedText() IEmbeddedTextContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEmbeddedTextContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEmbeddedTextContext)
}

func (s *EmbeddedCommandContext) SPL1() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSPL1, 0)
}

func (s *EmbeddedCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EmbeddedCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EmbeddedCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterEmbeddedCommand(s)
	}
}

func (s *EmbeddedCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitEmbeddedCommand(s)
	}
}

func (s *EmbeddedCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitEmbeddedCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) EmbeddedCommand() (localctx IEmbeddedCommandContext) {
	localctx = NewEmbeddedCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, SPL2ParserRULE_embeddedCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(262)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserSPL1 {
		{
			p.SetState(261)
			p.Match(SPL2ParserSPL1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(264)
		p.EmbeddedText()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEmbeddedTextContext is an interface to support dynamic dispatch.
type IEmbeddedTextContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BACKTICK() antlr.TerminalNode
	EMBEDDED_END() antlr.TerminalNode
	EMBEDDED_TEXT() antlr.TerminalNode

	// IsEmbeddedTextContext differentiates from other interfaces.
	IsEmbeddedTextContext()
}

type EmbeddedTextContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEmbeddedTextContext() *EmbeddedTextContext {
	var p = new(EmbeddedTextContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_embeddedText
	return p
}

func InitEmptyEmbeddedTextContext(p *EmbeddedTextContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_embeddedText
}

func (*EmbeddedTextContext) IsEmbeddedTextContext() {}

func NewEmbeddedTextContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EmbeddedTextContext {
	var p = new(EmbeddedTextContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_embeddedText

	return p
}

func (s *EmbeddedTextContext) GetParser() antlr.Parser { return s.parser }

func (s *EmbeddedTextContext) BACKTICK() antlr.TerminalNode {
	return s.GetToken(SPL2ParserBACKTICK, 0)
}

func (s *EmbeddedTextContext) EMBEDDED_END() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEMBEDDED_END, 0)
}

func (s *EmbeddedTextContext) EMBEDDED_TEXT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEMBEDDED_TEXT, 0)
}

func (s *EmbeddedTextContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EmbeddedTextContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EmbeddedTextContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterEmbeddedText(s)
	}
}

func (s *EmbeddedTextContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitEmbeddedText(s)
	}
}

func (s *EmbeddedTextContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitEmbeddedText(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) EmbeddedText() (localctx IEmbeddedTextContext) {
	localctx = NewEmbeddedTextContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, SPL2ParserRULE_embeddedText)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(266)
		p.Match(SPL2ParserBACKTICK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(268)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserEMBEDDED_TEXT {
		{
			p.SetState(267)
			p.Match(SPL2ParserEMBEDDED_TEXT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(270)
		p.Match(SPL2ParserEMBEDDED_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IModuleSuffixContext is an interface to support dynamic dispatch.
type IModuleSuffixContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SEMI() antlr.TerminalNode

	// IsModuleSuffixContext differentiates from other interfaces.
	IsModuleSuffixContext()
}

type ModuleSuffixContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyModuleSuffixContext() *ModuleSuffixContext {
	var p = new(ModuleSuffixContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_moduleSuffix
	return p
}

func InitEmptyModuleSuffixContext(p *ModuleSuffixContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_moduleSuffix
}

func (*ModuleSuffixContext) IsModuleSuffixContext() {}

func NewModuleSuffixContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ModuleSuffixContext {
	var p = new(ModuleSuffixContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_moduleSuffix

	return p
}

func (s *ModuleSuffixContext) GetParser() antlr.Parser { return s.parser }

func (s *ModuleSuffixContext) SEMI() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSEMI, 0)
}

func (s *ModuleSuffixContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ModuleSuffixContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ModuleSuffixContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterModuleSuffix(s)
	}
}

func (s *ModuleSuffixContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitModuleSuffix(s)
	}
}

func (s *ModuleSuffixContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitModuleSuffix(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) ModuleSuffix() (localctx IModuleSuffixContext) {
	localctx = NewModuleSuffixContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, SPL2ParserRULE_moduleSuffix)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(272)
		p.Match(SPL2ParserSEMI)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(276)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1+1 {
			p.SetState(273)
			p.MatchWildcard()

		}
		p.SetState(278)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IModuleDeclarationContext is an interface to support dynamic dispatch.
type IModuleDeclarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IMPORT() antlr.TerminalNode
	EXPORT() antlr.TerminalNode
	FUNCTION() antlr.TerminalNode
	LOCAL() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode

	// IsModuleDeclarationContext differentiates from other interfaces.
	IsModuleDeclarationContext()
}

type ModuleDeclarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyModuleDeclarationContext() *ModuleDeclarationContext {
	var p = new(ModuleDeclarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_moduleDeclaration
	return p
}

func InitEmptyModuleDeclarationContext(p *ModuleDeclarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_moduleDeclaration
}

func (*ModuleDeclarationContext) IsModuleDeclarationContext() {}

func NewModuleDeclarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ModuleDeclarationContext {
	var p = new(ModuleDeclarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_moduleDeclaration

	return p
}

func (s *ModuleDeclarationContext) GetParser() antlr.Parser { return s.parser }

func (s *ModuleDeclarationContext) IMPORT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIMPORT, 0)
}

func (s *ModuleDeclarationContext) EXPORT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEXPORT, 0)
}

func (s *ModuleDeclarationContext) FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPL2ParserFUNCTION, 0)
}

func (s *ModuleDeclarationContext) LOCAL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLOCAL, 0)
}

func (s *ModuleDeclarationContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *ModuleDeclarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ModuleDeclarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ModuleDeclarationContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterModuleDeclaration(s)
	}
}

func (s *ModuleDeclarationContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitModuleDeclaration(s)
	}
}

func (s *ModuleDeclarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitModuleDeclaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) ModuleDeclaration() (localctx IModuleDeclarationContext) {
	localctx = NewModuleDeclarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, SPL2ParserRULE_moduleDeclaration)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(284)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserIMPORT:
		{
			p.SetState(279)
			p.Match(SPL2ParserIMPORT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserEXPORT:
		{
			p.SetState(280)
			p.Match(SPL2ParserEXPORT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserFUNCTION:
		{
			p.SetState(281)
			p.Match(SPL2ParserFUNCTION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserLOCAL:
		{
			p.SetState(282)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(283)
			p.Match(SPL2ParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.SetState(289)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 23, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1+1 {
			p.SetState(286)
			p.MatchWildcard()

		}
		p.SetState(291)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 23, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchCommandContext is an interface to support dynamic dispatch.
type ISearchCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SEARCH() antlr.TerminalNode
	SearchExpression() ISearchExpressionContext

	// IsSearchCommandContext differentiates from other interfaces.
	IsSearchCommandContext()
}

type SearchCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchCommandContext() *SearchCommandContext {
	var p = new(SearchCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchCommand
	return p
}

func InitEmptySearchCommandContext(p *SearchCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchCommand
}

func (*SearchCommandContext) IsSearchCommandContext() {}

func NewSearchCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchCommandContext {
	var p = new(SearchCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchCommand

	return p
}

func (s *SearchCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchCommandContext) SEARCH() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSEARCH, 0)
}

func (s *SearchCommandContext) SearchExpression() ISearchExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchExpressionContext)
}

func (s *SearchCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchCommand(s)
	}
}

func (s *SearchCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchCommand(s)
	}
}

func (s *SearchCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchCommand() (localctx ISearchCommandContext) {
	localctx = NewSearchCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, SPL2ParserRULE_searchCommand)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(292)
		p.Match(SPL2ParserSEARCH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(293)
		p.SearchExpression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IImplicitSearchContext is an interface to support dynamic dispatch.
type IImplicitSearchContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SearchExpression() ISearchExpressionContext

	// IsImplicitSearchContext differentiates from other interfaces.
	IsImplicitSearchContext()
}

type ImplicitSearchContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyImplicitSearchContext() *ImplicitSearchContext {
	var p = new(ImplicitSearchContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_implicitSearch
	return p
}

func InitEmptyImplicitSearchContext(p *ImplicitSearchContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_implicitSearch
}

func (*ImplicitSearchContext) IsImplicitSearchContext() {}

func NewImplicitSearchContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ImplicitSearchContext {
	var p = new(ImplicitSearchContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_implicitSearch

	return p
}

func (s *ImplicitSearchContext) GetParser() antlr.Parser { return s.parser }

func (s *ImplicitSearchContext) SearchExpression() ISearchExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchExpressionContext)
}

func (s *ImplicitSearchContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ImplicitSearchContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ImplicitSearchContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterImplicitSearch(s)
	}
}

func (s *ImplicitSearchContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitImplicitSearch(s)
	}
}

func (s *ImplicitSearchContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitImplicitSearch(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) ImplicitSearch() (localctx IImplicitSearchContext) {
	localctx = NewImplicitSearchContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, SPL2ParserRULE_implicitSearch)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(295)

	if !(p.GetTokenStream().LA(1) == SPL2ParserINDEX && p.GetTokenStream().LA(2) == SPL2ParserASSIGN) {
		p.SetError(antlr.NewFailedPredicateException(p, "p.GetTokenStream().LA(1) == SPL2ParserINDEX && p.GetTokenStream().LA(2) == SPL2ParserASSIGN", ""))
		goto errorExit
	}
	{
		p.SetState(296)
		p.SearchExpression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchExpressionContext is an interface to support dynamic dispatch.
type ISearchExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SearchXor() ISearchXorContext

	// IsSearchExpressionContext differentiates from other interfaces.
	IsSearchExpressionContext()
}

type SearchExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchExpressionContext() *SearchExpressionContext {
	var p = new(SearchExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchExpression
	return p
}

func InitEmptySearchExpressionContext(p *SearchExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchExpression
}

func (*SearchExpressionContext) IsSearchExpressionContext() {}

func NewSearchExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchExpressionContext {
	var p = new(SearchExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchExpression

	return p
}

func (s *SearchExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchExpressionContext) SearchXor() ISearchXorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchXorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchXorContext)
}

func (s *SearchExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchExpression(s)
	}
}

func (s *SearchExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchExpression(s)
	}
}

func (s *SearchExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchExpression() (localctx ISearchExpressionContext) {
	localctx = NewSearchExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, SPL2ParserRULE_searchExpression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(298)
		p.SearchXor()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchXorContext is an interface to support dynamic dispatch.
type ISearchXorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSearchAnd() []ISearchAndContext
	SearchAnd(i int) ISearchAndContext
	AllXOR() []antlr.TerminalNode
	XOR(i int) antlr.TerminalNode

	// IsSearchXorContext differentiates from other interfaces.
	IsSearchXorContext()
}

type SearchXorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchXorContext() *SearchXorContext {
	var p = new(SearchXorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchXor
	return p
}

func InitEmptySearchXorContext(p *SearchXorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchXor
}

func (*SearchXorContext) IsSearchXorContext() {}

func NewSearchXorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchXorContext {
	var p = new(SearchXorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchXor

	return p
}

func (s *SearchXorContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchXorContext) AllSearchAnd() []ISearchAndContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISearchAndContext); ok {
			len++
		}
	}

	tst := make([]ISearchAndContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISearchAndContext); ok {
			tst[i] = t.(ISearchAndContext)
			i++
		}
	}

	return tst
}

func (s *SearchXorContext) SearchAnd(i int) ISearchAndContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchAndContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchAndContext)
}

func (s *SearchXorContext) AllXOR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserXOR)
}

func (s *SearchXorContext) XOR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserXOR, i)
}

func (s *SearchXorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchXorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchXorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchXor(s)
	}
}

func (s *SearchXorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchXor(s)
	}
}

func (s *SearchXorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchXor(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchXor() (localctx ISearchXorContext) {
	localctx = NewSearchXorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, SPL2ParserRULE_searchXor)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(300)
		p.SearchAnd()
	}
	p.SetState(305)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserXOR {
		{
			p.SetState(301)
			p.Match(SPL2ParserXOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(302)
			p.SearchAnd()
		}

		p.SetState(307)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchAndContext is an interface to support dynamic dispatch.
type ISearchAndContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSearchOr() []ISearchOrContext
	SearchOr(i int) ISearchOrContext
	AllAND() []antlr.TerminalNode
	AND(i int) antlr.TerminalNode

	// IsSearchAndContext differentiates from other interfaces.
	IsSearchAndContext()
}

type SearchAndContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchAndContext() *SearchAndContext {
	var p = new(SearchAndContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchAnd
	return p
}

func InitEmptySearchAndContext(p *SearchAndContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchAnd
}

func (*SearchAndContext) IsSearchAndContext() {}

func NewSearchAndContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchAndContext {
	var p = new(SearchAndContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchAnd

	return p
}

func (s *SearchAndContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchAndContext) AllSearchOr() []ISearchOrContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISearchOrContext); ok {
			len++
		}
	}

	tst := make([]ISearchOrContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISearchOrContext); ok {
			tst[i] = t.(ISearchOrContext)
			i++
		}
	}

	return tst
}

func (s *SearchAndContext) SearchOr(i int) ISearchOrContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchOrContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchOrContext)
}

func (s *SearchAndContext) AllAND() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserAND)
}

func (s *SearchAndContext) AND(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserAND, i)
}

func (s *SearchAndContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchAndContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchAndContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchAnd(s)
	}
}

func (s *SearchAndContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchAnd(s)
	}
}

func (s *SearchAndContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchAnd(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchAnd() (localctx ISearchAndContext) {
	localctx = NewSearchAndContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, SPL2ParserRULE_searchAnd)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(308)
		p.SearchOr()
	}
	p.SetState(315)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&6269574730767138832) != 0 {
		p.SetState(310)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserAND {
			{
				p.SetState(309)
				p.Match(SPL2ParserAND)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(312)
			p.SearchOr()
		}

		p.SetState(317)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchOrContext is an interface to support dynamic dispatch.
type ISearchOrContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSearchNot() []ISearchNotContext
	SearchNot(i int) ISearchNotContext
	AllOR() []antlr.TerminalNode
	OR(i int) antlr.TerminalNode

	// IsSearchOrContext differentiates from other interfaces.
	IsSearchOrContext()
}

type SearchOrContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchOrContext() *SearchOrContext {
	var p = new(SearchOrContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchOr
	return p
}

func InitEmptySearchOrContext(p *SearchOrContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchOr
}

func (*SearchOrContext) IsSearchOrContext() {}

func NewSearchOrContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchOrContext {
	var p = new(SearchOrContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchOr

	return p
}

func (s *SearchOrContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchOrContext) AllSearchNot() []ISearchNotContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISearchNotContext); ok {
			len++
		}
	}

	tst := make([]ISearchNotContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISearchNotContext); ok {
			tst[i] = t.(ISearchNotContext)
			i++
		}
	}

	return tst
}

func (s *SearchOrContext) SearchNot(i int) ISearchNotContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchNotContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchNotContext)
}

func (s *SearchOrContext) AllOR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserOR)
}

func (s *SearchOrContext) OR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserOR, i)
}

func (s *SearchOrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchOrContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchOrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchOr(s)
	}
}

func (s *SearchOrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchOr(s)
	}
}

func (s *SearchOrContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchOr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchOr() (localctx ISearchOrContext) {
	localctx = NewSearchOrContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, SPL2ParserRULE_searchOr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(318)
		p.SearchNot()
	}
	p.SetState(323)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserOR {
		{
			p.SetState(319)
			p.Match(SPL2ParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(320)
			p.SearchNot()
		}

		p.SetState(325)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchNotContext is an interface to support dynamic dispatch.
type ISearchNotContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NOT() antlr.TerminalNode
	SearchNot() ISearchNotContext
	SearchAtom() ISearchAtomContext

	// IsSearchNotContext differentiates from other interfaces.
	IsSearchNotContext()
}

type SearchNotContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchNotContext() *SearchNotContext {
	var p = new(SearchNotContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchNot
	return p
}

func InitEmptySearchNotContext(p *SearchNotContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchNot
}

func (*SearchNotContext) IsSearchNotContext() {}

func NewSearchNotContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchNotContext {
	var p = new(SearchNotContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchNot

	return p
}

func (s *SearchNotContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNOT, 0)
}

func (s *SearchNotContext) SearchNot() ISearchNotContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchNotContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchNotContext)
}

func (s *SearchNotContext) SearchAtom() ISearchAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchAtomContext)
}

func (s *SearchNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchNotContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchNotContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchNot(s)
	}
}

func (s *SearchNotContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchNot(s)
	}
}

func (s *SearchNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchNot(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchNot() (localctx ISearchNotContext) {
	localctx = NewSearchNotContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, SPL2ParserRULE_searchNot)
	p.SetState(329)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserNOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(326)
			p.Match(SPL2ParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(327)
			p.SearchNot()
		}

	case SPL2ParserINDEX, SPL2ParserSTAR, SPL2ParserLPAREN, SPL2ParserRAW_STRING, SPL2ParserDQUOTE, SPL2ParserSQUOTE, SPL2ParserNUMBER, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(328)
			p.SearchAtom()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchAtomContext is an interface to support dynamic dispatch.
type ISearchAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	SearchExpression() ISearchExpressionContext
	RPAREN() antlr.TerminalNode
	Identifier() IIdentifierContext
	Comparison() IComparisonContext
	AllSearchValue() []ISearchValueContext
	SearchValue(i int) ISearchValueContext
	IN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsSearchAtomContext differentiates from other interfaces.
	IsSearchAtomContext()
}

type SearchAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchAtomContext() *SearchAtomContext {
	var p = new(SearchAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchAtom
	return p
}

func InitEmptySearchAtomContext(p *SearchAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchAtom
}

func (*SearchAtomContext) IsSearchAtomContext() {}

func NewSearchAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchAtomContext {
	var p = new(SearchAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchAtom

	return p
}

func (s *SearchAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchAtomContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLPAREN, 0)
}

func (s *SearchAtomContext) SearchExpression() ISearchExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchExpressionContext)
}

func (s *SearchAtomContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRPAREN, 0)
}

func (s *SearchAtomContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *SearchAtomContext) Comparison() IComparisonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonContext)
}

func (s *SearchAtomContext) AllSearchValue() []ISearchValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISearchValueContext); ok {
			len++
		}
	}

	tst := make([]ISearchValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISearchValueContext); ok {
			tst[i] = t.(ISearchValueContext)
			i++
		}
	}

	return tst
}

func (s *SearchAtomContext) SearchValue(i int) ISearchValueContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchValueContext)
}

func (s *SearchAtomContext) IN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIN, 0)
}

func (s *SearchAtomContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *SearchAtomContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *SearchAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchAtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchAtom(s)
	}
}

func (s *SearchAtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchAtom(s)
	}
}

func (s *SearchAtomContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchAtom(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchAtom() (localctx ISearchAtomContext) {
	localctx = NewSearchAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 52, SPL2ParserRULE_searchAtom)
	var _la int

	p.SetState(352)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 30, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(331)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(332)
			p.SearchExpression()
		}
		{
			p.SetState(333)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(335)
			p.Identifier()
		}
		{
			p.SetState(336)
			p.Comparison()
		}
		{
			p.SetState(337)
			p.SearchValue()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(339)
			p.Identifier()
		}
		{
			p.SetState(340)
			p.Match(SPL2ParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(341)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(342)
			p.SearchValue()
		}
		p.SetState(345)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == SPL2ParserCOMMA {
			{
				p.SetState(343)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(344)
				p.SearchValue()
			}

			p.SetState(347)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(349)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(351)
			p.SearchValue()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchValueContext is an interface to support dynamic dispatch.
type ISearchValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	NUMBER() antlr.TerminalNode
	StringLiteral() IStringLiteralContext
	RAW_STRING() antlr.TerminalNode
	STAR() antlr.TerminalNode

	// IsSearchValueContext differentiates from other interfaces.
	IsSearchValueContext()
}

type SearchValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchValueContext() *SearchValueContext {
	var p = new(SearchValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchValue
	return p
}

func InitEmptySearchValueContext(p *SearchValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchValue
}

func (*SearchValueContext) IsSearchValueContext() {}

func NewSearchValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchValueContext {
	var p = new(SearchValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchValue

	return p
}

func (s *SearchValueContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchValueContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *SearchValueContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNUMBER, 0)
}

func (s *SearchValueContext) StringLiteral() IStringLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStringLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStringLiteralContext)
}

func (s *SearchValueContext) RAW_STRING() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRAW_STRING, 0)
}

func (s *SearchValueContext) STAR() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTAR, 0)
}

func (s *SearchValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchValue(s)
	}
}

func (s *SearchValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchValue(s)
	}
}

func (s *SearchValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchValue() (localctx ISearchValueContext) {
	localctx = NewSearchValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, SPL2ParserRULE_searchValue)
	p.SetState(359)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(354)
			p.Identifier()
		}

	case SPL2ParserNUMBER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(355)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(356)
			p.StringLiteral()
		}

	case SPL2ParserRAW_STRING:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(357)
			p.Match(SPL2ParserRAW_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserSTAR:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(358)
			p.Match(SPL2ParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LambdaExpression() ILambdaExpressionContext
	XorExpression() IXorExpressionContext

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) LambdaExpression() ILambdaExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILambdaExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILambdaExpressionContext)
}

func (s *ExpressionContext) XorExpression() IXorExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IXorExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IXorExpressionContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (s *ExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, SPL2ParserRULE_expression)
	p.SetState(363)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 32, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(361)
			p.LambdaExpression()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(362)
			p.XorExpression()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IXorExpressionContext is an interface to support dynamic dispatch.
type IXorExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllOrExpression() []IOrExpressionContext
	OrExpression(i int) IOrExpressionContext
	AllLogicalXor() []ILogicalXorContext
	LogicalXor(i int) ILogicalXorContext

	// IsXorExpressionContext differentiates from other interfaces.
	IsXorExpressionContext()
}

type XorExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyXorExpressionContext() *XorExpressionContext {
	var p = new(XorExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_xorExpression
	return p
}

func InitEmptyXorExpressionContext(p *XorExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_xorExpression
}

func (*XorExpressionContext) IsXorExpressionContext() {}

func NewXorExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *XorExpressionContext {
	var p = new(XorExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_xorExpression

	return p
}

func (s *XorExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *XorExpressionContext) AllOrExpression() []IOrExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrExpressionContext); ok {
			len++
		}
	}

	tst := make([]IOrExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrExpressionContext); ok {
			tst[i] = t.(IOrExpressionContext)
			i++
		}
	}

	return tst
}

func (s *XorExpressionContext) OrExpression(i int) IOrExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrExpressionContext)
}

func (s *XorExpressionContext) AllLogicalXor() []ILogicalXorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILogicalXorContext); ok {
			len++
		}
	}

	tst := make([]ILogicalXorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILogicalXorContext); ok {
			tst[i] = t.(ILogicalXorContext)
			i++
		}
	}

	return tst
}

func (s *XorExpressionContext) LogicalXor(i int) ILogicalXorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogicalXorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogicalXorContext)
}

func (s *XorExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *XorExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *XorExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterXorExpression(s)
	}
}

func (s *XorExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitXorExpression(s)
	}
}

func (s *XorExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitXorExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) XorExpression() (localctx IXorExpressionContext) {
	localctx = NewXorExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, SPL2ParserRULE_xorExpression)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(365)
		p.OrExpression()
	}
	p.SetState(371)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(366)
				p.LogicalXor()
			}
			{
				p.SetState(367)
				p.OrExpression()
			}

		}
		p.SetState(373)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrExpressionContext is an interface to support dynamic dispatch.
type IOrExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAndExpression() []IAndExpressionContext
	AndExpression(i int) IAndExpressionContext
	AllLogicalOr() []ILogicalOrContext
	LogicalOr(i int) ILogicalOrContext

	// IsOrExpressionContext differentiates from other interfaces.
	IsOrExpressionContext()
}

type OrExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrExpressionContext() *OrExpressionContext {
	var p = new(OrExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_orExpression
	return p
}

func InitEmptyOrExpressionContext(p *OrExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_orExpression
}

func (*OrExpressionContext) IsOrExpressionContext() {}

func NewOrExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrExpressionContext {
	var p = new(OrExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_orExpression

	return p
}

func (s *OrExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *OrExpressionContext) AllAndExpression() []IAndExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAndExpressionContext); ok {
			len++
		}
	}

	tst := make([]IAndExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAndExpressionContext); ok {
			tst[i] = t.(IAndExpressionContext)
			i++
		}
	}

	return tst
}

func (s *OrExpressionContext) AndExpression(i int) IAndExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAndExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAndExpressionContext)
}

func (s *OrExpressionContext) AllLogicalOr() []ILogicalOrContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILogicalOrContext); ok {
			len++
		}
	}

	tst := make([]ILogicalOrContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILogicalOrContext); ok {
			tst[i] = t.(ILogicalOrContext)
			i++
		}
	}

	return tst
}

func (s *OrExpressionContext) LogicalOr(i int) ILogicalOrContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogicalOrContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogicalOrContext)
}

func (s *OrExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterOrExpression(s)
	}
}

func (s *OrExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitOrExpression(s)
	}
}

func (s *OrExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitOrExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) OrExpression() (localctx IOrExpressionContext) {
	localctx = NewOrExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, SPL2ParserRULE_orExpression)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(374)
		p.AndExpression()
	}
	p.SetState(380)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(375)
				p.LogicalOr()
			}
			{
				p.SetState(376)
				p.AndExpression()
			}

		}
		p.SetState(382)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAndExpressionContext is an interface to support dynamic dispatch.
type IAndExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNotExpression() []INotExpressionContext
	NotExpression(i int) INotExpressionContext
	AllLogicalAnd() []ILogicalAndContext
	LogicalAnd(i int) ILogicalAndContext

	// IsAndExpressionContext differentiates from other interfaces.
	IsAndExpressionContext()
}

type AndExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAndExpressionContext() *AndExpressionContext {
	var p = new(AndExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_andExpression
	return p
}

func InitEmptyAndExpressionContext(p *AndExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_andExpression
}

func (*AndExpressionContext) IsAndExpressionContext() {}

func NewAndExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AndExpressionContext {
	var p = new(AndExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_andExpression

	return p
}

func (s *AndExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *AndExpressionContext) AllNotExpression() []INotExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INotExpressionContext); ok {
			len++
		}
	}

	tst := make([]INotExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INotExpressionContext); ok {
			tst[i] = t.(INotExpressionContext)
			i++
		}
	}

	return tst
}

func (s *AndExpressionContext) NotExpression(i int) INotExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INotExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(INotExpressionContext)
}

func (s *AndExpressionContext) AllLogicalAnd() []ILogicalAndContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILogicalAndContext); ok {
			len++
		}
	}

	tst := make([]ILogicalAndContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILogicalAndContext); ok {
			tst[i] = t.(ILogicalAndContext)
			i++
		}
	}

	return tst
}

func (s *AndExpressionContext) LogicalAnd(i int) ILogicalAndContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogicalAndContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogicalAndContext)
}

func (s *AndExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AndExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterAndExpression(s)
	}
}

func (s *AndExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitAndExpression(s)
	}
}

func (s *AndExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitAndExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) AndExpression() (localctx IAndExpressionContext) {
	localctx = NewAndExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, SPL2ParserRULE_andExpression)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(383)
		p.NotExpression()
	}
	p.SetState(389)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(384)
				p.LogicalAnd()
			}
			{
				p.SetState(385)
				p.NotExpression()
			}

		}
		p.SetState(391)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INotExpressionContext is an interface to support dynamic dispatch.
type INotExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Predicate() IPredicateContext
	LogicalNot() ILogicalNotContext
	NotExpression() INotExpressionContext

	// IsNotExpressionContext differentiates from other interfaces.
	IsNotExpressionContext()
}

type NotExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNotExpressionContext() *NotExpressionContext {
	var p = new(NotExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_notExpression
	return p
}

func InitEmptyNotExpressionContext(p *NotExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_notExpression
}

func (*NotExpressionContext) IsNotExpressionContext() {}

func NewNotExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NotExpressionContext {
	var p = new(NotExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_notExpression

	return p
}

func (s *NotExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *NotExpressionContext) Predicate() IPredicateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPredicateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPredicateContext)
}

func (s *NotExpressionContext) LogicalNot() ILogicalNotContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogicalNotContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogicalNotContext)
}

func (s *NotExpressionContext) NotExpression() INotExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INotExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INotExpressionContext)
}

func (s *NotExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NotExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NotExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterNotExpression(s)
	}
}

func (s *NotExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitNotExpression(s)
	}
}

func (s *NotExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitNotExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) NotExpression() (localctx INotExpressionContext) {
	localctx = NewNotExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, SPL2ParserRULE_notExpression)
	p.SetState(396)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 36, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(392)
			p.Predicate()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(393)
			p.LogicalNot()
		}
		{
			p.SetState(394)
			p.NotExpression()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPredicateContext is an interface to support dynamic dispatch.
type IPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAdditive() []IAdditiveContext
	Additive(i int) IAdditiveContext
	Comparison() IComparisonContext
	BetweenOperator() IBetweenOperatorContext
	BetweenConjunction() IBetweenConjunctionContext
	IN() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	RPAREN() antlr.TerminalNode
	LIKE() antlr.TerminalNode
	IS() antlr.TerminalNode
	TYPE() antlr.TerminalNode
	LogicalNot() ILogicalNotContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	NULL() antlr.TerminalNode
	NULL_TEST() antlr.TerminalNode

	// IsPredicateContext differentiates from other interfaces.
	IsPredicateContext()
}

type PredicateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPredicateContext() *PredicateContext {
	var p = new(PredicateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_predicate
	return p
}

func InitEmptyPredicateContext(p *PredicateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_predicate
}

func (*PredicateContext) IsPredicateContext() {}

func NewPredicateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PredicateContext {
	var p = new(PredicateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_predicate

	return p
}

func (s *PredicateContext) GetParser() antlr.Parser { return s.parser }

func (s *PredicateContext) AllAdditive() []IAdditiveContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAdditiveContext); ok {
			len++
		}
	}

	tst := make([]IAdditiveContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAdditiveContext); ok {
			tst[i] = t.(IAdditiveContext)
			i++
		}
	}

	return tst
}

func (s *PredicateContext) Additive(i int) IAdditiveContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAdditiveContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAdditiveContext)
}

func (s *PredicateContext) Comparison() IComparisonContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonContext)
}

func (s *PredicateContext) BetweenOperator() IBetweenOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBetweenOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBetweenOperatorContext)
}

func (s *PredicateContext) BetweenConjunction() IBetweenConjunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBetweenConjunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBetweenConjunctionContext)
}

func (s *PredicateContext) IN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIN, 0)
}

func (s *PredicateContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLPAREN, 0)
}

func (s *PredicateContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *PredicateContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PredicateContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRPAREN, 0)
}

func (s *PredicateContext) LIKE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLIKE, 0)
}

func (s *PredicateContext) IS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIS, 0)
}

func (s *PredicateContext) TYPE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserTYPE, 0)
}

func (s *PredicateContext) LogicalNot() ILogicalNotContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILogicalNotContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILogicalNotContext)
}

func (s *PredicateContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *PredicateContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *PredicateContext) NULL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNULL, 0)
}

func (s *PredicateContext) NULL_TEST() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNULL_TEST, 0)
}

func (s *PredicateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PredicateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PredicateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterPredicate(s)
	}
}

func (s *PredicateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitPredicate(s)
	}
}

func (s *PredicateContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitPredicate(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Predicate() (localctx IPredicateContext) {
	localctx = NewPredicateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, SPL2ParserRULE_predicate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(398)
		p.Additive()
	}
	p.SetState(441)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(399)
			p.Comparison()
		}
		{
			p.SetState(400)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) == 2 {
		p.SetState(403)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 37, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(402)
				p.LogicalNot()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}
		{
			p.SetState(405)
			p.BetweenOperator()
		}
		{
			p.SetState(406)
			p.Additive()
		}
		{
			p.SetState(407)
			p.BetweenConjunction()
		}
		{
			p.SetState(408)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) == 3 {
		p.SetState(411)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 38, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(410)
				p.LogicalNot()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}
		{
			p.SetState(413)
			p.Match(SPL2ParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(414)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(415)
			p.Expression()
		}
		p.SetState(420)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(416)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(417)
				p.Expression()
			}

			p.SetState(422)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(423)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) == 4 {
		p.SetState(426)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 40, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(425)
				p.LogicalNot()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}
		{
			p.SetState(428)
			p.Match(SPL2ParserLIKE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(429)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) == 5 {
		{
			p.SetState(430)
			p.Match(SPL2ParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(439)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 43, p.GetParserRuleContext()) {
		case 1:
			p.SetState(432)
			p.GetErrorHandler().Sync(p)

			if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 41, p.GetParserRuleContext()) == 1 {
				{
					p.SetState(431)
					p.LogicalNot()
				}

			} else if p.HasError() { // JIM
				goto errorExit
			}
			{
				p.SetState(434)
				_la = p.GetTokenStream().LA(1)

				if !(_la == SPL2ParserNULL || _la == SPL2ParserNULL_TEST) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

		case 2:
			p.SetState(436)
			p.GetErrorHandler().Sync(p)

			if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 42, p.GetParserRuleContext()) == 1 {
				{
					p.SetState(435)
					p.LogicalNot()
				}

			} else if p.HasError() { // JIM
				goto errorExit
			}
			{
				p.SetState(438)
				p.Match(SPL2ParserTYPE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case antlr.ATNInvalidAltNumber:
			goto errorExit
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogicalAndContext is an interface to support dynamic dispatch.
type ILogicalAndContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AND() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsLogicalAndContext differentiates from other interfaces.
	IsLogicalAndContext()
}

type LogicalAndContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogicalAndContext() *LogicalAndContext {
	var p = new(LogicalAndContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalAnd
	return p
}

func InitEmptyLogicalAndContext(p *LogicalAndContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalAnd
}

func (*LogicalAndContext) IsLogicalAndContext() {}

func NewLogicalAndContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LogicalAndContext {
	var p = new(LogicalAndContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_logicalAnd

	return p
}

func (s *LogicalAndContext) GetParser() antlr.Parser { return s.parser }

func (s *LogicalAndContext) AND() antlr.TerminalNode {
	return s.GetToken(SPL2ParserAND, 0)
}

func (s *LogicalAndContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *LogicalAndContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalAndContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LogicalAndContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLogicalAnd(s)
	}
}

func (s *LogicalAndContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLogicalAnd(s)
	}
}

func (s *LogicalAndContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLogicalAnd(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LogicalAnd() (localctx ILogicalAndContext) {
	localctx = NewLogicalAndContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 68, SPL2ParserRULE_logicalAnd)
	p.SetState(446)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(443)
			p.Match(SPL2ParserAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(444)

		if !(p.contextualKeyword("and")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"and\")", ""))
			goto errorExit
		}
		{
			p.SetState(445)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogicalOrContext is an interface to support dynamic dispatch.
type ILogicalOrContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OR() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsLogicalOrContext differentiates from other interfaces.
	IsLogicalOrContext()
}

type LogicalOrContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogicalOrContext() *LogicalOrContext {
	var p = new(LogicalOrContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalOr
	return p
}

func InitEmptyLogicalOrContext(p *LogicalOrContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalOr
}

func (*LogicalOrContext) IsLogicalOrContext() {}

func NewLogicalOrContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LogicalOrContext {
	var p = new(LogicalOrContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_logicalOr

	return p
}

func (s *LogicalOrContext) GetParser() antlr.Parser { return s.parser }

func (s *LogicalOrContext) OR() antlr.TerminalNode {
	return s.GetToken(SPL2ParserOR, 0)
}

func (s *LogicalOrContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *LogicalOrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalOrContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LogicalOrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLogicalOr(s)
	}
}

func (s *LogicalOrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLogicalOr(s)
	}
}

func (s *LogicalOrContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLogicalOr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LogicalOr() (localctx ILogicalOrContext) {
	localctx = NewLogicalOrContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 70, SPL2ParserRULE_logicalOr)
	p.SetState(451)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 46, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(448)
			p.Match(SPL2ParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(449)

		if !(p.contextualKeyword("or")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"or\")", ""))
			goto errorExit
		}
		{
			p.SetState(450)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogicalXorContext is an interface to support dynamic dispatch.
type ILogicalXorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	XOR() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsLogicalXorContext differentiates from other interfaces.
	IsLogicalXorContext()
}

type LogicalXorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogicalXorContext() *LogicalXorContext {
	var p = new(LogicalXorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalXor
	return p
}

func InitEmptyLogicalXorContext(p *LogicalXorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalXor
}

func (*LogicalXorContext) IsLogicalXorContext() {}

func NewLogicalXorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LogicalXorContext {
	var p = new(LogicalXorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_logicalXor

	return p
}

func (s *LogicalXorContext) GetParser() antlr.Parser { return s.parser }

func (s *LogicalXorContext) XOR() antlr.TerminalNode {
	return s.GetToken(SPL2ParserXOR, 0)
}

func (s *LogicalXorContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *LogicalXorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalXorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LogicalXorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLogicalXor(s)
	}
}

func (s *LogicalXorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLogicalXor(s)
	}
}

func (s *LogicalXorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLogicalXor(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LogicalXor() (localctx ILogicalXorContext) {
	localctx = NewLogicalXorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 72, SPL2ParserRULE_logicalXor)
	p.SetState(456)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 47, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(453)
			p.Match(SPL2ParserXOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(454)

		if !(p.contextualKeyword("xor")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"xor\")", ""))
			goto errorExit
		}
		{
			p.SetState(455)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILogicalNotContext is an interface to support dynamic dispatch.
type ILogicalNotContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NOT() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsLogicalNotContext differentiates from other interfaces.
	IsLogicalNotContext()
}

type LogicalNotContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLogicalNotContext() *LogicalNotContext {
	var p = new(LogicalNotContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalNot
	return p
}

func InitEmptyLogicalNotContext(p *LogicalNotContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_logicalNot
}

func (*LogicalNotContext) IsLogicalNotContext() {}

func NewLogicalNotContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LogicalNotContext {
	var p = new(LogicalNotContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_logicalNot

	return p
}

func (s *LogicalNotContext) GetParser() antlr.Parser { return s.parser }

func (s *LogicalNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNOT, 0)
}

func (s *LogicalNotContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *LogicalNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalNotContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LogicalNotContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLogicalNot(s)
	}
}

func (s *LogicalNotContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLogicalNot(s)
	}
}

func (s *LogicalNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLogicalNot(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LogicalNot() (localctx ILogicalNotContext) {
	localctx = NewLogicalNotContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 74, SPL2ParserRULE_logicalNot)
	p.SetState(461)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 48, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(458)
			p.Match(SPL2ParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(459)

		if !(p.contextualKeyword("not")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"not\")", ""))
			goto errorExit
		}
		{
			p.SetState(460)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBetweenOperatorContext is an interface to support dynamic dispatch.
type IBetweenOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BETWEEN() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsBetweenOperatorContext differentiates from other interfaces.
	IsBetweenOperatorContext()
}

type BetweenOperatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBetweenOperatorContext() *BetweenOperatorContext {
	var p = new(BetweenOperatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_betweenOperator
	return p
}

func InitEmptyBetweenOperatorContext(p *BetweenOperatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_betweenOperator
}

func (*BetweenOperatorContext) IsBetweenOperatorContext() {}

func NewBetweenOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BetweenOperatorContext {
	var p = new(BetweenOperatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_betweenOperator

	return p
}

func (s *BetweenOperatorContext) GetParser() antlr.Parser { return s.parser }

func (s *BetweenOperatorContext) BETWEEN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserBETWEEN, 0)
}

func (s *BetweenOperatorContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *BetweenOperatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BetweenOperatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BetweenOperatorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterBetweenOperator(s)
	}
}

func (s *BetweenOperatorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitBetweenOperator(s)
	}
}

func (s *BetweenOperatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitBetweenOperator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) BetweenOperator() (localctx IBetweenOperatorContext) {
	localctx = NewBetweenOperatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 76, SPL2ParserRULE_betweenOperator)
	p.SetState(466)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 49, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(463)
			p.Match(SPL2ParserBETWEEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(464)

		if !(p.contextualKeyword("between")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"between\")", ""))
			goto errorExit
		}
		{
			p.SetState(465)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBetweenConjunctionContext is an interface to support dynamic dispatch.
type IBetweenConjunctionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AND() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsBetweenConjunctionContext differentiates from other interfaces.
	IsBetweenConjunctionContext()
}

type BetweenConjunctionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBetweenConjunctionContext() *BetweenConjunctionContext {
	var p = new(BetweenConjunctionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_betweenConjunction
	return p
}

func InitEmptyBetweenConjunctionContext(p *BetweenConjunctionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_betweenConjunction
}

func (*BetweenConjunctionContext) IsBetweenConjunctionContext() {}

func NewBetweenConjunctionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BetweenConjunctionContext {
	var p = new(BetweenConjunctionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_betweenConjunction

	return p
}

func (s *BetweenConjunctionContext) GetParser() antlr.Parser { return s.parser }

func (s *BetweenConjunctionContext) AND() antlr.TerminalNode {
	return s.GetToken(SPL2ParserAND, 0)
}

func (s *BetweenConjunctionContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *BetweenConjunctionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BetweenConjunctionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BetweenConjunctionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterBetweenConjunction(s)
	}
}

func (s *BetweenConjunctionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitBetweenConjunction(s)
	}
}

func (s *BetweenConjunctionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitBetweenConjunction(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) BetweenConjunction() (localctx IBetweenConjunctionContext) {
	localctx = NewBetweenConjunctionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 78, SPL2ParserRULE_betweenConjunction)
	p.SetState(471)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 50, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(468)
			p.Match(SPL2ParserAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		p.SetState(469)

		if !(p.contextualKeyword("and")) {
			p.SetError(antlr.NewFailedPredicateException(p, "p.contextualKeyword(\"and\")", ""))
			goto errorExit
		}
		{
			p.SetState(470)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IComparisonContext is an interface to support dynamic dispatch.
type IComparisonContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ASSIGN() antlr.TerminalNode
	EQ() antlr.TerminalNode
	NE() antlr.TerminalNode
	LT() antlr.TerminalNode
	LE() antlr.TerminalNode
	GT() antlr.TerminalNode
	GE() antlr.TerminalNode

	// IsComparisonContext differentiates from other interfaces.
	IsComparisonContext()
}

type ComparisonContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyComparisonContext() *ComparisonContext {
	var p = new(ComparisonContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_comparison
	return p
}

func InitEmptyComparisonContext(p *ComparisonContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_comparison
}

func (*ComparisonContext) IsComparisonContext() {}

func NewComparisonContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ComparisonContext {
	var p = new(ComparisonContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_comparison

	return p
}

func (s *ComparisonContext) GetParser() antlr.Parser { return s.parser }

func (s *ComparisonContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *ComparisonContext) EQ() antlr.TerminalNode {
	return s.GetToken(SPL2ParserEQ, 0)
}

func (s *ComparisonContext) NE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNE, 0)
}

func (s *ComparisonContext) LT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLT, 0)
}

func (s *ComparisonContext) LE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLE, 0)
}

func (s *ComparisonContext) GT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserGT, 0)
}

func (s *ComparisonContext) GE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserGE, 0)
}

func (s *ComparisonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ComparisonContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterComparison(s)
	}
}

func (s *ComparisonContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitComparison(s)
	}
}

func (s *ComparisonContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitComparison(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Comparison() (localctx IComparisonContext) {
	localctx = NewComparisonContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 80, SPL2ParserRULE_comparison)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(473)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272730423296) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAdditiveContext is an interface to support dynamic dispatch.
type IAdditiveContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMultiplicative() []IMultiplicativeContext
	Multiplicative(i int) IMultiplicativeContext
	AllPLUS() []antlr.TerminalNode
	PLUS(i int) antlr.TerminalNode
	AllMINUS() []antlr.TerminalNode
	MINUS(i int) antlr.TerminalNode

	// IsAdditiveContext differentiates from other interfaces.
	IsAdditiveContext()
}

type AdditiveContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAdditiveContext() *AdditiveContext {
	var p = new(AdditiveContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_additive
	return p
}

func InitEmptyAdditiveContext(p *AdditiveContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_additive
}

func (*AdditiveContext) IsAdditiveContext() {}

func NewAdditiveContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AdditiveContext {
	var p = new(AdditiveContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_additive

	return p
}

func (s *AdditiveContext) GetParser() antlr.Parser { return s.parser }

func (s *AdditiveContext) AllMultiplicative() []IMultiplicativeContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMultiplicativeContext); ok {
			len++
		}
	}

	tst := make([]IMultiplicativeContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMultiplicativeContext); ok {
			tst[i] = t.(IMultiplicativeContext)
			i++
		}
	}

	return tst
}

func (s *AdditiveContext) Multiplicative(i int) IMultiplicativeContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMultiplicativeContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMultiplicativeContext)
}

func (s *AdditiveContext) AllPLUS() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserPLUS)
}

func (s *AdditiveContext) PLUS(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserPLUS, i)
}

func (s *AdditiveContext) AllMINUS() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserMINUS)
}

func (s *AdditiveContext) MINUS(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserMINUS, i)
}

func (s *AdditiveContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AdditiveContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AdditiveContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterAdditive(s)
	}
}

func (s *AdditiveContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitAdditive(s)
	}
}

func (s *AdditiveContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitAdditive(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Additive() (localctx IAdditiveContext) {
	localctx = NewAdditiveContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 82, SPL2ParserRULE_additive)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(475)
		p.Multiplicative()
	}
	p.SetState(480)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 51, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(476)
				_la = p.GetTokenStream().LA(1)

				if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(477)
				p.Multiplicative()
			}

		}
		p.SetState(482)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 51, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMultiplicativeContext is an interface to support dynamic dispatch.
type IMultiplicativeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllUnary() []IUnaryContext
	Unary(i int) IUnaryContext
	AllSTAR() []antlr.TerminalNode
	STAR(i int) antlr.TerminalNode
	AllSLASH() []antlr.TerminalNode
	SLASH(i int) antlr.TerminalNode
	AllMOD() []antlr.TerminalNode
	MOD(i int) antlr.TerminalNode

	// IsMultiplicativeContext differentiates from other interfaces.
	IsMultiplicativeContext()
}

type MultiplicativeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiplicativeContext() *MultiplicativeContext {
	var p = new(MultiplicativeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_multiplicative
	return p
}

func InitEmptyMultiplicativeContext(p *MultiplicativeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_multiplicative
}

func (*MultiplicativeContext) IsMultiplicativeContext() {}

func NewMultiplicativeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiplicativeContext {
	var p = new(MultiplicativeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_multiplicative

	return p
}

func (s *MultiplicativeContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiplicativeContext) AllUnary() []IUnaryContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUnaryContext); ok {
			len++
		}
	}

	tst := make([]IUnaryContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUnaryContext); ok {
			tst[i] = t.(IUnaryContext)
			i++
		}
	}

	return tst
}

func (s *MultiplicativeContext) Unary(i int) IUnaryContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnaryContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnaryContext)
}

func (s *MultiplicativeContext) AllSTAR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSTAR)
}

func (s *MultiplicativeContext) STAR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTAR, i)
}

func (s *MultiplicativeContext) AllSLASH() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSLASH)
}

func (s *MultiplicativeContext) SLASH(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSLASH, i)
}

func (s *MultiplicativeContext) AllMOD() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserMOD)
}

func (s *MultiplicativeContext) MOD(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserMOD, i)
}

func (s *MultiplicativeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiplicativeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiplicativeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterMultiplicative(s)
	}
}

func (s *MultiplicativeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitMultiplicative(s)
	}
}

func (s *MultiplicativeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitMultiplicative(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Multiplicative() (localctx IMultiplicativeContext) {
	localctx = NewMultiplicativeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 84, SPL2ParserRULE_multiplicative)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(483)
		p.Unary()
	}
	p.SetState(488)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 52, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(484)
				_la = p.GetTokenStream().LA(1)

				if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&27487790694400) != 0) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(485)
				p.Unary()
			}

		}
		p.SetState(490)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 52, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnaryContext is an interface to support dynamic dispatch.
type IUnaryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Unary() IUnaryContext
	PLUS() antlr.TerminalNode
	MINUS() antlr.TerminalNode
	Access() IAccessContext

	// IsUnaryContext differentiates from other interfaces.
	IsUnaryContext()
}

type UnaryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnaryContext() *UnaryContext {
	var p = new(UnaryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_unary
	return p
}

func InitEmptyUnaryContext(p *UnaryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_unary
}

func (*UnaryContext) IsUnaryContext() {}

func NewUnaryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnaryContext {
	var p = new(UnaryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_unary

	return p
}

func (s *UnaryContext) GetParser() antlr.Parser { return s.parser }

func (s *UnaryContext) Unary() IUnaryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnaryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnaryContext)
}

func (s *UnaryContext) PLUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserPLUS, 0)
}

func (s *UnaryContext) MINUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserMINUS, 0)
}

func (s *UnaryContext) Access() IAccessContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAccessContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAccessContext)
}

func (s *UnaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnaryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnaryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterUnary(s)
	}
}

func (s *UnaryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitUnary(s)
	}
}

func (s *UnaryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitUnary(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Unary() (localctx IUnaryContext) {
	localctx = NewUnaryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 86, SPL2ParserRULE_unary)
	var _la int

	p.SetState(494)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserPLUS, SPL2ParserMINUS:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(491)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(492)
			p.Unary()
		}

	case SPL2ParserINDEX, SPL2ParserNULL, SPL2ParserBOOLEAN, SPL2ParserLPAREN, SPL2ParserLBRACKET, SPL2ParserLBRACE, SPL2ParserRAW_STRING, SPL2ParserDQUOTE, SPL2ParserSQUOTE, SPL2ParserBACKTICK, SPL2ParserNUMBER, SPL2ParserLOCAL, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(493)
			p.Access()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAccessContext is an interface to support dynamic dispatch.
type IAccessContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Primary() IPrimaryContext
	AllAccessPart() []IAccessPartContext
	AccessPart(i int) IAccessPartContext

	// IsAccessContext differentiates from other interfaces.
	IsAccessContext()
}

type AccessContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAccessContext() *AccessContext {
	var p = new(AccessContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_access
	return p
}

func InitEmptyAccessContext(p *AccessContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_access
}

func (*AccessContext) IsAccessContext() {}

func NewAccessContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AccessContext {
	var p = new(AccessContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_access

	return p
}

func (s *AccessContext) GetParser() antlr.Parser { return s.parser }

func (s *AccessContext) Primary() IPrimaryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimaryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimaryContext)
}

func (s *AccessContext) AllAccessPart() []IAccessPartContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAccessPartContext); ok {
			len++
		}
	}

	tst := make([]IAccessPartContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAccessPartContext); ok {
			tst[i] = t.(IAccessPartContext)
			i++
		}
	}

	return tst
}

func (s *AccessContext) AccessPart(i int) IAccessPartContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAccessPartContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAccessPartContext)
}

func (s *AccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccessContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AccessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterAccess(s)
	}
}

func (s *AccessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitAccess(s)
	}
}

func (s *AccessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitAccess(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Access() (localctx IAccessContext) {
	localctx = NewAccessContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 88, SPL2ParserRULE_access)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(496)
		p.Primary()
	}
	p.SetState(500)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 54, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(497)
				p.AccessPart()
			}

		}
		p.SetState(502)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 54, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAccessPartContext is an interface to support dynamic dispatch.
type IAccessPartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DOT() antlr.TerminalNode
	Identifier() IIdentifierContext
	LBRACKET() antlr.TerminalNode
	Expression() IExpressionContext
	RBRACKET() antlr.TerminalNode

	// IsAccessPartContext differentiates from other interfaces.
	IsAccessPartContext()
}

type AccessPartContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAccessPartContext() *AccessPartContext {
	var p = new(AccessPartContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_accessPart
	return p
}

func InitEmptyAccessPartContext(p *AccessPartContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_accessPart
}

func (*AccessPartContext) IsAccessPartContext() {}

func NewAccessPartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AccessPartContext {
	var p = new(AccessPartContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_accessPart

	return p
}

func (s *AccessPartContext) GetParser() antlr.Parser { return s.parser }

func (s *AccessPartContext) DOT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserDOT, 0)
}

func (s *AccessPartContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *AccessPartContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLBRACKET, 0)
}

func (s *AccessPartContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AccessPartContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACKET, 0)
}

func (s *AccessPartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AccessPartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AccessPartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterAccessPart(s)
	}
}

func (s *AccessPartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitAccessPart(s)
	}
}

func (s *AccessPartContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitAccessPart(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) AccessPart() (localctx IAccessPartContext) {
	localctx = NewAccessPartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 90, SPL2ParserRULE_accessPart)
	p.SetState(509)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserDOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(503)
			p.Match(SPL2ParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(504)
			p.Identifier()
		}

	case SPL2ParserLBRACKET:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(505)
			p.Match(SPL2ParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(506)
			p.Expression()
		}
		{
			p.SetState(507)
			p.Match(SPL2ParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPrimaryContext is an interface to support dynamic dispatch.
type IPrimaryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Call() ICallContext
	FieldName() IFieldNameContext
	LOCAL() antlr.TerminalNode
	Literal() ILiteralContext
	Array() IArrayContext
	Object() IObjectContext
	LPAREN() antlr.TerminalNode
	Expression() IExpressionContext
	RPAREN() antlr.TerminalNode
	SearchLiteral() ISearchLiteralContext

	// IsPrimaryContext differentiates from other interfaces.
	IsPrimaryContext()
}

type PrimaryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPrimaryContext() *PrimaryContext {
	var p = new(PrimaryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_primary
	return p
}

func InitEmptyPrimaryContext(p *PrimaryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_primary
}

func (*PrimaryContext) IsPrimaryContext() {}

func NewPrimaryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PrimaryContext {
	var p = new(PrimaryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_primary

	return p
}

func (s *PrimaryContext) GetParser() antlr.Parser { return s.parser }

func (s *PrimaryContext) Call() ICallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICallContext)
}

func (s *PrimaryContext) FieldName() IFieldNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldNameContext)
}

func (s *PrimaryContext) LOCAL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLOCAL, 0)
}

func (s *PrimaryContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *PrimaryContext) Array() IArrayContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayContext)
}

func (s *PrimaryContext) Object() IObjectContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IObjectContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IObjectContext)
}

func (s *PrimaryContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLPAREN, 0)
}

func (s *PrimaryContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PrimaryContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRPAREN, 0)
}

func (s *PrimaryContext) SearchLiteral() ISearchLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchLiteralContext)
}

func (s *PrimaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrimaryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PrimaryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterPrimary(s)
	}
}

func (s *PrimaryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitPrimary(s)
	}
}

func (s *PrimaryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitPrimary(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Primary() (localctx IPrimaryContext) {
	localctx = NewPrimaryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 92, SPL2ParserRULE_primary)
	p.SetState(522)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 56, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(511)
			p.Call()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(512)
			p.FieldName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(513)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(514)
			p.Literal()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(515)
			p.Array()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(516)
			p.Object()
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(517)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(518)
			p.Expression()
		}
		{
			p.SetState(519)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(521)
			p.SearchLiteral()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICallContext is an interface to support dynamic dispatch.
type ICallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	Arguments() IArgumentsContext

	// IsCallContext differentiates from other interfaces.
	IsCallContext()
}

type CallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCallContext() *CallContext {
	var p = new(CallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_call
	return p
}

func InitEmptyCallContext(p *CallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_call
}

func (*CallContext) IsCallContext() {}

func NewCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CallContext {
	var p = new(CallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_call

	return p
}

func (s *CallContext) GetParser() antlr.Parser { return s.parser }

func (s *CallContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *CallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLPAREN, 0)
}

func (s *CallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRPAREN, 0)
}

func (s *CallContext) Arguments() IArgumentsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgumentsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgumentsContext)
}

func (s *CallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterCall(s)
	}
}

func (s *CallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitCall(s)
	}
}

func (s *CallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Call() (localctx ICallContext) {
	localctx = NewCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 94, SPL2ParserRULE_call)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(524)
		p.Identifier()
	}
	{
		p.SetState(525)
		p.Match(SPL2ParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(527)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 57, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(526)
			p.Arguments()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	{
		p.SetState(529)
		p.Match(SPL2ParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArgumentsContext is an interface to support dynamic dispatch.
type IArgumentsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNamedArgument() []INamedArgumentContext
	NamedArgument(i int) INamedArgumentContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext

	// IsArgumentsContext differentiates from other interfaces.
	IsArgumentsContext()
}

type ArgumentsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgumentsContext() *ArgumentsContext {
	var p = new(ArgumentsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_arguments
	return p
}

func InitEmptyArgumentsContext(p *ArgumentsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_arguments
}

func (*ArgumentsContext) IsArgumentsContext() {}

func NewArgumentsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgumentsContext {
	var p = new(ArgumentsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_arguments

	return p
}

func (s *ArgumentsContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgumentsContext) AllNamedArgument() []INamedArgumentContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INamedArgumentContext); ok {
			len++
		}
	}

	tst := make([]INamedArgumentContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INamedArgumentContext); ok {
			tst[i] = t.(INamedArgumentContext)
			i++
		}
	}

	return tst
}

func (s *ArgumentsContext) NamedArgument(i int) INamedArgumentContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INamedArgumentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(INamedArgumentContext)
}

func (s *ArgumentsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *ArgumentsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *ArgumentsContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ArgumentsContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ArgumentsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgumentsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgumentsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterArguments(s)
	}
}

func (s *ArgumentsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitArguments(s)
	}
}

func (s *ArgumentsContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitArguments(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Arguments() (localctx IArgumentsContext) {
	localctx = NewArgumentsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 96, SPL2ParserRULE_arguments)
	var _la int

	var _alt int

	p.SetState(554)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 61, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(531)
			p.NamedArgument()
		}
		p.SetState(536)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(532)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(533)
				p.NamedArgument()
			}

			p.SetState(538)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(539)
			p.Expression()
		}
		p.SetState(544)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 59, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(540)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(541)
					p.Expression()
				}

			}
			p.SetState(546)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 59, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(551)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(547)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(548)
				p.NamedArgument()
			}

			p.SetState(553)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INamedArgumentContext is an interface to support dynamic dispatch.
type INamedArgumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	COLON() antlr.TerminalNode
	Expression() IExpressionContext

	// IsNamedArgumentContext differentiates from other interfaces.
	IsNamedArgumentContext()
}

type NamedArgumentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNamedArgumentContext() *NamedArgumentContext {
	var p = new(NamedArgumentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_namedArgument
	return p
}

func InitEmptyNamedArgumentContext(p *NamedArgumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_namedArgument
}

func (*NamedArgumentContext) IsNamedArgumentContext() {}

func NewNamedArgumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NamedArgumentContext {
	var p = new(NamedArgumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_namedArgument

	return p
}

func (s *NamedArgumentContext) GetParser() antlr.Parser { return s.parser }

func (s *NamedArgumentContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *NamedArgumentContext) COLON() antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOLON, 0)
}

func (s *NamedArgumentContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *NamedArgumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NamedArgumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NamedArgumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterNamedArgument(s)
	}
}

func (s *NamedArgumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitNamedArgument(s)
	}
}

func (s *NamedArgumentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitNamedArgument(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) NamedArgument() (localctx INamedArgumentContext) {
	localctx = NewNamedArgumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 98, SPL2ParserRULE_namedArgument)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(556)
		p.Identifier()
	}
	{
		p.SetState(557)
		p.Match(SPL2ParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(558)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMBER() antlr.TerminalNode
	BOOLEAN() antlr.TerminalNode
	NULL() antlr.TerminalNode
	RAW_STRING() antlr.TerminalNode
	StringLiteral() IStringLiteralContext

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNUMBER, 0)
}

func (s *LiteralContext) BOOLEAN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserBOOLEAN, 0)
}

func (s *LiteralContext) NULL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNULL, 0)
}

func (s *LiteralContext) RAW_STRING() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRAW_STRING, 0)
}

func (s *LiteralContext) StringLiteral() IStringLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStringLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStringLiteralContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLiteral(s)
	}
}

func (s *LiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLiteral(s)
	}
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 100, SPL2ParserRULE_literal)
	p.SetState(565)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserNUMBER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(560)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserBOOLEAN:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(561)
			p.Match(SPL2ParserBOOLEAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserNULL:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(562)
			p.Match(SPL2ParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserRAW_STRING:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(563)
			p.Match(SPL2ParserRAW_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(564)
			p.StringLiteral()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStringLiteralContext is an interface to support dynamic dispatch.
type IStringLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DQUOTE() antlr.TerminalNode
	STRING_END() antlr.TerminalNode
	AllSTRING_TEXT() []antlr.TerminalNode
	STRING_TEXT(i int) antlr.TerminalNode
	AllSTRING_DOLLAR() []antlr.TerminalNode
	STRING_DOLLAR(i int) antlr.TerminalNode
	AllSTRING_INTERPOLATION() []antlr.TerminalNode
	STRING_INTERPOLATION(i int) antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllRBRACE() []antlr.TerminalNode
	RBRACE(i int) antlr.TerminalNode

	// IsStringLiteralContext differentiates from other interfaces.
	IsStringLiteralContext()
}

type StringLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStringLiteralContext() *StringLiteralContext {
	var p = new(StringLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_stringLiteral
	return p
}

func InitEmptyStringLiteralContext(p *StringLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_stringLiteral
}

func (*StringLiteralContext) IsStringLiteralContext() {}

func NewStringLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StringLiteralContext {
	var p = new(StringLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_stringLiteral

	return p
}

func (s *StringLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *StringLiteralContext) DQUOTE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserDQUOTE, 0)
}

func (s *StringLiteralContext) STRING_END() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTRING_END, 0)
}

func (s *StringLiteralContext) AllSTRING_TEXT() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSTRING_TEXT)
}

func (s *StringLiteralContext) STRING_TEXT(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTRING_TEXT, i)
}

func (s *StringLiteralContext) AllSTRING_DOLLAR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSTRING_DOLLAR)
}

func (s *StringLiteralContext) STRING_DOLLAR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTRING_DOLLAR, i)
}

func (s *StringLiteralContext) AllSTRING_INTERPOLATION() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSTRING_INTERPOLATION)
}

func (s *StringLiteralContext) STRING_INTERPOLATION(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSTRING_INTERPOLATION, i)
}

func (s *StringLiteralContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *StringLiteralContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *StringLiteralContext) AllRBRACE() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserRBRACE)
}

func (s *StringLiteralContext) RBRACE(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACE, i)
}

func (s *StringLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StringLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StringLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterStringLiteral(s)
	}
}

func (s *StringLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitStringLiteral(s)
	}
}

func (s *StringLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitStringLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) StringLiteral() (localctx IStringLiteralContext) {
	localctx = NewStringLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 102, SPL2ParserRULE_stringLiteral)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(567)
		p.Match(SPL2ParserDQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(576)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-66)) & ^0x3f) == 0 && ((int64(1)<<(_la-66))&7) != 0 {
		p.SetState(574)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserSTRING_TEXT:
			{
				p.SetState(568)
				p.Match(SPL2ParserSTRING_TEXT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserSTRING_DOLLAR:
			{
				p.SetState(569)
				p.Match(SPL2ParserSTRING_DOLLAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserSTRING_INTERPOLATION:
			{
				p.SetState(570)
				p.Match(SPL2ParserSTRING_INTERPOLATION)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(571)
				p.Expression()
			}
			{
				p.SetState(572)
				p.Match(SPL2ParserRBRACE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(578)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(579)
		p.Match(SPL2ParserSTRING_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQuotedNameContext is an interface to support dynamic dispatch.
type IQuotedNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SQUOTE() antlr.TerminalNode
	NAME_END() antlr.TerminalNode
	AllNAME_TEXT() []antlr.TerminalNode
	NAME_TEXT(i int) antlr.TerminalNode
	AllNAME_DOLLAR() []antlr.TerminalNode
	NAME_DOLLAR(i int) antlr.TerminalNode

	// IsQuotedNameContext differentiates from other interfaces.
	IsQuotedNameContext()
}

type QuotedNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQuotedNameContext() *QuotedNameContext {
	var p = new(QuotedNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_quotedName
	return p
}

func InitEmptyQuotedNameContext(p *QuotedNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_quotedName
}

func (*QuotedNameContext) IsQuotedNameContext() {}

func NewQuotedNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QuotedNameContext {
	var p = new(QuotedNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_quotedName

	return p
}

func (s *QuotedNameContext) GetParser() antlr.Parser { return s.parser }

func (s *QuotedNameContext) SQUOTE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSQUOTE, 0)
}

func (s *QuotedNameContext) NAME_END() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_END, 0)
}

func (s *QuotedNameContext) AllNAME_TEXT() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNAME_TEXT)
}

func (s *QuotedNameContext) NAME_TEXT(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_TEXT, i)
}

func (s *QuotedNameContext) AllNAME_DOLLAR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNAME_DOLLAR)
}

func (s *QuotedNameContext) NAME_DOLLAR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_DOLLAR, i)
}

func (s *QuotedNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QuotedNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QuotedNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterQuotedName(s)
	}
}

func (s *QuotedNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitQuotedName(s)
	}
}

func (s *QuotedNameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitQuotedName(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) QuotedName() (localctx IQuotedNameContext) {
	localctx = NewQuotedNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 104, SPL2ParserRULE_quotedName)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(581)
		p.Match(SPL2ParserSQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(583)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR {
		{
			p.SetState(582)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(585)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(587)
		p.Match(SPL2ParserNAME_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldTemplateContext is an interface to support dynamic dispatch.
type IFieldTemplateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SQUOTE() antlr.TerminalNode
	AllNAME_INTERPOLATION() []antlr.TerminalNode
	NAME_INTERPOLATION(i int) antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllRBRACE() []antlr.TerminalNode
	RBRACE(i int) antlr.TerminalNode
	NAME_END() antlr.TerminalNode
	AllNAME_TEXT() []antlr.TerminalNode
	NAME_TEXT(i int) antlr.TerminalNode
	AllNAME_DOLLAR() []antlr.TerminalNode
	NAME_DOLLAR(i int) antlr.TerminalNode

	// IsFieldTemplateContext differentiates from other interfaces.
	IsFieldTemplateContext()
}

type FieldTemplateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldTemplateContext() *FieldTemplateContext {
	var p = new(FieldTemplateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldTemplate
	return p
}

func InitEmptyFieldTemplateContext(p *FieldTemplateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldTemplate
}

func (*FieldTemplateContext) IsFieldTemplateContext() {}

func NewFieldTemplateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldTemplateContext {
	var p = new(FieldTemplateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_fieldTemplate

	return p
}

func (s *FieldTemplateContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldTemplateContext) SQUOTE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserSQUOTE, 0)
}

func (s *FieldTemplateContext) AllNAME_INTERPOLATION() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNAME_INTERPOLATION)
}

func (s *FieldTemplateContext) NAME_INTERPOLATION(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_INTERPOLATION, i)
}

func (s *FieldTemplateContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *FieldTemplateContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *FieldTemplateContext) AllRBRACE() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserRBRACE)
}

func (s *FieldTemplateContext) RBRACE(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACE, i)
}

func (s *FieldTemplateContext) NAME_END() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_END, 0)
}

func (s *FieldTemplateContext) AllNAME_TEXT() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNAME_TEXT)
}

func (s *FieldTemplateContext) NAME_TEXT(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_TEXT, i)
}

func (s *FieldTemplateContext) AllNAME_DOLLAR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNAME_DOLLAR)
}

func (s *FieldTemplateContext) NAME_DOLLAR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNAME_DOLLAR, i)
}

func (s *FieldTemplateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldTemplateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldTemplateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterFieldTemplate(s)
	}
}

func (s *FieldTemplateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitFieldTemplate(s)
	}
}

func (s *FieldTemplateContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitFieldTemplate(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) FieldTemplate() (localctx IFieldTemplateContext) {
	localctx = NewFieldTemplateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 106, SPL2ParserRULE_fieldTemplate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(589)
		p.Match(SPL2ParserSQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(593)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR {
		{
			p.SetState(590)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(595)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(596)
		p.Match(SPL2ParserNAME_INTERPOLATION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(597)
		p.Expression()
	}
	{
		p.SetState(598)
		p.Match(SPL2ParserRBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(607)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-70)) & ^0x3f) == 0 && ((int64(1)<<(_la-70))&7) != 0 {
		p.SetState(605)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserNAME_TEXT:
			{
				p.SetState(599)
				p.Match(SPL2ParserNAME_TEXT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserNAME_DOLLAR:
			{
				p.SetState(600)
				p.Match(SPL2ParserNAME_DOLLAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserNAME_INTERPOLATION:
			{
				p.SetState(601)
				p.Match(SPL2ParserNAME_INTERPOLATION)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(602)
				p.Expression()
			}
			{
				p.SetState(603)
				p.Match(SPL2ParserRBRACE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(609)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(610)
		p.Match(SPL2ParserNAME_END)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldNameContext is an interface to support dynamic dispatch.
type IFieldNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	FieldTemplate() IFieldTemplateContext

	// IsFieldNameContext differentiates from other interfaces.
	IsFieldNameContext()
}

type FieldNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldNameContext() *FieldNameContext {
	var p = new(FieldNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldName
	return p
}

func InitEmptyFieldNameContext(p *FieldNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_fieldName
}

func (*FieldNameContext) IsFieldNameContext() {}

func NewFieldNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldNameContext {
	var p = new(FieldNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_fieldName

	return p
}

func (s *FieldNameContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldNameContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *FieldNameContext) FieldTemplate() IFieldTemplateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldTemplateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldTemplateContext)
}

func (s *FieldNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterFieldName(s)
	}
}

func (s *FieldNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitFieldName(s)
	}
}

func (s *FieldNameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitFieldName(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) FieldName() (localctx IFieldNameContext) {
	localctx = NewFieldNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 108, SPL2ParserRULE_fieldName)
	p.SetState(614)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 69, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(612)
			p.Identifier()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(613)
			p.FieldTemplate()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentifierContext is an interface to support dynamic dispatch.
type IIdentifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	INDEX() antlr.TerminalNode
	QuotedName() IQuotedNameContext

	// IsIdentifierContext differentiates from other interfaces.
	IsIdentifierContext()
}

type IdentifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifierContext() *IdentifierContext {
	var p = new(IdentifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_identifier
	return p
}

func InitEmptyIdentifierContext(p *IdentifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_identifier
}

func (*IdentifierContext) IsIdentifierContext() {}

func NewIdentifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentifierContext {
	var p = new(IdentifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_identifier

	return p
}

func (s *IdentifierContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentifierContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserIDENTIFIER, 0)
}

func (s *IdentifierContext) INDEX() antlr.TerminalNode {
	return s.GetToken(SPL2ParserINDEX, 0)
}

func (s *IdentifierContext) QuotedName() IQuotedNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQuotedNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQuotedNameContext)
}

func (s *IdentifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentifierContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterIdentifier(s)
	}
}

func (s *IdentifierContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitIdentifier(s)
	}
}

func (s *IdentifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitIdentifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Identifier() (localctx IIdentifierContext) {
	localctx = NewIdentifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 110, SPL2ParserRULE_identifier)
	p.SetState(619)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(616)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserINDEX:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(617)
			p.Match(SPL2ParserINDEX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserSQUOTE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(618)
			p.QuotedName()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayContext is an interface to support dynamic dispatch.
type IArrayContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsArrayContext differentiates from other interfaces.
	IsArrayContext()
}

type ArrayContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayContext() *ArrayContext {
	var p = new(ArrayContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_array
	return p
}

func InitEmptyArrayContext(p *ArrayContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_array
}

func (*ArrayContext) IsArrayContext() {}

func NewArrayContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayContext {
	var p = new(ArrayContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_array

	return p
}

func (s *ArrayContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLBRACKET, 0)
}

func (s *ArrayContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACKET, 0)
}

func (s *ArrayContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ArrayContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ArrayContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *ArrayContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *ArrayContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterArray(s)
	}
}

func (s *ArrayContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitArray(s)
	}
}

func (s *ArrayContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitArray(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Array() (localctx IArrayContext) {
	localctx = NewArrayContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 112, SPL2ParserRULE_array)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(621)
		p.Match(SPL2ParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(633)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 73, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(622)
			p.Expression()
		}
		p.SetState(627)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 71, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(623)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(624)
					p.Expression()
				}

			}
			p.SetState(629)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 71, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(631)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserCOMMA {
			{
				p.SetState(630)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	{
		p.SetState(635)
		p.Match(SPL2ParserRBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IObjectContext is an interface to support dynamic dispatch.
type IObjectContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllObjectEntry() []IObjectEntryContext
	ObjectEntry(i int) IObjectEntryContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsObjectContext differentiates from other interfaces.
	IsObjectContext()
}

type ObjectContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyObjectContext() *ObjectContext {
	var p = new(ObjectContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_object
	return p
}

func InitEmptyObjectContext(p *ObjectContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_object
}

func (*ObjectContext) IsObjectContext() {}

func NewObjectContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ObjectContext {
	var p = new(ObjectContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_object

	return p
}

func (s *ObjectContext) GetParser() antlr.Parser { return s.parser }

func (s *ObjectContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLBRACE, 0)
}

func (s *ObjectContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACE, 0)
}

func (s *ObjectContext) AllObjectEntry() []IObjectEntryContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IObjectEntryContext); ok {
			len++
		}
	}

	tst := make([]IObjectEntryContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IObjectEntryContext); ok {
			tst[i] = t.(IObjectEntryContext)
			i++
		}
	}

	return tst
}

func (s *ObjectContext) ObjectEntry(i int) IObjectEntryContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IObjectEntryContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IObjectEntryContext)
}

func (s *ObjectContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *ObjectContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *ObjectContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ObjectContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ObjectContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterObject(s)
	}
}

func (s *ObjectContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitObject(s)
	}
}

func (s *ObjectContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitObject(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) Object() (localctx IObjectContext) {
	localctx = NewObjectContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 114, SPL2ParserRULE_object)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(637)
		p.Match(SPL2ParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(649)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&5044031582654955536) != 0 {
		{
			p.SetState(638)
			p.ObjectEntry()
		}
		p.SetState(643)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 74, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(639)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(640)
					p.ObjectEntry()
				}

			}
			p.SetState(645)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 74, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(647)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserCOMMA {
			{
				p.SetState(646)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	}
	{
		p.SetState(651)
		p.Match(SPL2ParserRBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IObjectEntryContext is an interface to support dynamic dispatch.
type IObjectEntryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ObjectKey() IObjectKeyContext
	COLON() antlr.TerminalNode
	Expression() IExpressionContext

	// IsObjectEntryContext differentiates from other interfaces.
	IsObjectEntryContext()
}

type ObjectEntryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyObjectEntryContext() *ObjectEntryContext {
	var p = new(ObjectEntryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_objectEntry
	return p
}

func InitEmptyObjectEntryContext(p *ObjectEntryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_objectEntry
}

func (*ObjectEntryContext) IsObjectEntryContext() {}

func NewObjectEntryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ObjectEntryContext {
	var p = new(ObjectEntryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_objectEntry

	return p
}

func (s *ObjectEntryContext) GetParser() antlr.Parser { return s.parser }

func (s *ObjectEntryContext) ObjectKey() IObjectKeyContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IObjectKeyContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IObjectKeyContext)
}

func (s *ObjectEntryContext) COLON() antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOLON, 0)
}

func (s *ObjectEntryContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ObjectEntryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ObjectEntryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ObjectEntryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterObjectEntry(s)
	}
}

func (s *ObjectEntryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitObjectEntry(s)
	}
}

func (s *ObjectEntryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitObjectEntry(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) ObjectEntry() (localctx IObjectEntryContext) {
	localctx = NewObjectEntryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 116, SPL2ParserRULE_objectEntry)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(653)
		p.ObjectKey()
	}
	{
		p.SetState(654)
		p.Match(SPL2ParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(655)
		p.Expression()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IObjectKeyContext is an interface to support dynamic dispatch.
type IObjectKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	StringLiteral() IStringLiteralContext

	// IsObjectKeyContext differentiates from other interfaces.
	IsObjectKeyContext()
}

type ObjectKeyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyObjectKeyContext() *ObjectKeyContext {
	var p = new(ObjectKeyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_objectKey
	return p
}

func InitEmptyObjectKeyContext(p *ObjectKeyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_objectKey
}

func (*ObjectKeyContext) IsObjectKeyContext() {}

func NewObjectKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ObjectKeyContext {
	var p = new(ObjectKeyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_objectKey

	return p
}

func (s *ObjectKeyContext) GetParser() antlr.Parser { return s.parser }

func (s *ObjectKeyContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ObjectKeyContext) StringLiteral() IStringLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStringLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStringLiteralContext)
}

func (s *ObjectKeyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ObjectKeyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ObjectKeyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterObjectKey(s)
	}
}

func (s *ObjectKeyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitObjectKey(s)
	}
}

func (s *ObjectKeyContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitObjectKey(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) ObjectKey() (localctx IObjectKeyContext) {
	localctx = NewObjectKeyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 118, SPL2ParserRULE_objectKey)
	p.SetState(659)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(657)
			p.Identifier()
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(658)
			p.StringLiteral()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISearchLiteralContext is an interface to support dynamic dispatch.
type ISearchLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EmbeddedText() IEmbeddedTextContext

	// IsSearchLiteralContext differentiates from other interfaces.
	IsSearchLiteralContext()
}

type SearchLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchLiteralContext() *SearchLiteralContext {
	var p = new(SearchLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchLiteral
	return p
}

func InitEmptySearchLiteralContext(p *SearchLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_searchLiteral
}

func (*SearchLiteralContext) IsSearchLiteralContext() {}

func NewSearchLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchLiteralContext {
	var p = new(SearchLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_searchLiteral

	return p
}

func (s *SearchLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchLiteralContext) EmbeddedText() IEmbeddedTextContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEmbeddedTextContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEmbeddedTextContext)
}

func (s *SearchLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterSearchLiteral(s)
	}
}

func (s *SearchLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitSearchLiteral(s)
	}
}

func (s *SearchLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitSearchLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) SearchLiteral() (localctx ISearchLiteralContext) {
	localctx = NewSearchLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 120, SPL2ParserRULE_searchLiteral)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(661)
		p.EmbeddedText()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILambdaExpressionContext is an interface to support dynamic dispatch.
type ILambdaExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ARROW() antlr.TerminalNode
	AllLambdaParameter() []ILambdaParameterContext
	LambdaParameter(i int) ILambdaParameterContext
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	Expression() IExpressionContext
	LambdaBlock() ILambdaBlockContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsLambdaExpressionContext differentiates from other interfaces.
	IsLambdaExpressionContext()
}

type LambdaExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLambdaExpressionContext() *LambdaExpressionContext {
	var p = new(LambdaExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaExpression
	return p
}

func InitEmptyLambdaExpressionContext(p *LambdaExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaExpression
}

func (*LambdaExpressionContext) IsLambdaExpressionContext() {}

func NewLambdaExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LambdaExpressionContext {
	var p = new(LambdaExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_lambdaExpression

	return p
}

func (s *LambdaExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *LambdaExpressionContext) ARROW() antlr.TerminalNode {
	return s.GetToken(SPL2ParserARROW, 0)
}

func (s *LambdaExpressionContext) AllLambdaParameter() []ILambdaParameterContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILambdaParameterContext); ok {
			len++
		}
	}

	tst := make([]ILambdaParameterContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILambdaParameterContext); ok {
			tst[i] = t.(ILambdaParameterContext)
			i++
		}
	}

	return tst
}

func (s *LambdaExpressionContext) LambdaParameter(i int) ILambdaParameterContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILambdaParameterContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILambdaParameterContext)
}

func (s *LambdaExpressionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLPAREN, 0)
}

func (s *LambdaExpressionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRPAREN, 0)
}

func (s *LambdaExpressionContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *LambdaExpressionContext) LambdaBlock() ILambdaBlockContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILambdaBlockContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILambdaBlockContext)
}

func (s *LambdaExpressionContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserCOMMA)
}

func (s *LambdaExpressionContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOMMA, i)
}

func (s *LambdaExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LambdaExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LambdaExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLambdaExpression(s)
	}
}

func (s *LambdaExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLambdaExpression(s)
	}
}

func (s *LambdaExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLambdaExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LambdaExpression() (localctx ILambdaExpressionContext) {
	localctx = NewLambdaExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 122, SPL2ParserRULE_lambdaExpression)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(676)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserLOCAL:
		{
			p.SetState(663)
			p.LambdaParameter()
		}

	case SPL2ParserLPAREN:
		{
			p.SetState(664)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(673)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserLOCAL {
			{
				p.SetState(665)
				p.LambdaParameter()
			}
			p.SetState(670)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserCOMMA {
				{
					p.SetState(666)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(667)
					p.LambdaParameter()
				}

				p.SetState(672)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(675)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(678)
		p.Match(SPL2ParserARROW)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(681)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 81, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(679)
			p.Expression()
		}

	case 2:
		{
			p.SetState(680)
			p.LambdaBlock()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILambdaParameterContext is an interface to support dynamic dispatch.
type ILambdaParameterContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LOCAL() antlr.TerminalNode
	COLON() antlr.TerminalNode
	TYPE() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	NUMBER() antlr.TerminalNode
	StringLiteral() IStringLiteralContext
	PLUS() antlr.TerminalNode
	MINUS() antlr.TerminalNode

	// IsLambdaParameterContext differentiates from other interfaces.
	IsLambdaParameterContext()
}

type LambdaParameterContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLambdaParameterContext() *LambdaParameterContext {
	var p = new(LambdaParameterContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaParameter
	return p
}

func InitEmptyLambdaParameterContext(p *LambdaParameterContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaParameter
}

func (*LambdaParameterContext) IsLambdaParameterContext() {}

func NewLambdaParameterContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LambdaParameterContext {
	var p = new(LambdaParameterContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_lambdaParameter

	return p
}

func (s *LambdaParameterContext) GetParser() antlr.Parser { return s.parser }

func (s *LambdaParameterContext) LOCAL() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLOCAL, 0)
}

func (s *LambdaParameterContext) COLON() antlr.TerminalNode {
	return s.GetToken(SPL2ParserCOLON, 0)
}

func (s *LambdaParameterContext) TYPE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserTYPE, 0)
}

func (s *LambdaParameterContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *LambdaParameterContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNUMBER, 0)
}

func (s *LambdaParameterContext) StringLiteral() IStringLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStringLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStringLiteralContext)
}

func (s *LambdaParameterContext) PLUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserPLUS, 0)
}

func (s *LambdaParameterContext) MINUS() antlr.TerminalNode {
	return s.GetToken(SPL2ParserMINUS, 0)
}

func (s *LambdaParameterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LambdaParameterContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LambdaParameterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLambdaParameter(s)
	}
}

func (s *LambdaParameterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLambdaParameter(s)
	}
}

func (s *LambdaParameterContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLambdaParameter(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LambdaParameter() (localctx ILambdaParameterContext) {
	localctx = NewLambdaParameterContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 124, SPL2ParserRULE_lambdaParameter)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(683)
		p.Match(SPL2ParserLOCAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(686)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserCOLON {
		{
			p.SetState(684)
			p.Match(SPL2ParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(685)
			p.Match(SPL2ParserTYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(696)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserASSIGN {
		{
			p.SetState(688)
			p.Match(SPL2ParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(694)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserPLUS, SPL2ParserMINUS, SPL2ParserNUMBER:
			p.SetState(690)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == SPL2ParserPLUS || _la == SPL2ParserMINUS {
				{
					p.SetState(689)
					_la = p.GetTokenStream().LA(1)

					if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}

			}
			{
				p.SetState(692)
				p.Match(SPL2ParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserDQUOTE:
			{
				p.SetState(693)
				p.StringLiteral()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILambdaBlockContext is an interface to support dynamic dispatch.
type ILambdaBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACE() antlr.TerminalNode
	RETURN() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	RBRACE() antlr.TerminalNode
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode
	AllLOCAL() []antlr.TerminalNode
	LOCAL(i int) antlr.TerminalNode
	AllASSIGN() []antlr.TerminalNode
	ASSIGN(i int) antlr.TerminalNode
	AllSEMI() []antlr.TerminalNode
	SEMI(i int) antlr.TerminalNode

	// IsLambdaBlockContext differentiates from other interfaces.
	IsLambdaBlockContext()
}

type LambdaBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLambdaBlockContext() *LambdaBlockContext {
	var p = new(LambdaBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaBlock
	return p
}

func InitEmptyLambdaBlockContext(p *LambdaBlockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPL2ParserRULE_lambdaBlock
}

func (*LambdaBlockContext) IsLambdaBlockContext() {}

func NewLambdaBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LambdaBlockContext {
	var p = new(LambdaBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPL2ParserRULE_lambdaBlock

	return p
}

func (s *LambdaBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *LambdaBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserLBRACE, 0)
}

func (s *LambdaBlockContext) RETURN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRETURN, 0)
}

func (s *LambdaBlockContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *LambdaBlockContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *LambdaBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(SPL2ParserRBRACE, 0)
}

func (s *LambdaBlockContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserNL)
}

func (s *LambdaBlockContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserNL, i)
}

func (s *LambdaBlockContext) AllLOCAL() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserLOCAL)
}

func (s *LambdaBlockContext) LOCAL(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserLOCAL, i)
}

func (s *LambdaBlockContext) AllASSIGN() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserASSIGN)
}

func (s *LambdaBlockContext) ASSIGN(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, i)
}

func (s *LambdaBlockContext) AllSEMI() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserSEMI)
}

func (s *LambdaBlockContext) SEMI(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserSEMI, i)
}

func (s *LambdaBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LambdaBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LambdaBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.EnterLambdaBlock(s)
	}
}

func (s *LambdaBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPL2ParserListener); ok {
		listenerT.ExitLambdaBlock(s)
	}
}

func (s *LambdaBlockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPL2ParserVisitor:
		return t.VisitLambdaBlock(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPL2Parser) LambdaBlock() (localctx ILambdaBlockContext) {
	localctx = NewLambdaBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 126, SPL2ParserRULE_lambdaBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(698)
		p.Match(SPL2ParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(702)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(699)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(704)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(724)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserLOCAL {
		{
			p.SetState(705)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(706)
			p.Match(SPL2ParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(707)
			p.Expression()
		}
		p.SetState(720)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserSEMI:
			{
				p.SetState(708)
				p.Match(SPL2ParserSEMI)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(712)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(709)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(714)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		case SPL2ParserNL:
			p.SetState(716)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for ok := true; ok; ok = _la == SPL2ParserNL {
				{
					p.SetState(715)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(718)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(726)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(727)
		p.Match(SPL2ParserRETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(728)
		p.Expression()
	}
	p.SetState(730)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserSEMI {
		{
			p.SetState(729)
			p.Match(SPL2ParserSEMI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(735)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(732)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(737)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(738)
		p.Match(SPL2ParserRBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *SPL2Parser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 20:
		var t *ImplicitSearchContext = nil
		if localctx != nil {
			t = localctx.(*ImplicitSearchContext)
		}
		return p.ImplicitSearch_Sempred(t, predIndex)

	case 34:
		var t *LogicalAndContext = nil
		if localctx != nil {
			t = localctx.(*LogicalAndContext)
		}
		return p.LogicalAnd_Sempred(t, predIndex)

	case 35:
		var t *LogicalOrContext = nil
		if localctx != nil {
			t = localctx.(*LogicalOrContext)
		}
		return p.LogicalOr_Sempred(t, predIndex)

	case 36:
		var t *LogicalXorContext = nil
		if localctx != nil {
			t = localctx.(*LogicalXorContext)
		}
		return p.LogicalXor_Sempred(t, predIndex)

	case 37:
		var t *LogicalNotContext = nil
		if localctx != nil {
			t = localctx.(*LogicalNotContext)
		}
		return p.LogicalNot_Sempred(t, predIndex)

	case 38:
		var t *BetweenOperatorContext = nil
		if localctx != nil {
			t = localctx.(*BetweenOperatorContext)
		}
		return p.BetweenOperator_Sempred(t, predIndex)

	case 39:
		var t *BetweenConjunctionContext = nil
		if localctx != nil {
			t = localctx.(*BetweenConjunctionContext)
		}
		return p.BetweenConjunction_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *SPL2Parser) ImplicitSearch_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.GetTokenStream().LA(1) == SPL2ParserINDEX && p.GetTokenStream().LA(2) == SPL2ParserASSIGN

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) LogicalAnd_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 1:
		return p.contextualKeyword("and")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) LogicalOr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 2:
		return p.contextualKeyword("or")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) LogicalXor_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 3:
		return p.contextualKeyword("xor")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) LogicalNot_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 4:
		return p.contextualKeyword("not")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) BetweenOperator_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 5:
		return p.contextualKeyword("between")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPL2Parser) BetweenConjunction_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 6:
		return p.contextualKeyword("and")

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
