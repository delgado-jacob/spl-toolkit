package mapper

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// ASTNode represents a node in the Abstract Syntax Tree
type ASTNode struct {
	Type     string          `json:"type"`
	Value    string          `json:"value,omitempty"`
	Children []*ASTNode      `json:"children,omitempty"`
	Context  antlr.ParseTree `json:"-"` // Reference to original ANTLR context
}

// Parser handles SPL query parsing using ANTLR4
type Parser struct {
	errorListener *CustomErrorListener
}

// CustomErrorListener handles parse errors
type CustomErrorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (c *CustomErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	c.errors = append(c.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		errorListener: &CustomErrorListener{
			DefaultErrorListener: antlr.NewDefaultErrorListener(),
			errors:               []string{},
		},
	}
}

// Parse converts a SPL query string into an AST
func (p *Parser) Parse(query string) (*ASTNode, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// Reset errors
	p.errorListener.errors = []string{}

	// Create input stream
	input := antlr.NewInputStream(query)

	// Create lexer
	lexer := parser.NewSPLLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(p.errorListener)

	// Create token stream
	stream := antlr.NewCommonTokenStream(lexer, 0)

	// Create parser
	splParser := parser.NewSPLParser(stream)
	splParser.RemoveErrorListeners()
	splParser.AddErrorListener(p.errorListener)

	// Parse the query
	tree := splParser.Query()

	// Check for errors
	if len(p.errorListener.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(p.errorListener.errors, "; "))
	}

	// Convert ANTLR tree to our AST
	visitor := &ASTVisitor{}
	ast := visitor.Visit(tree).(*ASTNode)

	return ast, nil
}

// ValidateQuery checks if a SPL query is syntactically valid
func (p *Parser) ValidateQuery(query string) error {
	_, err := p.Parse(query)
	return err
}

// ASTVisitor converts ANTLR parse tree to our AST format
type ASTVisitor struct {
	parser.BaseSPLParserVisitor
}

func (v *ASTVisitor) Visit(tree antlr.ParseTree) interface{} {
	switch ctx := tree.(type) {
	case *parser.QueryContext:
		return v.VisitQuery(ctx)
	case *parser.InitCommandContext:
		return v.VisitInitCommand(ctx)
	case *parser.NextCommandContext:
		return v.VisitNextCommand(ctx)
	default:
		return v.VisitChildren(tree)
	}
}

func (v *ASTVisitor) VisitQuery(ctx *parser.QueryContext) interface{} {
	node := &ASTNode{
		Type:     "query",
		Context:  ctx,
		Children: []*ASTNode{},
	}

	// Visit init command
	if ctx.InitCommand() != nil {
		initCmd := v.Visit(ctx.InitCommand()).(*ASTNode)
		node.Children = append(node.Children, initCmd)
	}

	// Visit next commands
	for _, nextCmd := range ctx.AllNextCommand() {
		cmdNode := v.Visit(nextCmd).(*ASTNode)
		node.Children = append(node.Children, cmdNode)
	}

	return node
}

func (v *ASTVisitor) VisitInitCommand(ctx *parser.InitCommandContext) interface{} {
	node := &ASTNode{
		Type:     "init_command",
		Context:  ctx,
		Children: []*ASTNode{},
	}

	// Get command name if present
	if ctx.INIT_COMMAND() != nil {
		node.Value = ctx.INIT_COMMAND().GetText()
	}

	// Visit operations
	for _, op := range ctx.AllOperation() {
		opNode := v.Visit(op).(*ASTNode)
		node.Children = append(node.Children, opNode)
	}

	return node
}

func (v *ASTVisitor) VisitNextCommand(ctx *parser.NextCommandContext) interface{} {
	node := &ASTNode{
		Type:     "next_command",
		Context:  ctx,
		Children: []*ASTNode{},
	}

	// Get command name
	if ctx.Command() != nil {
		node.Value = ctx.Command().GetText()
	}

	// Visit operations
	for _, op := range ctx.AllOperation() {
		opNode := v.Visit(op).(*ASTNode)
		node.Children = append(node.Children, opNode)
	}

	return node
}

func (v *ASTVisitor) VisitChildren(tree antlr.ParseTree) interface{} {
	// Get more specific type information
	nodeType := "node"
	if ctx, ok := tree.(antlr.RuleContext); ok {
		ruleIndex := ctx.GetRuleIndex()
		// Map rule indices to meaningful names
		switch ruleIndex {
		case 0:
			nodeType = "query"
		case 1:
			nodeType = "init_command"
		case 2:
			nodeType = "next_command"
		case 4:
			nodeType = "operation"
		case 5:
			nodeType = "expression"
		case 6:
			nodeType = "value"
		case 8:
			nodeType = "field"
		default:
			nodeType = fmt.Sprintf("rule_%d", ruleIndex)
		}
	}

	node := &ASTNode{
		Type:     nodeType,
		Value:    tree.GetText(),
		Context:  tree,
		Children: []*ASTNode{},
	}

	// Visit all children
	for i := 0; i < tree.GetChildCount(); i++ {
		child := tree.GetChild(i)
		if parseTree, ok := child.(antlr.ParseTree); ok {
			childNode := v.Visit(parseTree).(*ASTNode)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}
