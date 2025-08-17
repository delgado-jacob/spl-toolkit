// Code generated from spl-toolkit/grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SPLParser

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

type SPLParser struct {
	*antlr.BaseParser
}

var SPLParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func splparserParserInit() {
	staticData := &SPLParserParserStaticData
	staticData.LiteralNames = []string{
		"", "'+'", "'-'", "'*'", "'/'", "'%'", "'^'", "'AND'", "'OR'", "'NOT'",
		"'='", "'!='", "'>'", "'<'", "'>='", "'<='", "'|'", "'('", "')'", "'['",
		"']'", "','", "':'", "'@'", "'\"'", "'AS'", "'BY'", "'OUTPUT'", "'OUTPUTNEW'",
		"'IN'", "'LIKE'", "", "", "", "", "'now'",
	}
	staticData.SymbolicNames = []string{
		"", "ADD", "SUB", "MULT", "DIV", "MOD", "POW", "AND", "OR", "NOT", "EQ",
		"NE", "GT", "LT", "GE", "LE", "PIPE", "LPAREN", "RPAREN", "LBRACK",
		"RBRACK", "COMMA", "COLON", "AT", "QUOTE", "AS", "BY", "OUTPUT", "OUTPUTNEW",
		"IN", "LIKE", "INIT_COMMAND", "STD_COMMAND_AND_FUNCTION", "STD_COMMAND",
		"MODIFIER_AND_FUNCTION", "TIME_AND_FUNCTION", "FUNCTION", "TIME", "NUMBER",
		"STRING", "IDENTIFIER", "WS", "LINE_COMMENT", "BLOCK_COMMENT",
	}
	staticData.RuleNames = []string{
		"query", "initCommand", "nextCommand", "subquery", "operation", "expression",
		"value", "date", "id", "function", "command",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 43, 224, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 1, 0, 1, 0, 1, 0, 5, 0, 26, 8, 0, 10, 0, 12, 0, 29, 9, 0, 1, 0, 1,
		0, 1, 1, 3, 1, 34, 8, 1, 1, 1, 3, 1, 37, 8, 1, 1, 1, 4, 1, 40, 8, 1, 11,
		1, 12, 1, 41, 1, 1, 3, 1, 45, 8, 1, 1, 2, 1, 2, 4, 2, 49, 8, 2, 11, 2,
		12, 2, 50, 1, 2, 3, 2, 54, 8, 2, 1, 3, 1, 3, 1, 3, 1, 3, 5, 3, 60, 8, 3,
		10, 3, 12, 3, 63, 9, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 5, 4, 78, 8, 4, 10, 4, 12, 4, 81, 9, 4, 3,
		4, 83, 8, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 4, 4, 112, 8, 4, 11, 4, 12, 4, 113, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		3, 4, 129, 8, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 5, 4, 137, 8, 4, 10,
		4, 12, 4, 140, 9, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 5, 5, 148, 8,
		5, 10, 5, 12, 5, 151, 9, 5, 3, 5, 153, 8, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1,
		5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 4, 5, 170,
		8, 5, 11, 5, 12, 5, 171, 1, 5, 3, 5, 175, 8, 5, 1, 5, 1, 5, 1, 5, 1, 5,
		1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 5, 5, 191,
		8, 5, 10, 5, 12, 5, 194, 9, 5, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 200, 8, 6,
		1, 6, 3, 6, 203, 8, 6, 1, 7, 3, 7, 206, 8, 7, 1, 7, 1, 7, 3, 7, 210, 8,
		7, 1, 7, 3, 7, 213, 8, 7, 1, 8, 1, 8, 1, 8, 3, 8, 218, 8, 8, 1, 9, 1, 9,
		1, 10, 1, 10, 1, 10, 0, 2, 8, 10, 11, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18,
		20, 0, 6, 1, 0, 27, 28, 1, 0, 10, 15, 2, 0, 3, 3, 5, 5, 1, 0, 1, 2, 3,
		0, 30, 30, 32, 32, 34, 36, 1, 0, 31, 33, 258, 0, 22, 1, 0, 0, 0, 2, 33,
		1, 0, 0, 0, 4, 46, 1, 0, 0, 0, 6, 55, 1, 0, 0, 0, 8, 128, 1, 0, 0, 0, 10,
		174, 1, 0, 0, 0, 12, 202, 1, 0, 0, 0, 14, 212, 1, 0, 0, 0, 16, 217, 1,
		0, 0, 0, 18, 219, 1, 0, 0, 0, 20, 221, 1, 0, 0, 0, 22, 27, 3, 2, 1, 0,
		23, 24, 5, 16, 0, 0, 24, 26, 3, 4, 2, 0, 25, 23, 1, 0, 0, 0, 26, 29, 1,
		0, 0, 0, 27, 25, 1, 0, 0, 0, 27, 28, 1, 0, 0, 0, 28, 30, 1, 0, 0, 0, 29,
		27, 1, 0, 0, 0, 30, 31, 5, 0, 0, 1, 31, 1, 1, 0, 0, 0, 32, 34, 5, 16, 0,
		0, 33, 32, 1, 0, 0, 0, 33, 34, 1, 0, 0, 0, 34, 36, 1, 0, 0, 0, 35, 37,
		5, 31, 0, 0, 36, 35, 1, 0, 0, 0, 36, 37, 1, 0, 0, 0, 37, 39, 1, 0, 0, 0,
		38, 40, 3, 8, 4, 0, 39, 38, 1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41, 39, 1,
		0, 0, 0, 41, 42, 1, 0, 0, 0, 42, 44, 1, 0, 0, 0, 43, 45, 3, 6, 3, 0, 44,
		43, 1, 0, 0, 0, 44, 45, 1, 0, 0, 0, 45, 3, 1, 0, 0, 0, 46, 48, 3, 20, 10,
		0, 47, 49, 3, 8, 4, 0, 48, 47, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0, 50, 48,
		1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 51, 53, 1, 0, 0, 0, 52, 54, 3, 6, 3, 0,
		53, 52, 1, 0, 0, 0, 53, 54, 1, 0, 0, 0, 54, 5, 1, 0, 0, 0, 55, 56, 5, 19,
		0, 0, 56, 61, 3, 2, 1, 0, 57, 58, 5, 16, 0, 0, 58, 60, 3, 4, 2, 0, 59,
		57, 1, 0, 0, 0, 60, 63, 1, 0, 0, 0, 61, 59, 1, 0, 0, 0, 61, 62, 1, 0, 0,
		0, 62, 64, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 64, 65, 5, 20, 0, 0, 65, 7,
		1, 0, 0, 0, 66, 67, 6, 4, -1, 0, 67, 68, 3, 10, 5, 0, 68, 69, 5, 30, 0,
		0, 69, 70, 3, 12, 6, 0, 70, 129, 1, 0, 0, 0, 71, 72, 3, 10, 5, 0, 72, 73,
		5, 29, 0, 0, 73, 82, 5, 17, 0, 0, 74, 79, 3, 10, 5, 0, 75, 76, 5, 21, 0,
		0, 76, 78, 3, 10, 5, 0, 77, 75, 1, 0, 0, 0, 78, 81, 1, 0, 0, 0, 79, 77,
		1, 0, 0, 0, 79, 80, 1, 0, 0, 0, 80, 83, 1, 0, 0, 0, 81, 79, 1, 0, 0, 0,
		82, 74, 1, 0, 0, 0, 82, 83, 1, 0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 85, 5,
		18, 0, 0, 85, 129, 1, 0, 0, 0, 86, 87, 5, 9, 0, 0, 87, 129, 3, 8, 4, 9,
		88, 89, 3, 10, 5, 0, 89, 90, 7, 0, 0, 0, 90, 91, 3, 16, 8, 0, 91, 129,
		1, 0, 0, 0, 92, 93, 3, 10, 5, 0, 93, 94, 3, 10, 5, 0, 94, 95, 5, 27, 0,
		0, 95, 96, 3, 16, 8, 0, 96, 97, 5, 21, 0, 0, 97, 98, 3, 16, 8, 0, 98, 129,
		1, 0, 0, 0, 99, 100, 3, 10, 5, 0, 100, 101, 3, 10, 5, 0, 101, 102, 3, 10,
		5, 0, 102, 103, 7, 0, 0, 0, 103, 104, 3, 16, 8, 0, 104, 105, 5, 21, 0,
		0, 105, 106, 3, 16, 8, 0, 106, 107, 5, 21, 0, 0, 107, 108, 3, 16, 8, 0,
		108, 129, 1, 0, 0, 0, 109, 111, 5, 26, 0, 0, 110, 112, 3, 16, 8, 0, 111,
		110, 1, 0, 0, 0, 112, 113, 1, 0, 0, 0, 113, 111, 1, 0, 0, 0, 113, 114,
		1, 0, 0, 0, 114, 129, 1, 0, 0, 0, 115, 116, 3, 10, 5, 0, 116, 117, 5, 25,
		0, 0, 117, 118, 3, 16, 8, 0, 118, 129, 1, 0, 0, 0, 119, 120, 3, 16, 8,
		0, 120, 121, 7, 1, 0, 0, 121, 122, 3, 10, 5, 0, 122, 129, 1, 0, 0, 0, 123,
		129, 3, 10, 5, 0, 124, 125, 5, 17, 0, 0, 125, 126, 3, 8, 4, 0, 126, 127,
		5, 18, 0, 0, 127, 129, 1, 0, 0, 0, 128, 66, 1, 0, 0, 0, 128, 71, 1, 0,
		0, 0, 128, 86, 1, 0, 0, 0, 128, 88, 1, 0, 0, 0, 128, 92, 1, 0, 0, 0, 128,
		99, 1, 0, 0, 0, 128, 109, 1, 0, 0, 0, 128, 115, 1, 0, 0, 0, 128, 119, 1,
		0, 0, 0, 128, 123, 1, 0, 0, 0, 128, 124, 1, 0, 0, 0, 129, 138, 1, 0, 0,
		0, 130, 131, 10, 13, 0, 0, 131, 132, 5, 7, 0, 0, 132, 137, 3, 8, 4, 14,
		133, 134, 10, 12, 0, 0, 134, 135, 5, 8, 0, 0, 135, 137, 3, 8, 4, 13, 136,
		130, 1, 0, 0, 0, 136, 133, 1, 0, 0, 0, 137, 140, 1, 0, 0, 0, 138, 136,
		1, 0, 0, 0, 138, 139, 1, 0, 0, 0, 139, 9, 1, 0, 0, 0, 140, 138, 1, 0, 0,
		0, 141, 142, 6, 5, -1, 0, 142, 143, 3, 18, 9, 0, 143, 152, 5, 17, 0, 0,
		144, 149, 3, 10, 5, 0, 145, 146, 5, 21, 0, 0, 146, 148, 3, 10, 5, 0, 147,
		145, 1, 0, 0, 0, 148, 151, 1, 0, 0, 0, 149, 147, 1, 0, 0, 0, 149, 150,
		1, 0, 0, 0, 150, 153, 1, 0, 0, 0, 151, 149, 1, 0, 0, 0, 152, 144, 1, 0,
		0, 0, 152, 153, 1, 0, 0, 0, 153, 154, 1, 0, 0, 0, 154, 155, 5, 18, 0, 0,
		155, 175, 1, 0, 0, 0, 156, 157, 5, 17, 0, 0, 157, 158, 3, 10, 5, 0, 158,
		159, 5, 18, 0, 0, 159, 175, 1, 0, 0, 0, 160, 161, 5, 3, 0, 0, 161, 162,
		3, 10, 5, 0, 162, 163, 5, 3, 0, 0, 163, 175, 1, 0, 0, 0, 164, 165, 5, 3,
		0, 0, 165, 175, 3, 10, 5, 6, 166, 175, 5, 3, 0, 0, 167, 168, 5, 4, 0, 0,
		168, 170, 3, 16, 8, 0, 169, 167, 1, 0, 0, 0, 170, 171, 1, 0, 0, 0, 171,
		169, 1, 0, 0, 0, 171, 172, 1, 0, 0, 0, 172, 175, 1, 0, 0, 0, 173, 175,
		3, 12, 6, 0, 174, 141, 1, 0, 0, 0, 174, 156, 1, 0, 0, 0, 174, 160, 1, 0,
		0, 0, 174, 164, 1, 0, 0, 0, 174, 166, 1, 0, 0, 0, 174, 169, 1, 0, 0, 0,
		174, 173, 1, 0, 0, 0, 175, 192, 1, 0, 0, 0, 176, 177, 10, 10, 0, 0, 177,
		178, 5, 6, 0, 0, 178, 191, 3, 10, 5, 10, 179, 180, 10, 9, 0, 0, 180, 181,
		7, 2, 0, 0, 181, 191, 3, 10, 5, 10, 182, 183, 10, 8, 0, 0, 183, 184, 5,
		4, 0, 0, 184, 191, 3, 10, 5, 9, 185, 186, 10, 2, 0, 0, 186, 187, 7, 3,
		0, 0, 187, 191, 3, 10, 5, 3, 188, 189, 10, 5, 0, 0, 189, 191, 5, 3, 0,
		0, 190, 176, 1, 0, 0, 0, 190, 179, 1, 0, 0, 0, 190, 182, 1, 0, 0, 0, 190,
		185, 1, 0, 0, 0, 190, 188, 1, 0, 0, 0, 191, 194, 1, 0, 0, 0, 192, 190,
		1, 0, 0, 0, 192, 193, 1, 0, 0, 0, 193, 11, 1, 0, 0, 0, 194, 192, 1, 0,
		0, 0, 195, 203, 3, 14, 7, 0, 196, 203, 5, 39, 0, 0, 197, 203, 3, 16, 8,
		0, 198, 200, 7, 3, 0, 0, 199, 198, 1, 0, 0, 0, 199, 200, 1, 0, 0, 0, 200,
		201, 1, 0, 0, 0, 201, 203, 5, 38, 0, 0, 202, 195, 1, 0, 0, 0, 202, 196,
		1, 0, 0, 0, 202, 197, 1, 0, 0, 0, 202, 199, 1, 0, 0, 0, 203, 13, 1, 0,
		0, 0, 204, 206, 5, 24, 0, 0, 205, 204, 1, 0, 0, 0, 205, 206, 1, 0, 0, 0,
		206, 207, 1, 0, 0, 0, 207, 209, 5, 35, 0, 0, 208, 210, 5, 24, 0, 0, 209,
		208, 1, 0, 0, 0, 209, 210, 1, 0, 0, 0, 210, 213, 1, 0, 0, 0, 211, 213,
		5, 37, 0, 0, 212, 205, 1, 0, 0, 0, 212, 211, 1, 0, 0, 0, 213, 15, 1, 0,
		0, 0, 214, 218, 5, 40, 0, 0, 215, 218, 3, 20, 10, 0, 216, 218, 3, 18, 9,
		0, 217, 214, 1, 0, 0, 0, 217, 215, 1, 0, 0, 0, 217, 216, 1, 0, 0, 0, 218,
		17, 1, 0, 0, 0, 219, 220, 7, 4, 0, 0, 220, 19, 1, 0, 0, 0, 221, 222, 7,
		5, 0, 0, 222, 21, 1, 0, 0, 0, 26, 27, 33, 36, 41, 44, 50, 53, 61, 79, 82,
		113, 128, 136, 138, 149, 152, 171, 174, 190, 192, 199, 202, 205, 209, 212,
		217,
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

// SPLParserInit initializes any static state used to implement SPLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewSPLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func SPLParserInit() {
	staticData := &SPLParserParserStaticData
	staticData.once.Do(splparserParserInit)
}

// NewSPLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewSPLParser(input antlr.TokenStream) *SPLParser {
	SPLParserInit()
	this := new(SPLParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &SPLParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "SPLParser.g4"

	return this
}

// SPLParser tokens.
const (
	SPLParserEOF                      = antlr.TokenEOF
	SPLParserADD                      = 1
	SPLParserSUB                      = 2
	SPLParserMULT                     = 3
	SPLParserDIV                      = 4
	SPLParserMOD                      = 5
	SPLParserPOW                      = 6
	SPLParserAND                      = 7
	SPLParserOR                       = 8
	SPLParserNOT                      = 9
	SPLParserEQ                       = 10
	SPLParserNE                       = 11
	SPLParserGT                       = 12
	SPLParserLT                       = 13
	SPLParserGE                       = 14
	SPLParserLE                       = 15
	SPLParserPIPE                     = 16
	SPLParserLPAREN                   = 17
	SPLParserRPAREN                   = 18
	SPLParserLBRACK                   = 19
	SPLParserRBRACK                   = 20
	SPLParserCOMMA                    = 21
	SPLParserCOLON                    = 22
	SPLParserAT                       = 23
	SPLParserQUOTE                    = 24
	SPLParserAS                       = 25
	SPLParserBY                       = 26
	SPLParserOUTPUT                   = 27
	SPLParserOUTPUTNEW                = 28
	SPLParserIN                       = 29
	SPLParserLIKE                     = 30
	SPLParserINIT_COMMAND             = 31
	SPLParserSTD_COMMAND_AND_FUNCTION = 32
	SPLParserSTD_COMMAND              = 33
	SPLParserMODIFIER_AND_FUNCTION    = 34
	SPLParserTIME_AND_FUNCTION        = 35
	SPLParserFUNCTION                 = 36
	SPLParserTIME                     = 37
	SPLParserNUMBER                   = 38
	SPLParserSTRING                   = 39
	SPLParserIDENTIFIER               = 40
	SPLParserWS                       = 41
	SPLParserLINE_COMMENT             = 42
	SPLParserBLOCK_COMMENT            = 43
)

// SPLParser rules.
const (
	SPLParserRULE_query       = 0
	SPLParserRULE_initCommand = 1
	SPLParserRULE_nextCommand = 2
	SPLParserRULE_subquery    = 3
	SPLParserRULE_operation   = 4
	SPLParserRULE_expression  = 5
	SPLParserRULE_value       = 6
	SPLParserRULE_date        = 7
	SPLParserRULE_id          = 8
	SPLParserRULE_function    = 9
	SPLParserRULE_command     = 10
)

// IQueryContext is an interface to support dynamic dispatch.
type IQueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	InitCommand() IInitCommandContext
	EOF() antlr.TerminalNode
	AllPIPE() []antlr.TerminalNode
	PIPE(i int) antlr.TerminalNode
	AllNextCommand() []INextCommandContext
	NextCommand(i int) INextCommandContext

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
	p.RuleIndex = SPLParserRULE_query
	return p
}

