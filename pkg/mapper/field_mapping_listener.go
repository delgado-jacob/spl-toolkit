package mapper

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// FieldMappingListener implements grammar-aware field mapping using ANTLR token stream rewriting
type FieldMappingListener struct {
	parser.BaseSPLParserListener

	tokenStream *antlr.CommonTokenStream
	rewriter    *antlr.TokenStreamRewriter
	mappings    map[string]string
}

// NewFieldMappingListener creates a new field mapping listener
func NewFieldMappingListener(tokenStream *antlr.CommonTokenStream, mappings map[string]string) *FieldMappingListener {
	return &FieldMappingListener{
		tokenStream: tokenStream,
		rewriter:    antlr.NewTokenStreamRewriter(tokenStream),
		mappings:    mappings,
	}
}

// GetRewrittenText returns the rewritten query text
func (l *FieldMappingListener) GetRewrittenText() string {
	result := l.rewriter.GetTextDefault()
	// Clean up any EOF tokens that might appear
	if strings.HasSuffix(result, "<EOF>") {
		result = strings.TrimSuffix(result, "<EOF>")
	}
	return strings.TrimSpace(result)
}

// EnterKEYVALUEOP handles field=value operations for field mapping
func (l *FieldMappingListener) EnterKEYVALUEOP(ctx *parser.KEYVALUEOPContext) {
	fieldName := ctx.Id().GetText()

	// Check if this field should be mapped
	if mappedField, exists := l.mappings[fieldName]; exists {
		// Check if this is within an eval command (field assignment) - don't map those
		if !l.isInEvalAssignment(ctx) {
			// Replace the field name with the mapped field
			idCtx := ctx.Id()
			start := idCtx.GetStart().GetTokenIndex()
			stop := idCtx.GetStop().GetTokenIndex()
			l.rewriter.ReplaceDefault(start, stop, mappedField)
		}
	}
}

// EnterBYOP handles "by" operations in stats, etc.
func (l *FieldMappingListener) EnterBYOP(ctx *parser.BYOPContext) {
	// Map fields in "by" clause
	for _, id := range ctx.AllId() {
		fieldName := id.GetText()
		if mappedField, exists := l.mappings[fieldName]; exists {
			start := id.GetStart().GetTokenIndex()
			stop := id.GetStop().GetTokenIndex()
			l.rewriter.ReplaceDefault(start, stop, mappedField)
		}
	}
}

// EnterFieldUse handles general field references
func (l *FieldMappingListener) EnterFieldUse(ctx *parser.FieldUseContext) {
	if identifier := ctx.IDENTIFIER(); identifier != nil {
		fieldName := identifier.GetText()

		// Check if this field should be mapped
		if mappedField, exists := l.mappings[fieldName]; exists {
			// Only map if it's in a field position (not a value position)
			if l.isInFieldPosition(ctx) {
				start := identifier.GetSymbol().GetTokenIndex()
				l.rewriter.ReplaceDefault(start, start, mappedField)
			}
		}
	}
}

// EnterRENAMEOP handles rename operations
func (l *FieldMappingListener) EnterRENAMEOP(ctx *parser.RENAMEOPContext) {
	// Map the source field in rename operations
	if expr := ctx.Expression(); expr != nil {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				fieldName := id.GetText()
				if mappedField, exists := l.mappings[fieldName]; exists {
					start := id.GetStart().GetTokenIndex()
					stop := id.GetStop().GetTokenIndex()
					l.rewriter.ReplaceDefault(start, stop, mappedField)
				}
			}
		}
	}
}

// EnterOUTPUTOP handles OUTPUT operations in lookup commands
func (l *FieldMappingListener) EnterOUTPUTOP(ctx *parser.OUTPUTOPContext) {
	// Map the input field in OUTPUT operations
	if inputExpr := ctx.Expression(); inputExpr != nil {
		if value := inputExpr.Value(); value != nil {
			if id := value.Id(); id != nil {
				fieldName := id.GetText()
				if mappedField, exists := l.mappings[fieldName]; exists {
					start := id.GetStart().GetTokenIndex()
					stop := id.GetStop().GetTokenIndex()
					l.rewriter.ReplaceDefault(start, stop, mappedField)
				}
			}
		}
	}
}

// EnterOUTPUTMULTIOP handles multiple OUTPUT operations
func (l *FieldMappingListener) EnterOUTPUTMULTIOP(ctx *parser.OUTPUTMULTIOPContext) {
	// Map input fields in multi OUTPUT operations
	for _, expr := range ctx.AllExpression() {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				fieldName := id.GetText()
				if mappedField, exists := l.mappings[fieldName]; exists {
					start := id.GetStart().GetTokenIndex()
					stop := id.GetStop().GetTokenIndex()
					l.rewriter.ReplaceDefault(start, stop, mappedField)
				}
			}
		}
	}
}

// EnterOUTPUTMULTIINOP handles multiple input/output operations
func (l *FieldMappingListener) EnterOUTPUTMULTIINOP(ctx *parser.OUTPUTMULTIINOPContext) {
	// Map input fields in multi input/output operations
	for _, expr := range ctx.AllExpression() {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				fieldName := id.GetText()
				if mappedField, exists := l.mappings[fieldName]; exists {
					start := id.GetStart().GetTokenIndex()
					stop := id.GetStop().GetTokenIndex()
					l.rewriter.ReplaceDefault(start, stop, mappedField)
				}
			}
		}
	}
}

// Helper methods

// isInEvalAssignment checks if a KEYVALUEOP is within an eval command (field assignment)
func (l *FieldMappingListener) isInEvalAssignment(ctx *parser.KEYVALUEOPContext) bool {
	parent := ctx.GetParent()
	for parent != nil {
		if nextCmd, ok := parent.(*parser.NextCommandContext); ok {
			if nextCmd.Command() != nil && nextCmd.Command().GetText() == "eval" {
				return true
			}
		}
		parent = parent.GetParent()
	}
	return false
}

// isInFieldPosition determines if a field reference is in a position where it should be mapped
func (l *FieldMappingListener) isInFieldPosition(ctx *parser.FieldUseContext) bool {
	parent := ctx.GetParent()

	for parent != nil {
		switch p := parent.(type) {
		case *parser.KEYVALUEOPContext:
			// If we're the left side of a key=value operation, we're a field reference
			if p.Id() != nil && ctx.GetText() == p.Id().GetText() {
				return true
			}
			// If we're in the expression (right side), we might be a field reference in eval contexts
			return false
		case *parser.BYOPContext:
			// If we're in a "by" clause, we're definitely a field reference
			return true
		case *parser.NextCommandContext:
			// Check the command type
			if p.Command() != nil {
				cmdText := strings.ToLower(p.Command().GetText())
				switch cmdText {
				case "fields":
					return true
				case "lookup", "inputlookup":
					// Be more careful with lookup contexts
					return false
				default:
					return false
				}
			}
		}
		parent = parent.GetParent()
	}

	return false
}
