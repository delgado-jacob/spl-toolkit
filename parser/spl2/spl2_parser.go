// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

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
		"'function'", "'return'", "'AND'", "'OR'", "'XOR'", "'NOT'", "", "'IN'",
		"'LIKE'", "'IS'", "'null'", "'NULL'", "", "", "'->'", "'<='", "'>='",
		"'!='", "'=='", "", "'<'", "'>'", "'+'", "'-'", "'*'", "", "", "'/'",
		"'%'", "", "','", "':'", "'.'", "'('", "')'", "'['", "']'", "'{'", "'}'",
		"';'",
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
		"andExpression", "notExpression", "predicate", "comparison", "additive",
		"multiplicative", "unary", "access", "accessPart", "primary", "call",
		"arguments", "namedArgument", "literal", "stringLiteral", "quotedName",
		"fieldTemplate", "fieldName", "identifier", "array", "object", "objectEntry",
		"objectKey", "searchLiteral", "lambdaExpression", "lambdaParameter",
		"lambdaBlock",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 76, 698, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
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
		1, 0, 5, 0, 118, 8, 0, 10, 0, 12, 0, 121, 9, 0, 1, 0, 1, 0, 3, 0, 125,
		8, 0, 1, 0, 3, 0, 128, 8, 0, 1, 0, 5, 0, 131, 8, 0, 10, 0, 12, 0, 134,
		9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 5, 1, 140, 8, 1, 10, 1, 12, 1, 143, 9, 1,
		1, 1, 1, 1, 5, 1, 147, 8, 1, 10, 1, 12, 1, 150, 9, 1, 1, 1, 5, 1, 153,
		8, 1, 10, 1, 12, 1, 156, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2,
		164, 8, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 174, 8,
		3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 5, 5, 183, 8, 5, 10, 5, 12,
		5, 186, 9, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 3, 6, 194, 8, 6, 1, 7,
		1, 7, 3, 7, 198, 8, 7, 1, 8, 1, 8, 3, 8, 202, 8, 8, 1, 9, 1, 9, 1, 9, 1,
		9, 5, 9, 208, 8, 9, 10, 9, 12, 9, 211, 9, 9, 1, 10, 1, 10, 1, 10, 1, 10,
		1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 3, 12, 222, 8, 12, 1, 12, 1, 12, 1,
		12, 5, 12, 227, 8, 12, 10, 12, 12, 12, 230, 9, 12, 1, 13, 1, 13, 5, 13,
		234, 8, 13, 10, 13, 12, 13, 237, 9, 13, 1, 13, 1, 13, 1, 13, 3, 13, 242,
		8, 13, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 248, 8, 14, 1, 15, 3, 15, 251,
		8, 15, 1, 15, 1, 15, 1, 16, 1, 16, 3, 16, 257, 8, 16, 1, 16, 1, 16, 1,
		17, 1, 17, 5, 17, 263, 8, 17, 10, 17, 12, 17, 266, 9, 17, 1, 18, 1, 18,
		1, 18, 1, 18, 1, 18, 3, 18, 273, 8, 18, 1, 18, 5, 18, 276, 8, 18, 10, 18,
		12, 18, 279, 9, 18, 1, 19, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20, 1, 20, 3,
		20, 288, 8, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 5, 22, 295, 8, 22, 10,
		22, 12, 22, 298, 9, 22, 1, 23, 1, 23, 3, 23, 302, 8, 23, 1, 23, 5, 23,
		305, 8, 23, 10, 23, 12, 23, 308, 9, 23, 1, 24, 1, 24, 1, 24, 5, 24, 313,
		8, 24, 10, 24, 12, 24, 316, 9, 24, 1, 25, 1, 25, 1, 25, 3, 25, 321, 8,
		25, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 4, 26, 337, 8, 26, 11, 26, 12, 26, 338, 1,
		26, 1, 26, 1, 26, 3, 26, 344, 8, 26, 1, 27, 1, 27, 1, 27, 1, 27, 1, 27,
		3, 27, 351, 8, 27, 1, 28, 1, 28, 3, 28, 355, 8, 28, 1, 29, 1, 29, 1, 29,
		5, 29, 360, 8, 29, 10, 29, 12, 29, 363, 9, 29, 1, 30, 1, 30, 1, 30, 5,
		30, 368, 8, 30, 10, 30, 12, 30, 371, 9, 30, 1, 31, 1, 31, 1, 31, 5, 31,
		376, 8, 31, 10, 31, 12, 31, 379, 9, 31, 1, 32, 1, 32, 1, 32, 3, 32, 384,
		8, 32, 1, 33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 391, 8, 33, 1, 33, 1,
		33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 399, 8, 33, 1, 33, 1, 33, 1, 33,
		1, 33, 1, 33, 5, 33, 406, 8, 33, 10, 33, 12, 33, 409, 9, 33, 1, 33, 1,
		33, 1, 33, 3, 33, 414, 8, 33, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 420, 8,
		33, 1, 33, 1, 33, 3, 33, 424, 8, 33, 1, 33, 3, 33, 427, 8, 33, 3, 33, 429,
		8, 33, 1, 34, 1, 34, 1, 35, 1, 35, 1, 35, 5, 35, 436, 8, 35, 10, 35, 12,
		35, 439, 9, 35, 1, 36, 1, 36, 1, 36, 5, 36, 444, 8, 36, 10, 36, 12, 36,
		447, 9, 36, 1, 37, 1, 37, 1, 37, 3, 37, 452, 8, 37, 1, 38, 1, 38, 5, 38,
		456, 8, 38, 10, 38, 12, 38, 459, 9, 38, 1, 39, 1, 39, 1, 39, 1, 39, 1,
		39, 1, 39, 3, 39, 467, 8, 39, 1, 40, 1, 40, 1, 40, 1, 40, 1, 40, 1, 40,
		1, 40, 1, 40, 1, 40, 1, 40, 1, 40, 3, 40, 480, 8, 40, 1, 41, 1, 41, 1,
		41, 3, 41, 485, 8, 41, 1, 41, 1, 41, 1, 42, 1, 42, 1, 42, 5, 42, 492, 8,
		42, 10, 42, 12, 42, 495, 9, 42, 1, 42, 1, 42, 1, 42, 5, 42, 500, 8, 42,
		10, 42, 12, 42, 503, 9, 42, 1, 42, 1, 42, 5, 42, 507, 8, 42, 10, 42, 12,
		42, 510, 9, 42, 3, 42, 512, 8, 42, 1, 43, 1, 43, 1, 43, 1, 43, 1, 44, 1,
		44, 1, 44, 1, 44, 1, 44, 3, 44, 523, 8, 44, 1, 45, 1, 45, 1, 45, 1, 45,
		1, 45, 1, 45, 1, 45, 5, 45, 532, 8, 45, 10, 45, 12, 45, 535, 9, 45, 1,
		45, 1, 45, 1, 46, 1, 46, 4, 46, 541, 8, 46, 11, 46, 12, 46, 542, 1, 46,
		1, 46, 1, 47, 1, 47, 5, 47, 549, 8, 47, 10, 47, 12, 47, 552, 9, 47, 1,
		47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 5, 47, 563,
		8, 47, 10, 47, 12, 47, 566, 9, 47, 1, 47, 1, 47, 1, 48, 1, 48, 3, 48, 572,
		8, 48, 1, 49, 1, 49, 1, 49, 3, 49, 577, 8, 49, 1, 50, 1, 50, 1, 50, 1,
		50, 5, 50, 583, 8, 50, 10, 50, 12, 50, 586, 9, 50, 1, 50, 3, 50, 589, 8,
		50, 3, 50, 591, 8, 50, 1, 50, 1, 50, 1, 51, 1, 51, 1, 51, 1, 51, 5, 51,
		599, 8, 51, 10, 51, 12, 51, 602, 9, 51, 1, 51, 3, 51, 605, 8, 51, 3, 51,
		607, 8, 51, 1, 51, 1, 51, 1, 52, 1, 52, 1, 52, 1, 52, 1, 53, 1, 53, 3,
		53, 617, 8, 53, 1, 54, 1, 54, 1, 55, 1, 55, 1, 55, 1, 55, 1, 55, 5, 55,
		626, 8, 55, 10, 55, 12, 55, 629, 9, 55, 3, 55, 631, 8, 55, 1, 55, 3, 55,
		634, 8, 55, 1, 55, 1, 55, 1, 55, 3, 55, 639, 8, 55, 1, 56, 1, 56, 1, 56,
		3, 56, 644, 8, 56, 1, 56, 1, 56, 3, 56, 648, 8, 56, 1, 56, 1, 56, 3, 56,
		652, 8, 56, 3, 56, 654, 8, 56, 1, 57, 1, 57, 5, 57, 658, 8, 57, 10, 57,
		12, 57, 661, 9, 57, 1, 57, 1, 57, 1, 57, 1, 57, 1, 57, 5, 57, 668, 8, 57,
		10, 57, 12, 57, 671, 9, 57, 1, 57, 4, 57, 674, 8, 57, 11, 57, 12, 57, 675,
		3, 57, 678, 8, 57, 5, 57, 680, 8, 57, 10, 57, 12, 57, 683, 9, 57, 1, 57,
		1, 57, 1, 57, 3, 57, 688, 8, 57, 1, 57, 5, 57, 691, 8, 57, 10, 57, 12,
		57, 694, 9, 57, 1, 57, 1, 57, 1, 57, 2, 264, 277, 0, 58, 0, 2, 4, 6, 8,
		10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44,
		46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80,
		82, 84, 86, 88, 90, 92, 94, 96, 98, 100, 102, 104, 106, 108, 110, 112,
		114, 0, 6, 1, 0, 7, 8, 1, 0, 38, 39, 1, 0, 26, 27, 1, 0, 31, 37, 2, 0,
		40, 40, 43, 44, 1, 0, 71, 72, 761, 0, 119, 1, 0, 0, 0, 2, 137, 1, 0, 0,
		0, 4, 163, 1, 0, 0, 0, 6, 173, 1, 0, 0, 0, 8, 175, 1, 0, 0, 0, 10, 178,
		1, 0, 0, 0, 12, 190, 1, 0, 0, 0, 14, 197, 1, 0, 0, 0, 16, 199, 1, 0, 0,
		0, 18, 203, 1, 0, 0, 0, 20, 212, 1, 0, 0, 0, 22, 216, 1, 0, 0, 0, 24, 219,
		1, 0, 0, 0, 26, 231, 1, 0, 0, 0, 28, 243, 1, 0, 0, 0, 30, 250, 1, 0, 0,
		0, 32, 254, 1, 0, 0, 0, 34, 260, 1, 0, 0, 0, 36, 272, 1, 0, 0, 0, 38, 280,
		1, 0, 0, 0, 40, 283, 1, 0, 0, 0, 42, 289, 1, 0, 0, 0, 44, 291, 1, 0, 0,
		0, 46, 299, 1, 0, 0, 0, 48, 309, 1, 0, 0, 0, 50, 320, 1, 0, 0, 0, 52, 343,
		1, 0, 0, 0, 54, 350, 1, 0, 0, 0, 56, 354, 1, 0, 0, 0, 58, 356, 1, 0, 0,
		0, 60, 364, 1, 0, 0, 0, 62, 372, 1, 0, 0, 0, 64, 383, 1, 0, 0, 0, 66, 385,
		1, 0, 0, 0, 68, 430, 1, 0, 0, 0, 70, 432, 1, 0, 0, 0, 72, 440, 1, 0, 0,
		0, 74, 451, 1, 0, 0, 0, 76, 453, 1, 0, 0, 0, 78, 466, 1, 0, 0, 0, 80, 479,
		1, 0, 0, 0, 82, 481, 1, 0, 0, 0, 84, 511, 1, 0, 0, 0, 86, 513, 1, 0, 0,
		0, 88, 522, 1, 0, 0, 0, 90, 524, 1, 0, 0, 0, 92, 538, 1, 0, 0, 0, 94, 546,
		1, 0, 0, 0, 96, 571, 1, 0, 0, 0, 98, 576, 1, 0, 0, 0, 100, 578, 1, 0, 0,
		0, 102, 594, 1, 0, 0, 0, 104, 610, 1, 0, 0, 0, 106, 616, 1, 0, 0, 0, 108,
		618, 1, 0, 0, 0, 110, 633, 1, 0, 0, 0, 112, 640, 1, 0, 0, 0, 114, 655,
		1, 0, 0, 0, 116, 118, 5, 63, 0, 0, 117, 116, 1, 0, 0, 0, 118, 121, 1, 0,
		0, 0, 119, 117, 1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 127, 1, 0, 0, 0,
		121, 119, 1, 0, 0, 0, 122, 124, 3, 2, 1, 0, 123, 125, 3, 34, 17, 0, 124,
		123, 1, 0, 0, 0, 124, 125, 1, 0, 0, 0, 125, 128, 1, 0, 0, 0, 126, 128,
		3, 36, 18, 0, 127, 122, 1, 0, 0, 0, 127, 126, 1, 0, 0, 0, 128, 132, 1,
		0, 0, 0, 129, 131, 5, 63, 0, 0, 130, 129, 1, 0, 0, 0, 131, 134, 1, 0, 0,
		0, 132, 130, 1, 0, 0, 0, 132, 133, 1, 0, 0, 0, 133, 135, 1, 0, 0, 0, 134,
		132, 1, 0, 0, 0, 135, 136, 5, 0, 0, 1, 136, 1, 1, 0, 0, 0, 137, 154, 3,
		4, 2, 0, 138, 140, 5, 63, 0, 0, 139, 138, 1, 0, 0, 0, 140, 143, 1, 0, 0,
		0, 141, 139, 1, 0, 0, 0, 141, 142, 1, 0, 0, 0, 142, 144, 1, 0, 0, 0, 143,
		141, 1, 0, 0, 0, 144, 148, 5, 45, 0, 0, 145, 147, 5, 63, 0, 0, 146, 145,
		1, 0, 0, 0, 147, 150, 1, 0, 0, 0, 148, 146, 1, 0, 0, 0, 148, 149, 1, 0,
		0, 0, 149, 151, 1, 0, 0, 0, 150, 148, 1, 0, 0, 0, 151, 153, 3, 6, 3, 0,
		152, 141, 1, 0, 0, 0, 153, 156, 1, 0, 0, 0, 154, 152, 1, 0, 0, 0, 154,
		155, 1, 0, 0, 0, 155, 3, 1, 0, 0, 0, 156, 154, 1, 0, 0, 0, 157, 164, 3,
		8, 4, 0, 158, 164, 3, 10, 5, 0, 159, 164, 3, 38, 19, 0, 160, 164, 3, 40,
		20, 0, 161, 164, 3, 16, 8, 0, 162, 164, 3, 30, 15, 0, 163, 157, 1, 0, 0,
		0, 163, 158, 1, 0, 0, 0, 163, 159, 1, 0, 0, 0, 163, 160, 1, 0, 0, 0, 163,
		161, 1, 0, 0, 0, 163, 162, 1, 0, 0, 0, 164, 5, 1, 0, 0, 0, 165, 174, 3,
		18, 9, 0, 166, 174, 3, 22, 11, 0, 167, 174, 3, 24, 12, 0, 168, 174, 3,
		8, 4, 0, 169, 174, 3, 10, 5, 0, 170, 174, 3, 38, 19, 0, 171, 174, 3, 26,
		13, 0, 172, 174, 3, 30, 15, 0, 173, 165, 1, 0, 0, 0, 173, 166, 1, 0, 0,
		0, 173, 167, 1, 0, 0, 0, 173, 168, 1, 0, 0, 0, 173, 169, 1, 0, 0, 0, 173,
		170, 1, 0, 0, 0, 173, 171, 1, 0, 0, 0, 173, 172, 1, 0, 0, 0, 174, 7, 1,
		0, 0, 0, 175, 176, 5, 1, 0, 0, 176, 177, 3, 14, 7, 0, 177, 9, 1, 0, 0,
		0, 178, 179, 5, 2, 0, 0, 179, 184, 3, 12, 6, 0, 180, 181, 5, 46, 0, 0,
		181, 183, 3, 12, 6, 0, 182, 180, 1, 0, 0, 0, 183, 186, 1, 0, 0, 0, 184,
		182, 1, 0, 0, 0, 184, 185, 1, 0, 0, 0, 185, 187, 1, 0, 0, 0, 186, 184,
		1, 0, 0, 0, 187, 188, 5, 1, 0, 0, 188, 189, 3, 14, 7, 0, 189, 11, 1, 0,
		0, 0, 190, 193, 3, 56, 28, 0, 191, 192, 5, 9, 0, 0, 192, 194, 3, 98, 49,
		0, 193, 191, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 13, 1, 0, 0, 0, 195,
		198, 3, 98, 49, 0, 196, 198, 3, 100, 50, 0, 197, 195, 1, 0, 0, 0, 197,
		196, 1, 0, 0, 0, 198, 15, 1, 0, 0, 0, 199, 201, 5, 12, 0, 0, 200, 202,
		5, 60, 0, 0, 201, 200, 1, 0, 0, 0, 201, 202, 1, 0, 0, 0, 202, 17, 1, 0,
		0, 0, 203, 204, 5, 5, 0, 0, 204, 209, 3, 20, 10, 0, 205, 206, 5, 46, 0,
		0, 206, 208, 3, 20, 10, 0, 207, 205, 1, 0, 0, 0, 208, 211, 1, 0, 0, 0,
		209, 207, 1, 0, 0, 0, 209, 210, 1, 0, 0, 0, 210, 19, 1, 0, 0, 0, 211, 209,
		1, 0, 0, 0, 212, 213, 3, 96, 48, 0, 213, 214, 5, 35, 0, 0, 214, 215, 3,
		56, 28, 0, 215, 21, 1, 0, 0, 0, 216, 217, 5, 6, 0, 0, 217, 218, 3, 56,
		28, 0, 218, 23, 1, 0, 0, 0, 219, 221, 7, 0, 0, 0, 220, 222, 7, 1, 0, 0,
		221, 220, 1, 0, 0, 0, 221, 222, 1, 0, 0, 0, 222, 223, 1, 0, 0, 0, 223,
		228, 3, 98, 49, 0, 224, 225, 5, 46, 0, 0, 225, 227, 3, 98, 49, 0, 226,
		224, 1, 0, 0, 0, 227, 230, 1, 0, 0, 0, 228, 226, 1, 0, 0, 0, 228, 229,
		1, 0, 0, 0, 229, 25, 1, 0, 0, 0, 230, 228, 1, 0, 0, 0, 231, 235, 5, 11,
		0, 0, 232, 234, 3, 28, 14, 0, 233, 232, 1, 0, 0, 0, 234, 237, 1, 0, 0,
		0, 235, 233, 1, 0, 0, 0, 235, 236, 1, 0, 0, 0, 236, 241, 1, 0, 0, 0, 237,
		235, 1, 0, 0, 0, 238, 242, 5, 75, 0, 0, 239, 242, 3, 90, 45, 0, 240, 242,
		5, 56, 0, 0, 241, 238, 1, 0, 0, 0, 241, 239, 1, 0, 0, 0, 241, 240, 1, 0,
		0, 0, 242, 27, 1, 0, 0, 0, 243, 244, 5, 62, 0, 0, 244, 247, 5, 35, 0, 0,
		245, 248, 3, 98, 49, 0, 246, 248, 5, 60, 0, 0, 247, 245, 1, 0, 0, 0, 247,
		246, 1, 0, 0, 0, 248, 29, 1, 0, 0, 0, 249, 251, 5, 13, 0, 0, 250, 249,
		1, 0, 0, 0, 250, 251, 1, 0, 0, 0, 251, 252, 1, 0, 0, 0, 252, 253, 3, 32,
		16, 0, 253, 31, 1, 0, 0, 0, 254, 256, 5, 59, 0, 0, 255, 257, 5, 74, 0,
		0, 256, 255, 1, 0, 0, 0, 256, 257, 1, 0, 0, 0, 257, 258, 1, 0, 0, 0, 258,
		259, 5, 73, 0, 0, 259, 33, 1, 0, 0, 0, 260, 264, 5, 55, 0, 0, 261, 263,
		9, 0, 0, 0, 262, 261, 1, 0, 0, 0, 263, 266, 1, 0, 0, 0, 264, 265, 1, 0,
		0, 0, 264, 262, 1, 0, 0, 0, 265, 35, 1, 0, 0, 0, 266, 264, 1, 0, 0, 0,
		267, 273, 5, 14, 0, 0, 268, 273, 5, 15, 0, 0, 269, 273, 5, 16, 0, 0, 270,
		271, 5, 61, 0, 0, 271, 273, 5, 35, 0, 0, 272, 267, 1, 0, 0, 0, 272, 268,
		1, 0, 0, 0, 272, 269, 1, 0, 0, 0, 272, 270, 1, 0, 0, 0, 273, 277, 1, 0,
		0, 0, 274, 276, 9, 0, 0, 0, 275, 274, 1, 0, 0, 0, 276, 279, 1, 0, 0, 0,
		277, 278, 1, 0, 0, 0, 277, 275, 1, 0, 0, 0, 278, 37, 1, 0, 0, 0, 279, 277,
		1, 0, 0, 0, 280, 281, 5, 3, 0, 0, 281, 282, 3, 42, 21, 0, 282, 39, 1, 0,
		0, 0, 283, 284, 5, 4, 0, 0, 284, 285, 5, 35, 0, 0, 285, 287, 3, 54, 27,
		0, 286, 288, 3, 42, 21, 0, 287, 286, 1, 0, 0, 0, 287, 288, 1, 0, 0, 0,
		288, 41, 1, 0, 0, 0, 289, 290, 3, 44, 22, 0, 290, 43, 1, 0, 0, 0, 291,
		296, 3, 46, 23, 0, 292, 293, 5, 20, 0, 0, 293, 295, 3, 46, 23, 0, 294,
		292, 1, 0, 0, 0, 295, 298, 1, 0, 0, 0, 296, 294, 1, 0, 0, 0, 296, 297,
		1, 0, 0, 0, 297, 45, 1, 0, 0, 0, 298, 296, 1, 0, 0, 0, 299, 306, 3, 48,
		24, 0, 300, 302, 5, 18, 0, 0, 301, 300, 1, 0, 0, 0, 301, 302, 1, 0, 0,
		0, 302, 303, 1, 0, 0, 0, 303, 305, 3, 48, 24, 0, 304, 301, 1, 0, 0, 0,
		305, 308, 1, 0, 0, 0, 306, 304, 1, 0, 0, 0, 306, 307, 1, 0, 0, 0, 307,
		47, 1, 0, 0, 0, 308, 306, 1, 0, 0, 0, 309, 314, 3, 50, 25, 0, 310, 311,
		5, 19, 0, 0, 311, 313, 3, 50, 25, 0, 312, 310, 1, 0, 0, 0, 313, 316, 1,
		0, 0, 0, 314, 312, 1, 0, 0, 0, 314, 315, 1, 0, 0, 0, 315, 49, 1, 0, 0,
		0, 316, 314, 1, 0, 0, 0, 317, 318, 5, 21, 0, 0, 318, 321, 3, 50, 25, 0,
		319, 321, 3, 52, 26, 0, 320, 317, 1, 0, 0, 0, 320, 319, 1, 0, 0, 0, 321,
		51, 1, 0, 0, 0, 322, 323, 5, 49, 0, 0, 323, 324, 3, 42, 21, 0, 324, 325,
		5, 50, 0, 0, 325, 344, 1, 0, 0, 0, 326, 327, 3, 98, 49, 0, 327, 328, 3,
		68, 34, 0, 328, 329, 3, 54, 27, 0, 329, 344, 1, 0, 0, 0, 330, 331, 3, 98,
		49, 0, 331, 332, 5, 23, 0, 0, 332, 333, 5, 49, 0, 0, 333, 336, 3, 54, 27,
		0, 334, 335, 5, 46, 0, 0, 335, 337, 3, 54, 27, 0, 336, 334, 1, 0, 0, 0,
		337, 338, 1, 0, 0, 0, 338, 336, 1, 0, 0, 0, 338, 339, 1, 0, 0, 0, 339,
		340, 1, 0, 0, 0, 340, 341, 5, 50, 0, 0, 341, 344, 1, 0, 0, 0, 342, 344,
		3, 54, 27, 0, 343, 322, 1, 0, 0, 0, 343, 326, 1, 0, 0, 0, 343, 330, 1,
		0, 0, 0, 343, 342, 1, 0, 0, 0, 344, 53, 1, 0, 0, 0, 345, 351, 3, 98, 49,
		0, 346, 351, 5, 60, 0, 0, 347, 351, 3, 90, 45, 0, 348, 351, 5, 56, 0, 0,
		349, 351, 5, 40, 0, 0, 350, 345, 1, 0, 0, 0, 350, 346, 1, 0, 0, 0, 350,
		347, 1, 0, 0, 0, 350, 348, 1, 0, 0, 0, 350, 349, 1, 0, 0, 0, 351, 55, 1,
		0, 0, 0, 352, 355, 3, 110, 55, 0, 353, 355, 3, 58, 29, 0, 354, 352, 1,
		0, 0, 0, 354, 353, 1, 0, 0, 0, 355, 57, 1, 0, 0, 0, 356, 361, 3, 60, 30,
		0, 357, 358, 5, 20, 0, 0, 358, 360, 3, 60, 30, 0, 359, 357, 1, 0, 0, 0,
		360, 363, 1, 0, 0, 0, 361, 359, 1, 0, 0, 0, 361, 362, 1, 0, 0, 0, 362,
		59, 1, 0, 0, 0, 363, 361, 1, 0, 0, 0, 364, 369, 3, 62, 31, 0, 365, 366,
		5, 19, 0, 0, 366, 368, 3, 62, 31, 0, 367, 365, 1, 0, 0, 0, 368, 371, 1,
		0, 0, 0, 369, 367, 1, 0, 0, 0, 369, 370, 1, 0, 0, 0, 370, 61, 1, 0, 0,
		0, 371, 369, 1, 0, 0, 0, 372, 377, 3, 64, 32, 0, 373, 374, 5, 18, 0, 0,
		374, 376, 3, 64, 32, 0, 375, 373, 1, 0, 0, 0, 376, 379, 1, 0, 0, 0, 377,
		375, 1, 0, 0, 0, 377, 378, 1, 0, 0, 0, 378, 63, 1, 0, 0, 0, 379, 377, 1,
		0, 0, 0, 380, 381, 5, 21, 0, 0, 381, 384, 3, 64, 32, 0, 382, 384, 3, 66,
		33, 0, 383, 380, 1, 0, 0, 0, 383, 382, 1, 0, 0, 0, 384, 65, 1, 0, 0, 0,
		385, 428, 3, 70, 35, 0, 386, 387, 3, 68, 34, 0, 387, 388, 3, 70, 35, 0,
		388, 429, 1, 0, 0, 0, 389, 391, 5, 21, 0, 0, 390, 389, 1, 0, 0, 0, 390,
		391, 1, 0, 0, 0, 391, 392, 1, 0, 0, 0, 392, 393, 5, 22, 0, 0, 393, 394,
		3, 70, 35, 0, 394, 395, 5, 18, 0, 0, 395, 396, 3, 70, 35, 0, 396, 429,
		1, 0, 0, 0, 397, 399, 5, 21, 0, 0, 398, 397, 1, 0, 0, 0, 398, 399, 1, 0,
		0, 0, 399, 400, 1, 0, 0, 0, 400, 401, 5, 23, 0, 0, 401, 402, 5, 49, 0,
		0, 402, 407, 3, 56, 28, 0, 403, 404, 5, 46, 0, 0, 404, 406, 3, 56, 28,
		0, 405, 403, 1, 0, 0, 0, 406, 409, 1, 0, 0, 0, 407, 405, 1, 0, 0, 0, 407,
		408, 1, 0, 0, 0, 408, 410, 1, 0, 0, 0, 409, 407, 1, 0, 0, 0, 410, 411,
		5, 50, 0, 0, 411, 429, 1, 0, 0, 0, 412, 414, 5, 21, 0, 0, 413, 412, 1,
		0, 0, 0, 413, 414, 1, 0, 0, 0, 414, 415, 1, 0, 0, 0, 415, 416, 5, 24, 0,
		0, 416, 429, 3, 70, 35, 0, 417, 426, 5, 25, 0, 0, 418, 420, 5, 21, 0, 0,
		419, 418, 1, 0, 0, 0, 419, 420, 1, 0, 0, 0, 420, 421, 1, 0, 0, 0, 421,
		427, 7, 2, 0, 0, 422, 424, 5, 21, 0, 0, 423, 422, 1, 0, 0, 0, 423, 424,
		1, 0, 0, 0, 424, 425, 1, 0, 0, 0, 425, 427, 5, 29, 0, 0, 426, 419, 1, 0,
		0, 0, 426, 423, 1, 0, 0, 0, 427, 429, 1, 0, 0, 0, 428, 386, 1, 0, 0, 0,
		428, 390, 1, 0, 0, 0, 428, 398, 1, 0, 0, 0, 428, 413, 1, 0, 0, 0, 428,
		417, 1, 0, 0, 0, 428, 429, 1, 0, 0, 0, 429, 67, 1, 0, 0, 0, 430, 431, 7,
		3, 0, 0, 431, 69, 1, 0, 0, 0, 432, 437, 3, 72, 36, 0, 433, 434, 7, 1, 0,
		0, 434, 436, 3, 72, 36, 0, 435, 433, 1, 0, 0, 0, 436, 439, 1, 0, 0, 0,
		437, 435, 1, 0, 0, 0, 437, 438, 1, 0, 0, 0, 438, 71, 1, 0, 0, 0, 439, 437,
		1, 0, 0, 0, 440, 445, 3, 74, 37, 0, 441, 442, 7, 4, 0, 0, 442, 444, 3,
		74, 37, 0, 443, 441, 1, 0, 0, 0, 444, 447, 1, 0, 0, 0, 445, 443, 1, 0,
		0, 0, 445, 446, 1, 0, 0, 0, 446, 73, 1, 0, 0, 0, 447, 445, 1, 0, 0, 0,
		448, 449, 7, 1, 0, 0, 449, 452, 3, 74, 37, 0, 450, 452, 3, 76, 38, 0, 451,
		448, 1, 0, 0, 0, 451, 450, 1, 0, 0, 0, 452, 75, 1, 0, 0, 0, 453, 457, 3,
		80, 40, 0, 454, 456, 3, 78, 39, 0, 455, 454, 1, 0, 0, 0, 456, 459, 1, 0,
		0, 0, 457, 455, 1, 0, 0, 0, 457, 458, 1, 0, 0, 0, 458, 77, 1, 0, 0, 0,
		459, 457, 1, 0, 0, 0, 460, 461, 5, 48, 0, 0, 461, 467, 3, 98, 49, 0, 462,
		463, 5, 51, 0, 0, 463, 464, 3, 56, 28, 0, 464, 465, 5, 52, 0, 0, 465, 467,
		1, 0, 0, 0, 466, 460, 1, 0, 0, 0, 466, 462, 1, 0, 0, 0, 467, 79, 1, 0,
		0, 0, 468, 480, 3, 82, 41, 0, 469, 480, 3, 96, 48, 0, 470, 480, 5, 61,
		0, 0, 471, 480, 3, 88, 44, 0, 472, 480, 3, 100, 50, 0, 473, 480, 3, 102,
		51, 0, 474, 475, 5, 49, 0, 0, 475, 476, 3, 56, 28, 0, 476, 477, 5, 50,
		0, 0, 477, 480, 1, 0, 0, 0, 478, 480, 3, 108, 54, 0, 479, 468, 1, 0, 0,
		0, 479, 469, 1, 0, 0, 0, 479, 470, 1, 0, 0, 0, 479, 471, 1, 0, 0, 0, 479,
		472, 1, 0, 0, 0, 479, 473, 1, 0, 0, 0, 479, 474, 1, 0, 0, 0, 479, 478,
		1, 0, 0, 0, 480, 81, 1, 0, 0, 0, 481, 482, 3, 98, 49, 0, 482, 484, 5, 49,
		0, 0, 483, 485, 3, 84, 42, 0, 484, 483, 1, 0, 0, 0, 484, 485, 1, 0, 0,
		0, 485, 486, 1, 0, 0, 0, 486, 487, 5, 50, 0, 0, 487, 83, 1, 0, 0, 0, 488,
		493, 3, 86, 43, 0, 489, 490, 5, 46, 0, 0, 490, 492, 3, 86, 43, 0, 491,
		489, 1, 0, 0, 0, 492, 495, 1, 0, 0, 0, 493, 491, 1, 0, 0, 0, 493, 494,
		1, 0, 0, 0, 494, 512, 1, 0, 0, 0, 495, 493, 1, 0, 0, 0, 496, 501, 3, 56,
		28, 0, 497, 498, 5, 46, 0, 0, 498, 500, 3, 56, 28, 0, 499, 497, 1, 0, 0,
		0, 500, 503, 1, 0, 0, 0, 501, 499, 1, 0, 0, 0, 501, 502, 1, 0, 0, 0, 502,
		508, 1, 0, 0, 0, 503, 501, 1, 0, 0, 0, 504, 505, 5, 46, 0, 0, 505, 507,
		3, 86, 43, 0, 506, 504, 1, 0, 0, 0, 507, 510, 1, 0, 0, 0, 508, 506, 1,
		0, 0, 0, 508, 509, 1, 0, 0, 0, 509, 512, 1, 0, 0, 0, 510, 508, 1, 0, 0,
		0, 511, 488, 1, 0, 0, 0, 511, 496, 1, 0, 0, 0, 512, 85, 1, 0, 0, 0, 513,
		514, 3, 98, 49, 0, 514, 515, 5, 47, 0, 0, 515, 516, 3, 56, 28, 0, 516,
		87, 1, 0, 0, 0, 517, 523, 5, 60, 0, 0, 518, 523, 5, 28, 0, 0, 519, 523,
		5, 26, 0, 0, 520, 523, 5, 56, 0, 0, 521, 523, 3, 90, 45, 0, 522, 517, 1,
		0, 0, 0, 522, 518, 1, 0, 0, 0, 522, 519, 1, 0, 0, 0, 522, 520, 1, 0, 0,
		0, 522, 521, 1, 0, 0, 0, 523, 89, 1, 0, 0, 0, 524, 533, 5, 57, 0, 0, 525,
		532, 5, 67, 0, 0, 526, 532, 5, 68, 0, 0, 527, 528, 5, 66, 0, 0, 528, 529,
		3, 56, 28, 0, 529, 530, 5, 54, 0, 0, 530, 532, 1, 0, 0, 0, 531, 525, 1,
		0, 0, 0, 531, 526, 1, 0, 0, 0, 531, 527, 1, 0, 0, 0, 532, 535, 1, 0, 0,
		0, 533, 531, 1, 0, 0, 0, 533, 534, 1, 0, 0, 0, 534, 536, 1, 0, 0, 0, 535,
		533, 1, 0, 0, 0, 536, 537, 5, 65, 0, 0, 537, 91, 1, 0, 0, 0, 538, 540,
		5, 58, 0, 0, 539, 541, 7, 5, 0, 0, 540, 539, 1, 0, 0, 0, 541, 542, 1, 0,
		0, 0, 542, 540, 1, 0, 0, 0, 542, 543, 1, 0, 0, 0, 543, 544, 1, 0, 0, 0,
		544, 545, 5, 69, 0, 0, 545, 93, 1, 0, 0, 0, 546, 550, 5, 58, 0, 0, 547,
		549, 7, 5, 0, 0, 548, 547, 1, 0, 0, 0, 549, 552, 1, 0, 0, 0, 550, 548,
		1, 0, 0, 0, 550, 551, 1, 0, 0, 0, 551, 553, 1, 0, 0, 0, 552, 550, 1, 0,
		0, 0, 553, 554, 5, 70, 0, 0, 554, 555, 3, 56, 28, 0, 555, 564, 5, 54, 0,
		0, 556, 563, 5, 71, 0, 0, 557, 563, 5, 72, 0, 0, 558, 559, 5, 70, 0, 0,
		559, 560, 3, 56, 28, 0, 560, 561, 5, 54, 0, 0, 561, 563, 1, 0, 0, 0, 562,
		556, 1, 0, 0, 0, 562, 557, 1, 0, 0, 0, 562, 558, 1, 0, 0, 0, 563, 566,
		1, 0, 0, 0, 564, 562, 1, 0, 0, 0, 564, 565, 1, 0, 0, 0, 565, 567, 1, 0,
		0, 0, 566, 564, 1, 0, 0, 0, 567, 568, 5, 69, 0, 0, 568, 95, 1, 0, 0, 0,
		569, 572, 3, 98, 49, 0, 570, 572, 3, 94, 47, 0, 571, 569, 1, 0, 0, 0, 571,
		570, 1, 0, 0, 0, 572, 97, 1, 0, 0, 0, 573, 577, 5, 62, 0, 0, 574, 577,
		5, 4, 0, 0, 575, 577, 3, 92, 46, 0, 576, 573, 1, 0, 0, 0, 576, 574, 1,
		0, 0, 0, 576, 575, 1, 0, 0, 0, 577, 99, 1, 0, 0, 0, 578, 590, 5, 51, 0,
		0, 579, 584, 3, 56, 28, 0, 580, 581, 5, 46, 0, 0, 581, 583, 3, 56, 28,
		0, 582, 580, 1, 0, 0, 0, 583, 586, 1, 0, 0, 0, 584, 582, 1, 0, 0, 0, 584,
		585, 1, 0, 0, 0, 585, 588, 1, 0, 0, 0, 586, 584, 1, 0, 0, 0, 587, 589,
		5, 46, 0, 0, 588, 587, 1, 0, 0, 0, 588, 589, 1, 0, 0, 0, 589, 591, 1, 0,
		0, 0, 590, 579, 1, 0, 0, 0, 590, 591, 1, 0, 0, 0, 591, 592, 1, 0, 0, 0,
		592, 593, 5, 52, 0, 0, 593, 101, 1, 0, 0, 0, 594, 606, 5, 53, 0, 0, 595,
		600, 3, 104, 52, 0, 596, 597, 5, 46, 0, 0, 597, 599, 3, 104, 52, 0, 598,
		596, 1, 0, 0, 0, 599, 602, 1, 0, 0, 0, 600, 598, 1, 0, 0, 0, 600, 601,
		1, 0, 0, 0, 601, 604, 1, 0, 0, 0, 602, 600, 1, 0, 0, 0, 603, 605, 5, 46,
		0, 0, 604, 603, 1, 0, 0, 0, 604, 605, 1, 0, 0, 0, 605, 607, 1, 0, 0, 0,
		606, 595, 1, 0, 0, 0, 606, 607, 1, 0, 0, 0, 607, 608, 1, 0, 0, 0, 608,
		609, 5, 54, 0, 0, 609, 103, 1, 0, 0, 0, 610, 611, 3, 106, 53, 0, 611, 612,
		5, 47, 0, 0, 612, 613, 3, 56, 28, 0, 613, 105, 1, 0, 0, 0, 614, 617, 3,
		98, 49, 0, 615, 617, 3, 90, 45, 0, 616, 614, 1, 0, 0, 0, 616, 615, 1, 0,
		0, 0, 617, 107, 1, 0, 0, 0, 618, 619, 3, 32, 16, 0, 619, 109, 1, 0, 0,
		0, 620, 634, 3, 112, 56, 0, 621, 630, 5, 49, 0, 0, 622, 627, 3, 112, 56,
		0, 623, 624, 5, 46, 0, 0, 624, 626, 3, 112, 56, 0, 625, 623, 1, 0, 0, 0,
		626, 629, 1, 0, 0, 0, 627, 625, 1, 0, 0, 0, 627, 628, 1, 0, 0, 0, 628,
		631, 1, 0, 0, 0, 629, 627, 1, 0, 0, 0, 630, 622, 1, 0, 0, 0, 630, 631,
		1, 0, 0, 0, 631, 632, 1, 0, 0, 0, 632, 634, 5, 50, 0, 0, 633, 620, 1, 0,
		0, 0, 633, 621, 1, 0, 0, 0, 634, 635, 1, 0, 0, 0, 635, 638, 5, 30, 0, 0,
		636, 639, 3, 56, 28, 0, 637, 639, 3, 114, 57, 0, 638, 636, 1, 0, 0, 0,
		638, 637, 1, 0, 0, 0, 639, 111, 1, 0, 0, 0, 640, 643, 5, 61, 0, 0, 641,
		642, 5, 47, 0, 0, 642, 644, 5, 29, 0, 0, 643, 641, 1, 0, 0, 0, 643, 644,
		1, 0, 0, 0, 644, 653, 1, 0, 0, 0, 645, 651, 5, 35, 0, 0, 646, 648, 7, 1,
		0, 0, 647, 646, 1, 0, 0, 0, 647, 648, 1, 0, 0, 0, 648, 649, 1, 0, 0, 0,
		649, 652, 5, 60, 0, 0, 650, 652, 3, 90, 45, 0, 651, 647, 1, 0, 0, 0, 651,
		650, 1, 0, 0, 0, 652, 654, 1, 0, 0, 0, 653, 645, 1, 0, 0, 0, 653, 654,
		1, 0, 0, 0, 654, 113, 1, 0, 0, 0, 655, 659, 5, 53, 0, 0, 656, 658, 5, 63,
		0, 0, 657, 656, 1, 0, 0, 0, 658, 661, 1, 0, 0, 0, 659, 657, 1, 0, 0, 0,
		659, 660, 1, 0, 0, 0, 660, 681, 1, 0, 0, 0, 661, 659, 1, 0, 0, 0, 662,
		663, 5, 61, 0, 0, 663, 664, 5, 35, 0, 0, 664, 677, 3, 56, 28, 0, 665, 669,
		5, 55, 0, 0, 666, 668, 5, 63, 0, 0, 667, 666, 1, 0, 0, 0, 668, 671, 1,
		0, 0, 0, 669, 667, 1, 0, 0, 0, 669, 670, 1, 0, 0, 0, 670, 678, 1, 0, 0,
		0, 671, 669, 1, 0, 0, 0, 672, 674, 5, 63, 0, 0, 673, 672, 1, 0, 0, 0, 674,
		675, 1, 0, 0, 0, 675, 673, 1, 0, 0, 0, 675, 676, 1, 0, 0, 0, 676, 678,
		1, 0, 0, 0, 677, 665, 1, 0, 0, 0, 677, 673, 1, 0, 0, 0, 678, 680, 1, 0,
		0, 0, 679, 662, 1, 0, 0, 0, 680, 683, 1, 0, 0, 0, 681, 679, 1, 0, 0, 0,
		681, 682, 1, 0, 0, 0, 682, 684, 1, 0, 0, 0, 683, 681, 1, 0, 0, 0, 684,
		685, 5, 17, 0, 0, 685, 687, 3, 56, 28, 0, 686, 688, 5, 55, 0, 0, 687, 686,
		1, 0, 0, 0, 687, 688, 1, 0, 0, 0, 688, 692, 1, 0, 0, 0, 689, 691, 5, 63,
		0, 0, 690, 689, 1, 0, 0, 0, 691, 694, 1, 0, 0, 0, 692, 690, 1, 0, 0, 0,
		692, 693, 1, 0, 0, 0, 693, 695, 1, 0, 0, 0, 694, 692, 1, 0, 0, 0, 695,
		696, 5, 54, 0, 0, 696, 115, 1, 0, 0, 0, 88, 119, 124, 127, 132, 141, 148,
		154, 163, 173, 184, 193, 197, 201, 209, 221, 228, 235, 241, 247, 250, 256,
		264, 272, 277, 287, 296, 301, 306, 314, 320, 338, 343, 350, 354, 361, 369,
		377, 383, 390, 398, 407, 413, 419, 423, 426, 428, 437, 445, 451, 457, 466,
		479, 484, 493, 501, 508, 511, 522, 531, 533, 542, 550, 562, 564, 571, 576,
		584, 588, 590, 600, 604, 606, 616, 627, 630, 633, 638, 643, 647, 651, 653,
		659, 669, 675, 677, 681, 687, 692,
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
	SPL2ParserRULE_query             = 0
	SPL2ParserRULE_pipeline          = 1
	SPL2ParserRULE_start             = 2
	SPL2ParserRULE_command           = 3
	SPL2ParserRULE_fromCommand       = 4
	SPL2ParserRULE_selectCommand     = 5
	SPL2ParserRULE_projection        = 6
	SPL2ParserRULE_dataset           = 7
	SPL2ParserRULE_generator         = 8
	SPL2ParserRULE_evalCommand       = 9
	SPL2ParserRULE_assignment        = 10
	SPL2ParserRULE_whereCommand      = 11
	SPL2ParserRULE_fieldsCommand     = 12
	SPL2ParserRULE_rexCommand        = 13
	SPL2ParserRULE_rexOption         = 14
	SPL2ParserRULE_embeddedCommand   = 15
	SPL2ParserRULE_embeddedText      = 16
	SPL2ParserRULE_moduleSuffix      = 17
	SPL2ParserRULE_moduleDeclaration = 18
	SPL2ParserRULE_searchCommand     = 19
	SPL2ParserRULE_implicitSearch    = 20
	SPL2ParserRULE_searchExpression  = 21
	SPL2ParserRULE_searchXor         = 22
	SPL2ParserRULE_searchAnd         = 23
	SPL2ParserRULE_searchOr          = 24
	SPL2ParserRULE_searchNot         = 25
	SPL2ParserRULE_searchAtom        = 26
	SPL2ParserRULE_searchValue       = 27
	SPL2ParserRULE_expression        = 28
	SPL2ParserRULE_xorExpression     = 29
	SPL2ParserRULE_orExpression      = 30
	SPL2ParserRULE_andExpression     = 31
	SPL2ParserRULE_notExpression     = 32
	SPL2ParserRULE_predicate         = 33
	SPL2ParserRULE_comparison        = 34
	SPL2ParserRULE_additive          = 35
	SPL2ParserRULE_multiplicative    = 36
	SPL2ParserRULE_unary             = 37
	SPL2ParserRULE_access            = 38
	SPL2ParserRULE_accessPart        = 39
	SPL2ParserRULE_primary           = 40
	SPL2ParserRULE_call              = 41
	SPL2ParserRULE_arguments         = 42
	SPL2ParserRULE_namedArgument     = 43
	SPL2ParserRULE_literal           = 44
	SPL2ParserRULE_stringLiteral     = 45
	SPL2ParserRULE_quotedName        = 46
	SPL2ParserRULE_fieldTemplate     = 47
	SPL2ParserRULE_fieldName         = 48
	SPL2ParserRULE_identifier        = 49
	SPL2ParserRULE_array             = 50
	SPL2ParserRULE_object            = 51
	SPL2ParserRULE_objectEntry       = 52
	SPL2ParserRULE_objectKey         = 53
	SPL2ParserRULE_searchLiteral     = 54
	SPL2ParserRULE_lambdaExpression  = 55
	SPL2ParserRULE_lambdaParameter   = 56
	SPL2ParserRULE_lambdaBlock       = 57
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

	p.EnterOuterAlt(localctx, 1)
	p.SetState(119)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(116)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(121)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(127)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserFROM, SPL2ParserSELECT, SPL2ParserSEARCH, SPL2ParserINDEX, SPL2ParserMAKERESULTS, SPL2ParserSPL1, SPL2ParserBACKTICK:
		{
			p.SetState(122)
			p.Pipeline()
		}
		p.SetState(124)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserSEMI {
			{
				p.SetState(123)
				p.ModuleSuffix()
			}

		}

	case SPL2ParserIMPORT, SPL2ParserEXPORT, SPL2ParserFUNCTION, SPL2ParserLOCAL:
		{
			p.SetState(126)
			p.ModuleDeclaration()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.SetState(132)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(129)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(134)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(135)
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
		p.SetState(137)
		p.Start_()
	}
	p.SetState(154)
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
			p.SetState(141)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(138)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(143)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(144)
				p.Match(SPL2ParserPIPE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(148)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(145)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(150)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(151)
				p.Command()
			}

		}
		p.SetState(156)
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
	p.SetState(163)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserFROM:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(157)
			p.FromCommand()
		}

	case SPL2ParserSELECT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(158)
			p.SelectCommand()
		}

	case SPL2ParserSEARCH:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(159)
			p.SearchCommand()
		}

	case SPL2ParserINDEX:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(160)
			p.ImplicitSearch()
		}

	case SPL2ParserMAKERESULTS:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(161)
			p.Generator()
		}

	case SPL2ParserSPL1, SPL2ParserBACKTICK:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(162)
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
	p.SetState(173)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserEVAL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(165)
			p.EvalCommand()
		}

	case SPL2ParserWHERE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(166)
			p.WhereCommand()
		}

	case SPL2ParserFIELDS, SPL2ParserTABLE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(167)
			p.FieldsCommand()
		}

	case SPL2ParserFROM:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(168)
			p.FromCommand()
		}

	case SPL2ParserSELECT:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(169)
			p.SelectCommand()
		}

	case SPL2ParserSEARCH:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(170)
			p.SearchCommand()
		}

	case SPL2ParserREX:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(171)
			p.RexCommand()
		}

	case SPL2ParserSPL1, SPL2ParserBACKTICK:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(172)
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
		p.SetState(175)
		p.Match(SPL2ParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
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
		p.SetState(178)
		p.Match(SPL2ParserSELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(179)
		p.Projection()
	}
	p.SetState(184)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(180)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(181)
			p.Projection()
		}

		p.SetState(186)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
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
		p.SetState(190)
		p.Expression()
	}
	p.SetState(193)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserAS {
		{
			p.SetState(191)
			p.Match(SPL2ParserAS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(192)
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
	p.SetState(197)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(195)
			p.Identifier()
		}

	case SPL2ParserLBRACKET:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(196)
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
		p.SetState(199)
		p.Match(SPL2ParserMAKERESULTS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(201)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserNUMBER {
		{
			p.SetState(200)
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
		p.SetState(203)
		p.Match(SPL2ParserEVAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(204)
		p.Assignment()
	}
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(205)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(206)
			p.Assignment()
		}

		p.SetState(211)
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
		p.SetState(212)
		p.FieldName()
	}
	{
		p.SetState(213)
		p.Match(SPL2ParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(214)
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
		p.SetState(216)
		p.Match(SPL2ParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(217)
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
		p.SetState(219)
		_la = p.GetTokenStream().LA(1)

		if !(_la == SPL2ParserFIELDS || _la == SPL2ParserTABLE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	p.SetState(221)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserPLUS || _la == SPL2ParserMINUS {
		{
			p.SetState(220)
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
		p.SetState(223)
		p.Identifier()
	}
	p.SetState(228)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserCOMMA {
		{
			p.SetState(224)
			p.Match(SPL2ParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(225)
			p.Identifier()
		}

		p.SetState(230)
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
		p.SetState(231)
		p.Match(SPL2ParserREX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(235)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserIDENTIFIER {
		{
			p.SetState(232)
			p.RexOption()
		}

		p.SetState(237)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(241)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserREGEX:
		{
			p.SetState(238)
			p.Match(SPL2ParserREGEX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		{
			p.SetState(239)
			p.StringLiteral()
		}

	case SPL2ParserRAW_STRING:
		{
			p.SetState(240)
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
		p.SetState(243)
		p.Match(SPL2ParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(244)
		p.Match(SPL2ParserASSIGN)
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

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		{
			p.SetState(245)
			p.Identifier()
		}

	case SPL2ParserNUMBER:
		{
			p.SetState(246)
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
	p.SetState(250)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserSPL1 {
		{
			p.SetState(249)
			p.Match(SPL2ParserSPL1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(252)
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
		p.SetState(254)
		p.Match(SPL2ParserBACKTICK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(256)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserEMBEDDED_TEXT {
		{
			p.SetState(255)
			p.Match(SPL2ParserEMBEDDED_TEXT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(258)
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
		p.SetState(260)
		p.Match(SPL2ParserSEMI)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(264)
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
			p.SetState(261)
			p.MatchWildcard()

		}
		p.SetState(266)
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
	p.SetState(272)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserIMPORT:
		{
			p.SetState(267)
			p.Match(SPL2ParserIMPORT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserEXPORT:
		{
			p.SetState(268)
			p.Match(SPL2ParserEXPORT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserFUNCTION:
		{
			p.SetState(269)
			p.Match(SPL2ParserFUNCTION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserLOCAL:
		{
			p.SetState(270)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(271)
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
	p.SetState(277)
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
			p.SetState(274)
			p.MatchWildcard()

		}
		p.SetState(279)
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
		p.SetState(280)
		p.Match(SPL2ParserSEARCH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(281)
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
	INDEX() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	SearchValue() ISearchValueContext
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

func (s *ImplicitSearchContext) INDEX() antlr.TerminalNode {
	return s.GetToken(SPL2ParserINDEX, 0)
}

func (s *ImplicitSearchContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserASSIGN, 0)
}

func (s *ImplicitSearchContext) SearchValue() ISearchValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISearchValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISearchValueContext)
}

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
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(283)
		p.Match(SPL2ParserINDEX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(284)
		p.Match(SPL2ParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(285)
		p.SearchValue()
	}
	p.SetState(287)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&6269574730766876688) != 0 {
		{
			p.SetState(286)
			p.SearchExpression()
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
		p.SetState(289)
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
		p.SetState(291)
		p.SearchAnd()
	}
	p.SetState(296)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserXOR {
		{
			p.SetState(292)
			p.Match(SPL2ParserXOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(293)
			p.SearchAnd()
		}

		p.SetState(298)
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
		p.SetState(299)
		p.SearchOr()
	}
	p.SetState(306)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&6269574730767138832) != 0 {
		p.SetState(301)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserAND {
			{
				p.SetState(300)
				p.Match(SPL2ParserAND)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(303)
			p.SearchOr()
		}

		p.SetState(308)
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
		p.SetState(309)
		p.SearchNot()
	}
	p.SetState(314)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserOR {
		{
			p.SetState(310)
			p.Match(SPL2ParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(311)
			p.SearchNot()
		}

		p.SetState(316)
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
	p.SetState(320)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserNOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(317)
			p.Match(SPL2ParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(318)
			p.SearchNot()
		}

	case SPL2ParserINDEX, SPL2ParserSTAR, SPL2ParserLPAREN, SPL2ParserRAW_STRING, SPL2ParserDQUOTE, SPL2ParserSQUOTE, SPL2ParserNUMBER, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(319)
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

	p.SetState(343)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(322)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(323)
			p.SearchExpression()
		}
		{
			p.SetState(324)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(326)
			p.Identifier()
		}
		{
			p.SetState(327)
			p.Comparison()
		}
		{
			p.SetState(328)
			p.SearchValue()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(330)
			p.Identifier()
		}
		{
			p.SetState(331)
			p.Match(SPL2ParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(332)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(333)
			p.SearchValue()
		}
		p.SetState(336)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == SPL2ParserCOMMA {
			{
				p.SetState(334)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(335)
				p.SearchValue()
			}

			p.SetState(338)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(340)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(342)
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
	p.SetState(350)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(345)
			p.Identifier()
		}

	case SPL2ParserNUMBER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(346)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(347)
			p.StringLiteral()
		}

	case SPL2ParserRAW_STRING:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(348)
			p.Match(SPL2ParserRAW_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserSTAR:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(349)
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
	p.SetState(354)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(352)
			p.LambdaExpression()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(353)
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
	AllXOR() []antlr.TerminalNode
	XOR(i int) antlr.TerminalNode

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

func (s *XorExpressionContext) AllXOR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserXOR)
}

func (s *XorExpressionContext) XOR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserXOR, i)
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
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(356)
		p.OrExpression()
	}
	p.SetState(361)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserXOR {
		{
			p.SetState(357)
			p.Match(SPL2ParserXOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(358)
			p.OrExpression()
		}

		p.SetState(363)
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

// IOrExpressionContext is an interface to support dynamic dispatch.
type IOrExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAndExpression() []IAndExpressionContext
	AndExpression(i int) IAndExpressionContext
	AllOR() []antlr.TerminalNode
	OR(i int) antlr.TerminalNode

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

func (s *OrExpressionContext) AllOR() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserOR)
}

func (s *OrExpressionContext) OR(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserOR, i)
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
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(364)
		p.AndExpression()
	}
	p.SetState(369)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserOR {
		{
			p.SetState(365)
			p.Match(SPL2ParserOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(366)
			p.AndExpression()
		}

		p.SetState(371)
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

// IAndExpressionContext is an interface to support dynamic dispatch.
type IAndExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNotExpression() []INotExpressionContext
	NotExpression(i int) INotExpressionContext
	AllAND() []antlr.TerminalNode
	AND(i int) antlr.TerminalNode

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

func (s *AndExpressionContext) AllAND() []antlr.TerminalNode {
	return s.GetTokens(SPL2ParserAND)
}

func (s *AndExpressionContext) AND(i int) antlr.TerminalNode {
	return s.GetToken(SPL2ParserAND, i)
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
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(372)
		p.NotExpression()
	}
	p.SetState(377)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserAND {
		{
			p.SetState(373)
			p.Match(SPL2ParserAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(374)
			p.NotExpression()
		}

		p.SetState(379)
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

// INotExpressionContext is an interface to support dynamic dispatch.
type INotExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NOT() antlr.TerminalNode
	NotExpression() INotExpressionContext
	Predicate() IPredicateContext

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

func (s *NotExpressionContext) NOT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNOT, 0)
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
	p.SetState(383)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserNOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(380)
			p.Match(SPL2ParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(381)
			p.NotExpression()
		}

	case SPL2ParserINDEX, SPL2ParserNULL, SPL2ParserBOOLEAN, SPL2ParserPLUS, SPL2ParserMINUS, SPL2ParserLPAREN, SPL2ParserLBRACKET, SPL2ParserLBRACE, SPL2ParserRAW_STRING, SPL2ParserDQUOTE, SPL2ParserSQUOTE, SPL2ParserBACKTICK, SPL2ParserNUMBER, SPL2ParserLOCAL, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(382)
			p.Predicate()
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

// IPredicateContext is an interface to support dynamic dispatch.
type IPredicateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllAdditive() []IAdditiveContext
	Additive(i int) IAdditiveContext
	Comparison() IComparisonContext
	BETWEEN() antlr.TerminalNode
	AND() antlr.TerminalNode
	IN() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	RPAREN() antlr.TerminalNode
	LIKE() antlr.TerminalNode
	IS() antlr.TerminalNode
	TYPE() antlr.TerminalNode
	NOT() antlr.TerminalNode
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

func (s *PredicateContext) BETWEEN() antlr.TerminalNode {
	return s.GetToken(SPL2ParserBETWEEN, 0)
}

func (s *PredicateContext) AND() antlr.TerminalNode {
	return s.GetToken(SPL2ParserAND, 0)
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

func (s *PredicateContext) NOT() antlr.TerminalNode {
	return s.GetToken(SPL2ParserNOT, 0)
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
		p.SetState(385)
		p.Additive()
	}
	p.SetState(428)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(386)
			p.Comparison()
		}
		{
			p.SetState(387)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) == 2 {
		p.SetState(390)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserNOT {
			{
				p.SetState(389)
				p.Match(SPL2ParserNOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(392)
			p.Match(SPL2ParserBETWEEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(393)
			p.Additive()
		}
		{
			p.SetState(394)
			p.Match(SPL2ParserAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(395)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) == 3 {
		p.SetState(398)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserNOT {
			{
				p.SetState(397)
				p.Match(SPL2ParserNOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(400)
			p.Match(SPL2ParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(401)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(402)
			p.Expression()
		}
		p.SetState(407)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(403)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(404)
				p.Expression()
			}

			p.SetState(409)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(410)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) == 4 {
		p.SetState(413)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserNOT {
			{
				p.SetState(412)
				p.Match(SPL2ParserNOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(415)
			p.Match(SPL2ParserLIKE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(416)
			p.Additive()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	} else if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) == 5 {
		{
			p.SetState(417)
			p.Match(SPL2ParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(426)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) {
		case 1:
			p.SetState(419)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == SPL2ParserNOT {
				{
					p.SetState(418)
					p.Match(SPL2ParserNOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			{
				p.SetState(421)
				_la = p.GetTokenStream().LA(1)

				if !(_la == SPL2ParserNULL || _la == SPL2ParserNULL_TEST) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

		case 2:
			p.SetState(423)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == SPL2ParserNOT {
				{
					p.SetState(422)
					p.Match(SPL2ParserNOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			{
				p.SetState(425)
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
	p.EnterRule(localctx, 68, SPL2ParserRULE_comparison)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(430)
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
	p.EnterRule(localctx, 70, SPL2ParserRULE_additive)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(432)
		p.Multiplicative()
	}
	p.SetState(437)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserPLUS || _la == SPL2ParserMINUS {
		{
			p.SetState(433)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(434)
			p.Multiplicative()
		}

		p.SetState(439)
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
	p.EnterRule(localctx, 72, SPL2ParserRULE_multiplicative)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(440)
		p.Unary()
	}
	p.SetState(445)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&27487790694400) != 0 {
		{
			p.SetState(441)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&27487790694400) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(442)
			p.Unary()
		}

		p.SetState(447)
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
	p.EnterRule(localctx, 74, SPL2ParserRULE_unary)
	var _la int

	p.SetState(451)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserPLUS, SPL2ParserMINUS:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(448)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserPLUS || _la == SPL2ParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(449)
			p.Unary()
		}

	case SPL2ParserINDEX, SPL2ParserNULL, SPL2ParserBOOLEAN, SPL2ParserLPAREN, SPL2ParserLBRACKET, SPL2ParserLBRACE, SPL2ParserRAW_STRING, SPL2ParserDQUOTE, SPL2ParserSQUOTE, SPL2ParserBACKTICK, SPL2ParserNUMBER, SPL2ParserLOCAL, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(450)
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
	p.EnterRule(localctx, 76, SPL2ParserRULE_access)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(453)
		p.Primary()
	}
	p.SetState(457)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserDOT || _la == SPL2ParserLBRACKET {
		{
			p.SetState(454)
			p.AccessPart()
		}

		p.SetState(459)
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
	p.EnterRule(localctx, 78, SPL2ParserRULE_accessPart)
	p.SetState(466)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserDOT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(460)
			p.Match(SPL2ParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(461)
			p.Identifier()
		}

	case SPL2ParserLBRACKET:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(462)
			p.Match(SPL2ParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(463)
			p.Expression()
		}
		{
			p.SetState(464)
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
	p.EnterRule(localctx, 80, SPL2ParserRULE_primary)
	p.SetState(479)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 51, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(468)
			p.Call()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(469)
			p.FieldName()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(470)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(471)
			p.Literal()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(472)
			p.Array()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(473)
			p.Object()
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(474)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(475)
			p.Expression()
		}
		{
			p.SetState(476)
			p.Match(SPL2ParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(478)
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
	p.EnterRule(localctx, 82, SPL2ParserRULE_call)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(481)
		p.Identifier()
	}
	{
		p.SetState(482)
		p.Match(SPL2ParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(484)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&9163137216810057744) != 0 {
		{
			p.SetState(483)
			p.Arguments()
		}

	}
	{
		p.SetState(486)
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
	p.EnterRule(localctx, 84, SPL2ParserRULE_arguments)
	var _la int

	var _alt int

	p.SetState(511)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 56, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(488)
			p.NamedArgument()
		}
		p.SetState(493)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(489)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(490)
				p.NamedArgument()
			}

			p.SetState(495)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(496)
			p.Expression()
		}
		p.SetState(501)
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
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(498)
					p.Expression()
				}

			}
			p.SetState(503)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 54, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(508)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SPL2ParserCOMMA {
			{
				p.SetState(504)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(505)
				p.NamedArgument()
			}

			p.SetState(510)
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
	p.EnterRule(localctx, 86, SPL2ParserRULE_namedArgument)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(513)
		p.Identifier()
	}
	{
		p.SetState(514)
		p.Match(SPL2ParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(515)
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
	p.EnterRule(localctx, 88, SPL2ParserRULE_literal)
	p.SetState(522)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserNUMBER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(517)
			p.Match(SPL2ParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserBOOLEAN:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(518)
			p.Match(SPL2ParserBOOLEAN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserNULL:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(519)
			p.Match(SPL2ParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserRAW_STRING:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(520)
			p.Match(SPL2ParserRAW_STRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(521)
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
	p.EnterRule(localctx, 90, SPL2ParserRULE_stringLiteral)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(524)
		p.Match(SPL2ParserDQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(533)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-66)) & ^0x3f) == 0 && ((int64(1)<<(_la-66))&7) != 0 {
		p.SetState(531)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserSTRING_TEXT:
			{
				p.SetState(525)
				p.Match(SPL2ParserSTRING_TEXT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserSTRING_DOLLAR:
			{
				p.SetState(526)
				p.Match(SPL2ParserSTRING_DOLLAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserSTRING_INTERPOLATION:
			{
				p.SetState(527)
				p.Match(SPL2ParserSTRING_INTERPOLATION)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(528)
				p.Expression()
			}
			{
				p.SetState(529)
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

		p.SetState(535)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(536)
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
	p.EnterRule(localctx, 92, SPL2ParserRULE_quotedName)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(538)
		p.Match(SPL2ParserSQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(540)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR {
		{
			p.SetState(539)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(542)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(544)
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
	p.EnterRule(localctx, 94, SPL2ParserRULE_fieldTemplate)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(546)
		p.Match(SPL2ParserSQUOTE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(550)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR {
		{
			p.SetState(547)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPL2ParserNAME_TEXT || _la == SPL2ParserNAME_DOLLAR) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

		p.SetState(552)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(553)
		p.Match(SPL2ParserNAME_INTERPOLATION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(554)
		p.Expression()
	}
	{
		p.SetState(555)
		p.Match(SPL2ParserRBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(564)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-70)) & ^0x3f) == 0 && ((int64(1)<<(_la-70))&7) != 0 {
		p.SetState(562)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserNAME_TEXT:
			{
				p.SetState(556)
				p.Match(SPL2ParserNAME_TEXT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserNAME_DOLLAR:
			{
				p.SetState(557)
				p.Match(SPL2ParserNAME_DOLLAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserNAME_INTERPOLATION:
			{
				p.SetState(558)
				p.Match(SPL2ParserNAME_INTERPOLATION)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(559)
				p.Expression()
			}
			{
				p.SetState(560)
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

		p.SetState(566)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(567)
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
	p.EnterRule(localctx, 96, SPL2ParserRULE_fieldName)
	p.SetState(571)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 64, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(569)
			p.Identifier()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(570)
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
	p.EnterRule(localctx, 98, SPL2ParserRULE_identifier)
	p.SetState(576)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(573)
			p.Match(SPL2ParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserINDEX:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(574)
			p.Match(SPL2ParserINDEX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SPL2ParserSQUOTE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(575)
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
	p.EnterRule(localctx, 100, SPL2ParserRULE_array)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(578)
		p.Match(SPL2ParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(590)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&9163137216810057744) != 0 {
		{
			p.SetState(579)
			p.Expression()
		}
		p.SetState(584)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 66, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(580)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(581)
					p.Expression()
				}

			}
			p.SetState(586)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 66, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(588)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserCOMMA {
			{
				p.SetState(587)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	}
	{
		p.SetState(592)
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
	p.EnterRule(localctx, 102, SPL2ParserRULE_object)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(594)
		p.Match(SPL2ParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(606)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&5044031582654955536) != 0 {
		{
			p.SetState(595)
			p.ObjectEntry()
		}
		p.SetState(600)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 69, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(596)
					p.Match(SPL2ParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(597)
					p.ObjectEntry()
				}

			}
			p.SetState(602)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 69, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(604)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserCOMMA {
			{
				p.SetState(603)
				p.Match(SPL2ParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	}
	{
		p.SetState(608)
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
	p.EnterRule(localctx, 104, SPL2ParserRULE_objectEntry)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(610)
		p.ObjectKey()
	}
	{
		p.SetState(611)
		p.Match(SPL2ParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(612)
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
	p.EnterRule(localctx, 106, SPL2ParserRULE_objectKey)
	p.SetState(616)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserINDEX, SPL2ParserSQUOTE, SPL2ParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(614)
			p.Identifier()
		}

	case SPL2ParserDQUOTE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(615)
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
	p.EnterRule(localctx, 108, SPL2ParserRULE_searchLiteral)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(618)
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
	p.EnterRule(localctx, 110, SPL2ParserRULE_lambdaExpression)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(633)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPL2ParserLOCAL:
		{
			p.SetState(620)
			p.LambdaParameter()
		}

	case SPL2ParserLPAREN:
		{
			p.SetState(621)
			p.Match(SPL2ParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(630)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPL2ParserLOCAL {
			{
				p.SetState(622)
				p.LambdaParameter()
			}
			p.SetState(627)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserCOMMA {
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
					p.LambdaParameter()
				}

				p.SetState(629)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(632)
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
		p.SetState(635)
		p.Match(SPL2ParserARROW)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(638)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 76, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(636)
			p.Expression()
		}

	case 2:
		{
			p.SetState(637)
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
	p.EnterRule(localctx, 112, SPL2ParserRULE_lambdaParameter)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(640)
		p.Match(SPL2ParserLOCAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(643)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserCOLON {
		{
			p.SetState(641)
			p.Match(SPL2ParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(642)
			p.Match(SPL2ParserTYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(653)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserASSIGN {
		{
			p.SetState(645)
			p.Match(SPL2ParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(651)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserPLUS, SPL2ParserMINUS, SPL2ParserNUMBER:
			p.SetState(647)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == SPL2ParserPLUS || _la == SPL2ParserMINUS {
				{
					p.SetState(646)
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
				p.SetState(649)
				p.Match(SPL2ParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case SPL2ParserDQUOTE:
			{
				p.SetState(650)
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
	p.EnterRule(localctx, 114, SPL2ParserRULE_lambdaBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(655)
		p.Match(SPL2ParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(659)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(656)
			p.Match(SPL2ParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(661)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(681)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserLOCAL {
		{
			p.SetState(662)
			p.Match(SPL2ParserLOCAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(663)
			p.Match(SPL2ParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(664)
			p.Expression()
		}
		p.SetState(677)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SPL2ParserSEMI:
			{
				p.SetState(665)
				p.Match(SPL2ParserSEMI)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(669)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPL2ParserNL {
				{
					p.SetState(666)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(671)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		case SPL2ParserNL:
			p.SetState(673)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for ok := true; ok; ok = _la == SPL2ParserNL {
				{
					p.SetState(672)
					p.Match(SPL2ParserNL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(675)
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

		p.SetState(683)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(684)
		p.Match(SPL2ParserRETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(685)
		p.Expression()
	}
	p.SetState(687)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPL2ParserSEMI {
		{
			p.SetState(686)
			p.Match(SPL2ParserSEMI)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(692)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPL2ParserNL {
		{
			p.SetState(689)
			p.Match(SPL2ParserNL)
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
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(695)
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