func InitEmptyQueryContext(p *QueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_query
}

func (*QueryContext) IsQueryContext() {}

func NewQueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QueryContext {
	var p = new(QueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_query

	return p
}

func (s *QueryContext) GetParser() antlr.Parser { return s.parser }

func (s *QueryContext) InitCommand() IInitCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInitCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInitCommandContext)
}

func (s *QueryContext) EOF() antlr.TerminalNode {
	return s.GetToken(SPLParserEOF, 0)
}

func (s *QueryContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(SPLParserPIPE)
}

func (s *QueryContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserPIPE, i)
}

func (s *QueryContext) AllNextCommand() []INextCommandContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INextCommandContext); ok {
			len++
		}
	}

	tst := make([]INextCommandContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INextCommandContext); ok {
			tst[i] = t.(INextCommandContext)
			i++
		}
	}

	return tst
}

func (s *QueryContext) NextCommand(i int) INextCommandContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INextCommandContext); ok {
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

	return t.(INextCommandContext)
}

func (s *QueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QueryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterQuery(s)
	}
}

func (s *QueryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitQuery(s)
	}
}

func (s *QueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitQuery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Query() (localctx IQueryContext) {
	localctx = NewQueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, SPLParserRULE_query)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(22)
		p.InitCommand()
	}
	p.SetState(27)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPLParserPIPE {
		{
			p.SetState(23)
			p.Match(SPLParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(24)
			p.NextCommand()
		}

		p.SetState(29)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(30)
		p.Match(SPLParserEOF)
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

// IInitCommandContext is an interface to support dynamic dispatch.
type IInitCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PIPE() antlr.TerminalNode
	INIT_COMMAND() antlr.TerminalNode
	AllOperation() []IOperationContext
	Operation(i int) IOperationContext
	Subquery() ISubqueryContext

	// IsInitCommandContext differentiates from other interfaces.
	IsInitCommandContext()
}

type InitCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInitCommandContext() *InitCommandContext {
	var p = new(InitCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_initCommand
	return p
}

func InitEmptyInitCommandContext(p *InitCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_initCommand
}

func (*InitCommandContext) IsInitCommandContext() {}

func NewInitCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InitCommandContext {
	var p = new(InitCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_initCommand

	return p
}

func (s *InitCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *InitCommandContext) PIPE() antlr.TerminalNode {
	return s.GetToken(SPLParserPIPE, 0)
}

func (s *InitCommandContext) INIT_COMMAND() antlr.TerminalNode {
	return s.GetToken(SPLParserINIT_COMMAND, 0)
}

func (s *InitCommandContext) AllOperation() []IOperationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOperationContext); ok {
			len++
		}
	}

	tst := make([]IOperationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOperationContext); ok {
			tst[i] = t.(IOperationContext)
			i++
		}
	}

	return tst
}

func (s *InitCommandContext) Operation(i int) IOperationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
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

	return t.(IOperationContext)
}

func (s *InitCommandContext) Subquery() ISubqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISubqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISubqueryContext)
}

func (s *InitCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InitCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InitCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterInitCommand(s)
	}
}

func (s *InitCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitInitCommand(s)
	}
}

func (s *InitCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitInitCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) InitCommand() (localctx IInitCommandContext) {
	localctx = NewInitCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, SPLParserRULE_initCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(33)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPLParserPIPE {
		{
			p.SetState(32)
			p.Match(SPLParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(36)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(35)
			p.Match(SPLParserINIT_COMMAND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	p.SetState(39)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2198033531422) != 0) {
		{
			p.SetState(38)
			p.operation(0)
		}

		p.SetState(41)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPLParserLBRACK {
		{
			p.SetState(43)
			p.Subquery()
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

// INextCommandContext is an interface to support dynamic dispatch.
type INextCommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Command() ICommandContext
	AllOperation() []IOperationContext
	Operation(i int) IOperationContext
	Subquery() ISubqueryContext

	// IsNextCommandContext differentiates from other interfaces.
	IsNextCommandContext()
}

type NextCommandContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNextCommandContext() *NextCommandContext {
	var p = new(NextCommandContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_nextCommand
	return p
}

func InitEmptyNextCommandContext(p *NextCommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_nextCommand
}

func (*NextCommandContext) IsNextCommandContext() {}

func NewNextCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NextCommandContext {
	var p = new(NextCommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_nextCommand

	return p
}

func (s *NextCommandContext) GetParser() antlr.Parser { return s.parser }

func (s *NextCommandContext) Command() ICommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommandContext)
}

func (s *NextCommandContext) AllOperation() []IOperationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOperationContext); ok {
			len++
		}
	}

	tst := make([]IOperationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOperationContext); ok {
			tst[i] = t.(IOperationContext)
			i++
		}
	}

	return tst
}

func (s *NextCommandContext) Operation(i int) IOperationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
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

	return t.(IOperationContext)
}

func (s *NextCommandContext) Subquery() ISubqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISubqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISubqueryContext)
}

func (s *NextCommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NextCommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NextCommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterNextCommand(s)
	}
}

func (s *NextCommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitNextCommand(s)
	}
}

func (s *NextCommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitNextCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) NextCommand() (localctx INextCommandContext) {
	localctx = NewNextCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, SPLParserRULE_nextCommand)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(46)
		p.Command()
	}
	p.SetState(48)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2198033531422) != 0) {
		{
			p.SetState(47)
			p.operation(0)
		}

		p.SetState(50)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(53)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SPLParserLBRACK {
		{
			p.SetState(52)
			p.Subquery()
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

// ISubqueryContext is an interface to support dynamic dispatch.
type ISubqueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACK() antlr.TerminalNode
	InitCommand() IInitCommandContext
	RBRACK() antlr.TerminalNode
	AllPIPE() []antlr.TerminalNode
	PIPE(i int) antlr.TerminalNode
	AllNextCommand() []INextCommandContext
	NextCommand(i int) INextCommandContext

	// IsSubqueryContext differentiates from other interfaces.
	IsSubqueryContext()
}

type SubqueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySubqueryContext() *SubqueryContext {
	var p = new(SubqueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_subquery
	return p
}

func InitEmptySubqueryContext(p *SubqueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_subquery
}

func (*SubqueryContext) IsSubqueryContext() {}

func NewSubqueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SubqueryContext {
	var p = new(SubqueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_subquery

	return p
}

func (s *SubqueryContext) GetParser() antlr.Parser { return s.parser }

func (s *SubqueryContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(SPLParserLBRACK, 0)
}

func (s *SubqueryContext) InitCommand() IInitCommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInitCommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInitCommandContext)
}

func (s *SubqueryContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(SPLParserRBRACK, 0)
}

func (s *SubqueryContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(SPLParserPIPE)
}

func (s *SubqueryContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserPIPE, i)
}

func (s *SubqueryContext) AllNextCommand() []INextCommandContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INextCommandContext); ok {
			len++
		}
	}

	tst := make([]INextCommandContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INextCommandContext); ok {
			tst[i] = t.(INextCommandContext)
			i++
		}
	}

	return tst
}

func (s *SubqueryContext) NextCommand(i int) INextCommandContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INextCommandContext); ok {
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

	return t.(INextCommandContext)
}

func (s *SubqueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SubqueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SubqueryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterSubquery(s)
	}
}

func (s *SubqueryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitSubquery(s)
	}
}

func (s *SubqueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitSubquery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Subquery() (localctx ISubqueryContext) {
	localctx = NewSubqueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, SPLParserRULE_subquery)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(55)
		p.Match(SPLParserLBRACK)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(56)
		p.InitCommand()
	}
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SPLParserPIPE {
		{
			p.SetState(57)
			p.Match(SPLParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(58)
			p.NextCommand()
		}

		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(64)
		p.Match(SPLParserRBRACK)
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

// IOperationContext is an interface to support dynamic dispatch.
type IOperationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsOperationContext differentiates from other interfaces.
	IsOperationContext()
}

type OperationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOperationContext() *OperationContext {
	var p = new(OperationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_operation
	return p
}

func InitEmptyOperationContext(p *OperationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_operation
}

func (*OperationContext) IsOperationContext() {}

func NewOperationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OperationContext {
	var p = new(OperationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_operation

	return p
}

func (s *OperationContext) GetParser() antlr.Parser { return s.parser }

func (s *OperationContext) CopyAll(ctx *OperationContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *OperationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OperationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LIKEOPContext struct {
	OperationContext
}

func NewLIKEOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LIKEOPContext {
	var p = new(LIKEOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *LIKEOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LIKEOPContext) Expression() IExpressionContext {
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

func (s *LIKEOPContext) LIKE() antlr.TerminalNode {
	return s.GetToken(SPLParserLIKE, 0)
}

func (s *LIKEOPContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *LIKEOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterLIKEOP(s)
	}
}

func (s *LIKEOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitLIKEOP(s)
	}
}

func (s *LIKEOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitLIKEOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type ANDOPContext struct {
	OperationContext
}

func NewANDOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ANDOPContext {
	var p = new(ANDOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *ANDOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ANDOPContext) AllOperation() []IOperationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOperationContext); ok {
			len++
		}
	}

	tst := make([]IOperationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOperationContext); ok {
			tst[i] = t.(IOperationContext)
			i++
		}
	}

	return tst
}

func (s *ANDOPContext) Operation(i int) IOperationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
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

	return t.(IOperationContext)
}

func (s *ANDOPContext) AND() antlr.TerminalNode {
	return s.GetToken(SPLParserAND, 0)
}

func (s *ANDOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterANDOP(s)
	}
}

func (s *ANDOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitANDOP(s)
	}
}

func (s *ANDOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitANDOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type OROPContext struct {
	OperationContext
}

func NewOROPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OROPContext {
	var p = new(OROPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *OROPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OROPContext) AllOperation() []IOperationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOperationContext); ok {
			len++
		}
	}

	tst := make([]IOperationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOperationContext); ok {
			tst[i] = t.(IOperationContext)
			i++
		}
	}

	return tst
}

func (s *OROPContext) Operation(i int) IOperationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
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

	return t.(IOperationContext)
}

func (s *OROPContext) OR() antlr.TerminalNode {
	return s.GetToken(SPLParserOR, 0)
}

func (s *OROPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterOROP(s)
	}
}

func (s *OROPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitOROP(s)
	}
}

func (s *OROPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitOROP(s)

	default:
		return t.VisitChildren(s)
	}
}

type INOPContext struct {
	OperationContext
}

func NewINOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *INOPContext {
	var p = new(INOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *INOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *INOPContext) AllExpression() []IExpressionContext {
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

func (s *INOPContext) Expression(i int) IExpressionContext {
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

func (s *INOPContext) IN() antlr.TerminalNode {
	return s.GetToken(SPLParserIN, 0)
}

func (s *INOPContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserLPAREN, 0)
}

func (s *INOPContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserRPAREN, 0)
}

func (s *INOPContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPLParserCOMMA)
}

func (s *INOPContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserCOMMA, i)
}

func (s *INOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterINOP(s)
	}
}

func (s *INOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitINOP(s)
	}
}

func (s *INOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitINOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type NOTOPContext struct {
	OperationContext
}

func NewNOTOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NOTOPContext {
	var p = new(NOTOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *NOTOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NOTOPContext) NOT() antlr.TerminalNode {
	return s.GetToken(SPLParserNOT, 0)
}

func (s *NOTOPContext) Operation() IOperationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOperationContext)
}

func (s *NOTOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterNOTOP(s)
	}
}

func (s *NOTOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitNOTOP(s)
	}
}

func (s *NOTOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitNOTOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type BYOPContext struct {
	OperationContext
}

func NewBYOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BYOPContext {
	var p = new(BYOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *BYOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BYOPContext) BY() antlr.TerminalNode {
	return s.GetToken(SPLParserBY, 0)
}

func (s *BYOPContext) AllId() []IIdContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdContext); ok {
			len++
		}
	}

	tst := make([]IIdContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdContext); ok {
			tst[i] = t.(IIdContext)
			i++
		}
	}

	return tst
}

func (s *BYOPContext) Id(i int) IIdContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
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

	return t.(IIdContext)
}

func (s *BYOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterBYOP(s)
	}
}

func (s *BYOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitBYOP(s)
	}
}

func (s *BYOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitBYOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type PARENOPContext struct {
	OperationContext
}

func NewPARENOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PARENOPContext {
	var p = new(PARENOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *PARENOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PARENOPContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserLPAREN, 0)
}

func (s *PARENOPContext) Operation() IOperationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOperationContext)
}

func (s *PARENOPContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserRPAREN, 0)
}

func (s *PARENOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterPARENOP(s)
	}
}

func (s *PARENOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitPARENOP(s)
	}
}

func (s *PARENOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitPARENOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type OUTPUTMULTIINOPContext struct {
	OperationContext
}

func NewOUTPUTMULTIINOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OUTPUTMULTIINOPContext {
	var p = new(OUTPUTMULTIINOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *OUTPUTMULTIINOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OUTPUTMULTIINOPContext) AllExpression() []IExpressionContext {
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

func (s *OUTPUTMULTIINOPContext) Expression(i int) IExpressionContext {
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

func (s *OUTPUTMULTIINOPContext) AllId() []IIdContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdContext); ok {
			len++
		}
	}

	tst := make([]IIdContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdContext); ok {
			tst[i] = t.(IIdContext)
			i++
		}
	}

	return tst
}

func (s *OUTPUTMULTIINOPContext) Id(i int) IIdContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
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

	return t.(IIdContext)
}

func (s *OUTPUTMULTIINOPContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPLParserCOMMA)
}

func (s *OUTPUTMULTIINOPContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserCOMMA, i)
}

func (s *OUTPUTMULTIINOPContext) OUTPUT() antlr.TerminalNode {
	return s.GetToken(SPLParserOUTPUT, 0)
}

func (s *OUTPUTMULTIINOPContext) OUTPUTNEW() antlr.TerminalNode {
	return s.GetToken(SPLParserOUTPUTNEW, 0)
}

func (s *OUTPUTMULTIINOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterOUTPUTMULTIINOP(s)
	}
}

func (s *OUTPUTMULTIINOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitOUTPUTMULTIINOP(s)
	}
}

func (s *OUTPUTMULTIINOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitOUTPUTMULTIINOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type KEYVALUEOPContext struct {
	OperationContext
}

func NewKEYVALUEOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *KEYVALUEOPContext {
	var p = new(KEYVALUEOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *KEYVALUEOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *KEYVALUEOPContext) Id() IIdContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdContext)
}

func (s *KEYVALUEOPContext) Expression() IExpressionContext {
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

func (s *KEYVALUEOPContext) EQ() antlr.TerminalNode {
	return s.GetToken(SPLParserEQ, 0)
}

func (s *KEYVALUEOPContext) NE() antlr.TerminalNode {
	return s.GetToken(SPLParserNE, 0)
}

func (s *KEYVALUEOPContext) GT() antlr.TerminalNode {
	return s.GetToken(SPLParserGT, 0)
}

func (s *KEYVALUEOPContext) LT() antlr.TerminalNode {
	return s.GetToken(SPLParserLT, 0)
}

func (s *KEYVALUEOPContext) GE() antlr.TerminalNode {
	return s.GetToken(SPLParserGE, 0)
}

func (s *KEYVALUEOPContext) LE() antlr.TerminalNode {
	return s.GetToken(SPLParserLE, 0)
}

func (s *KEYVALUEOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterKEYVALUEOP(s)
	}
}

func (s *KEYVALUEOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitKEYVALUEOP(s)
	}
}

func (s *KEYVALUEOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitKEYVALUEOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type OUTPUTOPContext struct {
	OperationContext
}

func NewOUTPUTOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OUTPUTOPContext {
	var p = new(OUTPUTOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *OUTPUTOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OUTPUTOPContext) Expression() IExpressionContext {
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

func (s *OUTPUTOPContext) Id() IIdContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdContext)
}

func (s *OUTPUTOPContext) OUTPUT() antlr.TerminalNode {
	return s.GetToken(SPLParserOUTPUT, 0)
}

func (s *OUTPUTOPContext) OUTPUTNEW() antlr.TerminalNode {
	return s.GetToken(SPLParserOUTPUTNEW, 0)
}

func (s *OUTPUTOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterOUTPUTOP(s)
	}
}

func (s *OUTPUTOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitOUTPUTOP(s)
	}
}

func (s *OUTPUTOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitOUTPUTOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type EXPRESSIONOPContext struct {
	OperationContext
}

func NewEXPRESSIONOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPRESSIONOPContext {
	var p = new(EXPRESSIONOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *EXPRESSIONOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPRESSIONOPContext) Expression() IExpressionContext {
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

func (s *EXPRESSIONOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterEXPRESSIONOP(s)
	}
}

func (s *EXPRESSIONOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitEXPRESSIONOP(s)
	}
}

func (s *EXPRESSIONOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitEXPRESSIONOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type OUTPUTMULTIOPContext struct {
	OperationContext
}

func NewOUTPUTMULTIOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OUTPUTMULTIOPContext {
	var p = new(OUTPUTMULTIOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *OUTPUTMULTIOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OUTPUTMULTIOPContext) AllExpression() []IExpressionContext {
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

func (s *OUTPUTMULTIOPContext) Expression(i int) IExpressionContext {
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

func (s *OUTPUTMULTIOPContext) OUTPUT() antlr.TerminalNode {
	return s.GetToken(SPLParserOUTPUT, 0)
}

func (s *OUTPUTMULTIOPContext) AllId() []IIdContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdContext); ok {
			len++
		}
	}

	tst := make([]IIdContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdContext); ok {
			tst[i] = t.(IIdContext)
			i++
		}
	}

	return tst
}

func (s *OUTPUTMULTIOPContext) Id(i int) IIdContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
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

	return t.(IIdContext)
}

func (s *OUTPUTMULTIOPContext) COMMA() antlr.TerminalNode {
	return s.GetToken(SPLParserCOMMA, 0)
}

func (s *OUTPUTMULTIOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterOUTPUTMULTIOP(s)
	}
}

func (s *OUTPUTMULTIOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitOUTPUTMULTIOP(s)
	}
}

func (s *OUTPUTMULTIOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitOUTPUTMULTIOP(s)

	default:
		return t.VisitChildren(s)
	}
}

type RENAMEOPContext struct {
	OperationContext
}

func NewRENAMEOPContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RENAMEOPContext {
	var p = new(RENAMEOPContext)

	InitEmptyOperationContext(&p.OperationContext)
	p.parser = parser
	p.CopyAll(ctx.(*OperationContext))

	return p
}

func (s *RENAMEOPContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RENAMEOPContext) Expression() IExpressionContext {
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

func (s *RENAMEOPContext) AS() antlr.TerminalNode {
	return s.GetToken(SPLParserAS, 0)
}

func (s *RENAMEOPContext) Id() IIdContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdContext)
}

func (s *RENAMEOPContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterRENAMEOP(s)
	}
}

func (s *RENAMEOPContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitRENAMEOP(s)
	}
}

func (s *RENAMEOPContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitRENAMEOP(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Operation() (localctx IOperationContext) {
	return p.operation(0)
}

func (p *SPLParser) operation(_p int) (localctx IOperationContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewOperationContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IOperationContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 8
	p.EnterRecursionRule(localctx, 8, SPLParserRULE_operation, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLIKEOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(67)
			p.expression(0)
		}
		{
			p.SetState(68)
			p.Match(SPLParserLIKE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(69)
			p.Value()
		}

	case 2:
		localctx = NewINOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(71)
			p.expression(0)
		}
		{
			p.SetState(72)
			p.Match(SPLParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(73)
			p.Match(SPLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(82)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2197966422046) != 0 {
			{
				p.SetState(74)
				p.expression(0)
			}
			p.SetState(79)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPLParserCOMMA {
				{
					p.SetState(75)
					p.Match(SPLParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(76)
					p.expression(0)
				}

				p.SetState(81)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(84)
			p.Match(SPLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewNOTOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(86)
			p.Match(SPLParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(87)
			p.operation(9)
		}

	case 4:
		localctx = NewOUTPUTOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(88)
			p.expression(0)
		}
		{
			p.SetState(89)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPLParserOUTPUT || _la == SPLParserOUTPUTNEW) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(90)
			p.Id()
		}

	case 5:
		localctx = NewOUTPUTMULTIOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(92)
			p.expression(0)
		}
		{
			p.SetState(93)
			p.expression(0)
		}
		{
			p.SetState(94)
			p.Match(SPLParserOUTPUT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(95)
			p.Id()
		}
		{
			p.SetState(96)
			p.Match(SPLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(97)
			p.Id()
		}

	case 6:
		localctx = NewOUTPUTMULTIINOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(99)
			p.expression(0)
		}
		{
			p.SetState(100)
			p.expression(0)
		}
		{
			p.SetState(101)
			p.expression(0)
		}
		{
			p.SetState(102)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SPLParserOUTPUT || _la == SPLParserOUTPUTNEW) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(103)
			p.Id()
		}
		{
			p.SetState(104)
			p.Match(SPLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(105)
			p.Id()
		}
		{
			p.SetState(106)
			p.Match(SPLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(107)
			p.Id()
		}

	case 7:
		localctx = NewBYOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(109)
			p.Match(SPLParserBY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(111)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(110)
					p.Id()
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(113)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 8:
		localctx = NewRENAMEOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(115)
			p.expression(0)
		}
		{
			p.SetState(116)
			p.Match(SPLParserAS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(117)
			p.Id()
		}

	case 9:
		localctx = NewKEYVALUEOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(119)
			p.Id()
		}
		{
			p.SetState(120)
			_la = p.GetTokenStream().LA(1)

			if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&64512) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(121)
			p.expression(0)
		}

	case 10:
		localctx = NewEXPRESSIONOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(123)
			p.expression(0)
		}

	case 11:
		localctx = NewPARENOPContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(124)
			p.Match(SPLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(125)
			p.operation(0)
		}
		{
			p.SetState(126)
			p.Match(SPLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(136)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) {
			case 1:
				localctx = NewANDOPContext(p, NewOperationContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_operation)
				p.SetState(130)

				if !(p.Precpred(p.GetParserRuleContext(), 13)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 13)", ""))
					goto errorExit
				}
				{
					p.SetState(131)
					p.Match(SPLParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(132)
					p.operation(14)
				}

			case 2:
				localctx = NewOROPContext(p, NewOperationContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_operation)
				p.SetState(133)

				if !(p.Precpred(p.GetParserRuleContext(), 12)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 12)", ""))
					goto errorExit
				}
				{
					p.SetState(134)
					p.Match(SPLParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(135)
					p.operation(13)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(140)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
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
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Function() IFunctionContext
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllMULT() []antlr.TerminalNode
	MULT(i int) antlr.TerminalNode
	AllDIV() []antlr.TerminalNode
	DIV(i int) antlr.TerminalNode
	AllId() []IIdContext
	Id(i int) IIdContext
	Value() IValueContext
	POW() antlr.TerminalNode
	MOD() antlr.TerminalNode
	ADD() antlr.TerminalNode
	SUB() antlr.TerminalNode

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
	p.RuleIndex = SPLParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *ExpressionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserLPAREN, 0)
}

func (s *ExpressionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(SPLParserRPAREN, 0)
}

func (s *ExpressionContext) AllExpression() []IExpressionContext {
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

func (s *ExpressionContext) Expression(i int) IExpressionContext {
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

func (s *ExpressionContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SPLParserCOMMA)
}

func (s *ExpressionContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserCOMMA, i)
}

func (s *ExpressionContext) AllMULT() []antlr.TerminalNode {
	return s.GetTokens(SPLParserMULT)
}

func (s *ExpressionContext) MULT(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserMULT, i)
}

func (s *ExpressionContext) AllDIV() []antlr.TerminalNode {
	return s.GetTokens(SPLParserDIV)
}

func (s *ExpressionContext) DIV(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserDIV, i)
}

func (s *ExpressionContext) AllId() []IIdContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdContext); ok {
			len++
		}
	}

	tst := make([]IIdContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdContext); ok {
			tst[i] = t.(IIdContext)
			i++
		}
	}

	return tst
}

func (s *ExpressionContext) Id(i int) IIdContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
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

	return t.(IIdContext)
}

func (s *ExpressionContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ExpressionContext) POW() antlr.TerminalNode {
	return s.GetToken(SPLParserPOW, 0)
}

func (s *ExpressionContext) MOD() antlr.TerminalNode {
	return s.GetToken(SPLParserMOD, 0)
}

func (s *ExpressionContext) ADD() antlr.TerminalNode {
	return s.GetToken(SPLParserADD, 0)
}

func (s *ExpressionContext) SUB() antlr.TerminalNode {
	return s.GetToken(SPLParserSUB, 0)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (s *ExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *SPLParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 10
	p.EnterRecursionRule(localctx, 10, SPLParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(174)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 17, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(142)
			p.Function()
		}
		{
			p.SetState(143)
			p.Match(SPLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(152)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2197966422046) != 0 {
			{
				p.SetState(144)
				p.expression(0)
			}
			p.SetState(149)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == SPLParserCOMMA {
				{
					p.SetState(145)
					p.Match(SPLParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(146)
					p.expression(0)
				}

				p.SetState(151)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(154)
			p.Match(SPLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		{
			p.SetState(156)
			p.Match(SPLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(157)
			p.expression(0)
		}
		{
			p.SetState(158)
			p.Match(SPLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		{
			p.SetState(160)
			p.Match(SPLParserMULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(161)
			p.expression(0)
		}
		{
			p.SetState(162)
			p.Match(SPLParserMULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		{
			p.SetState(164)
			p.Match(SPLParserMULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(165)
			p.expression(6)
		}

	case 5:
		{
			p.SetState(166)
			p.Match(SPLParserMULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		p.SetState(169)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(167)
					p.Match(SPLParserDIV)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(168)
					p.Id()
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(171)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 7:
		{
			p.SetState(173)
			p.Value()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(192)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(190)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext()) {
			case 1:
				localctx = NewExpressionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_expression)
				p.SetState(176)

				if !(p.Precpred(p.GetParserRuleContext(), 10)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 10)", ""))
					goto errorExit
				}
				{
					p.SetState(177)
					p.Match(SPLParserPOW)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(178)
					p.expression(10)
				}

			case 2:
				localctx = NewExpressionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_expression)
				p.SetState(179)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				{
					p.SetState(180)
					_la = p.GetTokenStream().LA(1)

					if !(_la == SPLParserMULT || _la == SPLParserMOD) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(181)
					p.expression(10)
				}

			case 3:
				localctx = NewExpressionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_expression)
				p.SetState(182)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				{
					p.SetState(183)
					p.Match(SPLParserDIV)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(184)
					p.expression(9)
				}

			case 4:
				localctx = NewExpressionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_expression)
				p.SetState(185)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(186)
					_la = p.GetTokenStream().LA(1)

					if !(_la == SPLParserADD || _la == SPLParserSUB) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(187)
					p.expression(3)
				}

			case 5:
				localctx = NewExpressionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, SPLParserRULE_expression)
				p.SetState(188)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(189)
					p.Match(SPLParserMULT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(194)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext())
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
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Date() IDateContext
	STRING() antlr.TerminalNode
	Id() IIdContext
	NUMBER() antlr.TerminalNode
	ADD() antlr.TerminalNode
	SUB() antlr.TerminalNode

	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) Date() IDateContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDateContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDateContext)
}

func (s *ValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(SPLParserSTRING, 0)
}

func (s *ValueContext) Id() IIdContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdContext)
}

func (s *ValueContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(SPLParserNUMBER, 0)
}

func (s *ValueContext) ADD() antlr.TerminalNode {
	return s.GetToken(SPLParserADD, 0)
}

func (s *ValueContext) SUB() antlr.TerminalNode {
	return s.GetToken(SPLParserSUB, 0)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitValue(s)
	}
}

func (s *ValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, SPLParserRULE_value)
	var _la int

	p.SetState(202)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(195)
			p.Date()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(196)
			p.Match(SPLParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(197)
			p.Id()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		p.SetState(199)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPLParserADD || _la == SPLParserSUB {
			{
				p.SetState(198)
				_la = p.GetTokenStream().LA(1)

				if !(_la == SPLParserADD || _la == SPLParserSUB) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}

		}
		{
			p.SetState(201)
			p.Match(SPLParserNUMBER)
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

// IDateContext is an interface to support dynamic dispatch.
type IDateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TIME_AND_FUNCTION() antlr.TerminalNode
	AllQUOTE() []antlr.TerminalNode
	QUOTE(i int) antlr.TerminalNode
	TIME() antlr.TerminalNode

	// IsDateContext differentiates from other interfaces.
	IsDateContext()
}

type DateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDateContext() *DateContext {
	var p = new(DateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_date
	return p
}

func InitEmptyDateContext(p *DateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_date
}

func (*DateContext) IsDateContext() {}

func NewDateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DateContext {
	var p = new(DateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_date

	return p
}

func (s *DateContext) GetParser() antlr.Parser { return s.parser }

func (s *DateContext) TIME_AND_FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserTIME_AND_FUNCTION, 0)
}

func (s *DateContext) AllQUOTE() []antlr.TerminalNode {
	return s.GetTokens(SPLParserQUOTE)
}

func (s *DateContext) QUOTE(i int) antlr.TerminalNode {
	return s.GetToken(SPLParserQUOTE, i)
}

func (s *DateContext) TIME() antlr.TerminalNode {
	return s.GetToken(SPLParserTIME, 0)
}

func (s *DateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterDate(s)
	}
}

func (s *DateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitDate(s)
	}
}

func (s *DateContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitDate(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Date() (localctx IDateContext) {
	localctx = NewDateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, SPLParserRULE_date)
	var _la int

	p.SetState(212)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SPLParserQUOTE, SPLParserTIME_AND_FUNCTION:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(205)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SPLParserQUOTE {
			{
				p.SetState(204)
				p.Match(SPLParserQUOTE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(207)
			p.Match(SPLParserTIME_AND_FUNCTION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(209)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 23, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(208)
				p.Match(SPLParserQUOTE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case SPLParserTIME:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(211)
			p.Match(SPLParserTIME)
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

// IIdContext is an interface to support dynamic dispatch.
type IIdContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsIdContext differentiates from other interfaces.
	IsIdContext()
}

type IdContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdContext() *IdContext {
	var p = new(IdContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_id
	return p
}

func InitEmptyIdContext(p *IdContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_id
}

func (*IdContext) IsIdContext() {}

func NewIdContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdContext {
	var p = new(IdContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_id

	return p
}

func (s *IdContext) GetParser() antlr.Parser { return s.parser }

func (s *IdContext) CopyAll(ctx *IdContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *IdContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type CommandUseContext struct {
	IdContext
}

func NewCommandUseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CommandUseContext {
	var p = new(CommandUseContext)

	InitEmptyIdContext(&p.IdContext)
	p.parser = parser
	p.CopyAll(ctx.(*IdContext))

	return p
}

func (s *CommandUseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommandUseContext) Command() ICommandContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommandContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommandContext)
}

func (s *CommandUseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterCommandUse(s)
	}
}

func (s *CommandUseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitCommandUse(s)
	}
}

func (s *CommandUseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitCommandUse(s)

	default:
		return t.VisitChildren(s)
	}
}

type FieldUseContext struct {
	IdContext
}

func NewFieldUseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FieldUseContext {
	var p = new(FieldUseContext)

	InitEmptyIdContext(&p.IdContext)
	p.parser = parser
	p.CopyAll(ctx.(*IdContext))

	return p
}

func (s *FieldUseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldUseContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SPLParserIDENTIFIER, 0)
}

func (s *FieldUseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterFieldUse(s)
	}
}

func (s *FieldUseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitFieldUse(s)
	}
}

func (s *FieldUseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitFieldUse(s)

	default:
		return t.VisitChildren(s)
	}
}

type FunctionUseContext struct {
	IdContext
}

func NewFunctionUseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FunctionUseContext {
	var p = new(FunctionUseContext)

	InitEmptyIdContext(&p.IdContext)
	p.parser = parser
	p.CopyAll(ctx.(*IdContext))

	return p
}

func (s *FunctionUseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionUseContext) Function() IFunctionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionContext)
}

func (s *FunctionUseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterFunctionUse(s)
	}
}

func (s *FunctionUseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitFunctionUse(s)
	}
}

func (s *FunctionUseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitFunctionUse(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Id() (localctx IIdContext) {
	localctx = NewIdContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, SPLParserRULE_id)
	p.SetState(217)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 25, p.GetParserRuleContext()) {
	case 1:
		localctx = NewFieldUseContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(214)
			p.Match(SPLParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewCommandUseContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(215)
			p.Command()
		}

	case 3:
		localctx = NewFunctionUseContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(216)
			p.Function()
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

// IFunctionContext is an interface to support dynamic dispatch.
type IFunctionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FUNCTION() antlr.TerminalNode
	STD_COMMAND_AND_FUNCTION() antlr.TerminalNode
	MODIFIER_AND_FUNCTION() antlr.TerminalNode
	TIME_AND_FUNCTION() antlr.TerminalNode
	LIKE() antlr.TerminalNode

	// IsFunctionContext differentiates from other interfaces.
	IsFunctionContext()
}

type FunctionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionContext() *FunctionContext {
	var p = new(FunctionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_function
	return p
}

func InitEmptyFunctionContext(p *FunctionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_function
}

func (*FunctionContext) IsFunctionContext() {}

func NewFunctionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionContext {
	var p = new(FunctionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_function

	return p
}

func (s *FunctionContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionContext) FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserFUNCTION, 0)
}

func (s *FunctionContext) STD_COMMAND_AND_FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserSTD_COMMAND_AND_FUNCTION, 0)
}

func (s *FunctionContext) MODIFIER_AND_FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserMODIFIER_AND_FUNCTION, 0)
}

func (s *FunctionContext) TIME_AND_FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserTIME_AND_FUNCTION, 0)
}

func (s *FunctionContext) LIKE() antlr.TerminalNode {
	return s.GetToken(SPLParserLIKE, 0)
}

func (s *FunctionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterFunction(s)
	}
}

func (s *FunctionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitFunction(s)
	}
}

func (s *FunctionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitFunction(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Function() (localctx IFunctionContext) {
	localctx = NewFunctionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, SPLParserRULE_function)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(219)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&125627793408) != 0) {
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

// ICommandContext is an interface to support dynamic dispatch.
type ICommandContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INIT_COMMAND() antlr.TerminalNode
	STD_COMMAND() antlr.TerminalNode
	STD_COMMAND_AND_FUNCTION() antlr.TerminalNode

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
	p.RuleIndex = SPLParserRULE_command
	return p
}

func InitEmptyCommandContext(p *CommandContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SPLParserRULE_command
}

func (*CommandContext) IsCommandContext() {}

func NewCommandContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CommandContext {
	var p = new(CommandContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SPLParserRULE_command

	return p
}

func (s *CommandContext) GetParser() antlr.Parser { return s.parser }

func (s *CommandContext) INIT_COMMAND() antlr.TerminalNode {
	return s.GetToken(SPLParserINIT_COMMAND, 0)
}

func (s *CommandContext) STD_COMMAND() antlr.TerminalNode {
	return s.GetToken(SPLParserSTD_COMMAND, 0)
}

func (s *CommandContext) STD_COMMAND_AND_FUNCTION() antlr.TerminalNode {
	return s.GetToken(SPLParserSTD_COMMAND_AND_FUNCTION, 0)
}

func (s *CommandContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommandContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CommandContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.EnterCommand(s)
	}
}

func (s *CommandContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SPLParserListener); ok {
		listenerT.ExitCommand(s)
	}
}

func (s *CommandContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SPLParserVisitor:
		return t.VisitCommand(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SPLParser) Command() (localctx ICommandContext) {
	localctx = NewCommandContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, SPLParserRULE_command)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(221)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&15032385536) != 0) {
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

func (p *SPLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 4:
		var t *OperationContext = nil
		if localctx != nil {
			t = localctx.(*OperationContext)
		}
		return p.Operation_Sempred(t, predIndex)

	case 5:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *SPLParser) Operation_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 13)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 12)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *SPLParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 2:
		return p.Precpred(p.GetParserRuleContext(), 10)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 6:
		return p.Precpred(p.GetParserRuleContext(), 5)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
