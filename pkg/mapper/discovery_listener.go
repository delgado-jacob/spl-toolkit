package mapper

import (
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strings"
)

// FieldDiscoveryListener implements grammar-aware field discovery using ANTLR listener pattern
type FieldDiscoveryListener struct {
	parser.BaseSPLParserListener

	// Discovered information
	InputFields []string
	SourceTypes []string
	Sources     []string
	DataModels  []string
	Datasets    []string // Format: "DataModel.Dataset"
	Lookups     []string
	Macros      []string

	// Derived field tracking
	derivedFields      map[string]struct{}
	derivedFieldsStack []map[string]struct{}
}

// NewFieldDiscoveryListener creates a new field discovery listener
func NewFieldDiscoveryListener() *FieldDiscoveryListener {
	return &FieldDiscoveryListener{
		InputFields:        []string{},
		SourceTypes:        []string{},
		Sources:            []string{},
		DataModels:         []string{},
		Datasets:           []string{},
		Lookups:            []string{},
		Macros:             []string{},
		derivedFields:      make(map[string]struct{}),
		derivedFieldsStack: []map[string]struct{}{},
	}
}

// Helper methods for derived field tracking
func (l *FieldDiscoveryListener) markDerived(fieldName string) {
	if l.derivedFields == nil {
		l.derivedFields = make(map[string]struct{})
	}
	l.derivedFields[fieldName] = struct{}{}
}

func (l *FieldDiscoveryListener) isDerived(fieldName string) bool {
	_, exists := l.derivedFields[fieldName]
	return exists
}

func (l *FieldDiscoveryListener) pushDerivedContext() {
	l.derivedFieldsStack = append(l.derivedFieldsStack, l.derivedFields)
	l.derivedFields = make(map[string]struct{})
}

func (l *FieldDiscoveryListener) popDerivedContext() {
	if len(l.derivedFieldsStack) > 0 {
		l.derivedFields = l.derivedFieldsStack[len(l.derivedFieldsStack)-1]
		l.derivedFieldsStack = l.derivedFieldsStack[:len(l.derivedFieldsStack)-1]
	}
}

// Helper to add field without duplicates
func (l *FieldDiscoveryListener) addInputField(fieldName string) {
	if fieldName == "" || l.isDerived(fieldName) {
		return
	}

	// Only add if it looks like a legitimate field reference
	if !l.isLikelyFieldReference(fieldName) {
		return
	}

	// Check if already exists
	for _, existing := range l.InputFields {
		if existing == fieldName {
			return
		}
	}

	l.InputFields = append(l.InputFields, fieldName)
}

func (l *FieldDiscoveryListener) addSourceType(sourcetype string) {
	if sourcetype == "" {
		return
	}
	for _, existing := range l.SourceTypes {
		if existing == sourcetype {
			return
		}
	}
	l.SourceTypes = append(l.SourceTypes, sourcetype)
}

func (l *FieldDiscoveryListener) addSource(source string) {
	if source == "" {
		return
	}
	for _, existing := range l.Sources {
		if existing == source {
			return
		}
	}
	l.Sources = append(l.Sources, source)
}

func (l *FieldDiscoveryListener) addLookup(lookup string) {
	if lookup == "" {
		return
	}
	for _, existing := range l.Lookups {
		if existing == lookup {
			return
		}
	}
	l.Lookups = append(l.Lookups, lookup)
}

func (l *FieldDiscoveryListener) addDataModel(datamodel string) {
	if datamodel == "" {
		return
	}
	for _, existing := range l.DataModels {
		if existing == datamodel {
			return
		}
	}
	l.DataModels = append(l.DataModels, datamodel)
}

func (l *FieldDiscoveryListener) addDataset(dataset string) {
	if dataset == "" {
		return
	}
	for _, existing := range l.Datasets {
		if existing == dataset {
			return
		}
	}
	l.Datasets = append(l.Datasets, dataset)
}

func (l *FieldDiscoveryListener) addMacro(macro string) {
	if macro == "" {
		return
	}
	for _, existing := range l.Macros {
		if existing == macro {
			return
		}
	}
	l.Macros = append(l.Macros, macro)
}

// EnterKEYVALUEOP handles field=value operations
func (l *FieldDiscoveryListener) EnterKEYVALUEOP(ctx *parser.KEYVALUEOPContext) {
	fieldName := ctx.Id().GetText()

	// Check if this is within an eval command (field assignment)
	if parentCtx := ctx.GetParent(); parentCtx != nil {
		if nextCmd, ok := parentCtx.(*parser.NextCommandContext); ok {
			if nextCmd.Command() != nil && nextCmd.Command().GetText() == "eval" {
				// This is an eval assignment - mark the field as derived
				l.markDerived(fieldName)

				// Check if the value being assigned is a field reference
				if expr := ctx.Expression(); expr != nil {
					l.extractFieldReferencesFromExpression(expr)
				}
				return
			}
		}
	}

	// Check for special field types
	switch strings.ToLower(fieldName) {
	case "sourcetype":
		if expr := ctx.Expression(); expr != nil {
			if value := expr.Value(); value != nil {
				l.addSourceType(strings.Trim(value.GetText(), "\""))
			}
		}
	case "source":
		if expr := ctx.Expression(); expr != nil {
			if value := expr.Value(); value != nil {
				l.addSource(strings.Trim(value.GetText(), "\""))
			}
		}
	case "datamodel":
		if expr := ctx.Expression(); expr != nil {
			if value := expr.Value(); value != nil {
				l.addDataModel(strings.Trim(value.GetText(), "\""))
			}
		}
	default:
		// Regular field reference
		l.addInputField(fieldName)
	}
}

// extractFieldReferencesFromExpression extracts field references from complex expressions
func (l *FieldDiscoveryListener) extractFieldReferencesFromExpression(expr parser.IExpressionContext) {
	if expr == nil {
		return
	}

	// Check if this expression contains a Value that could be a field reference
	if value := expr.Value(); value != nil {
		valueText := value.GetText()
		// If value is not quoted and looks like a field name, it's likely a field reference
		if !strings.HasPrefix(valueText, "\"") && !strings.HasSuffix(valueText, "\"") && l.isLikelyFieldReference(valueText) {
			l.addInputField(valueText)
		}
	}

	// Recursively check child expressions
	for _, child := range expr.GetChildren() {
		if childExpr, ok := child.(parser.IExpressionContext); ok {
			l.extractFieldReferencesFromExpression(childExpr)
		}
	}
}

// isLikelyFieldReference determines if a string looks like a field reference vs a literal value
func (l *FieldDiscoveryListener) isLikelyFieldReference(text string) bool {
	if text == "" {
		return false
	}

	// Skip concatenated command+field patterns (e.g., "byField", "fromTable", etc.)
	// These indicate parsing artifacts rather than legitimate field names
	commandPrefixes := []string{"by", "from", "where", "select", "as", "into", "with"}
	lowerText := strings.ToLower(text)
	for _, prefix := range commandPrefixes {
		if strings.HasPrefix(lowerText, prefix) && len(text) > len(prefix) {
			// Check if what follows the prefix looks like a field (contains . or starts with uppercase)
			remainder := text[len(prefix):]
			if strings.Contains(remainder, ".") || (len(remainder) > 0 && remainder[0] >= 'A' && remainder[0] <= 'Z') {
				return false
			}
		}
	}

	// Skip numeric values (including IP addresses)
	if strings.ContainsAny(text, "0123456789") && !strings.ContainsAny(text, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_") {
		return false
	}

	// Skip partial IP addresses like ".1.1"
	if strings.HasPrefix(text, ".") && strings.ContainsAny(text, "0123456789") {
		return false
	}

	// Removed hardcoded literal values list - let grammar context determine field vs value

	// Skip obvious file extensions - these are typically values, not field names
	if strings.HasSuffix(strings.ToLower(text), ".csv") || strings.HasSuffix(strings.ToLower(text), ".txt") || strings.HasSuffix(strings.ToLower(text), ".log") {
		return false
	}

	// Skip obvious operators
	operators := []string{"+", "-", "*", "/", "=", "!", "<", ">", "(", ")", "[", "]", "{", "}", ",", ";", ":", " ", "\t", "\n"}
	for _, op := range operators {
		if strings.Contains(text, op) {
			return false
		}
	}

	// Field names typically contain letters, numbers, and underscores
	for _, r := range text {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.') {
			return false
		}
	}

	return true
}

// EnterBYOP handles "by" operations in stats, etc.
func (l *FieldDiscoveryListener) EnterBYOP(ctx *parser.BYOPContext) {
	// All fields in "by" clause are input fields
	for _, id := range ctx.AllId() {
		fieldName := id.GetText()
		// Only add legitimate field references, not concatenated text with "by" prefix
		if l.isLikelyFieldReference(fieldName) && !strings.HasPrefix(fieldName, "by") {
			l.addInputField(fieldName)
		}
	}
}

// EnterInitCommand handles commands that can appear at the beginning of a query (like | inputlookup)
func (l *FieldDiscoveryListener) EnterInitCommand(ctx *parser.InitCommandContext) {
	if ctx.INIT_COMMAND() != nil {
		command := ctx.INIT_COMMAND().GetText()
		switch strings.ToLower(command) {
		case "inputlookup":
			l.handleInputLookupInitCommand(ctx)
		case "datamodel":
			l.handleDataModelInitCommand(ctx)
		case "tstats":
			l.handleTstatsInitCommand(ctx)
		case "from":
			l.handleFromInitCommand(ctx)
		case "pivot":
			l.handlePivotInitCommand(ctx)
		}
	}
}

// EnterNextCommand handles different SPL commands
func (l *FieldDiscoveryListener) EnterNextCommand(ctx *parser.NextCommandContext) {
	if ctx.Command() == nil {
		return
	}

	command := ctx.Command().GetText()

	switch strings.ToLower(command) {
	case "lookup":
		l.handleLookupCommand(ctx)
	case "rename":
		l.handleRenameCommand(ctx)
	case "fields":
		l.handleFieldsCommand(ctx)
	case "append", "join", "multisearch":
		// Push context for subqueries
		l.pushDerivedContext()
	case "inputlookup":
		l.handleInputLookupCommand(ctx)
	case "datamodel":
		l.handleDataModelCommand(ctx)
	case "tstats":
		l.handleTstatsCommand(ctx)
	case "from":
		l.handleFromCommand(ctx)
	case "pivot":
		l.handlePivotCommand(ctx)
	}
}

// ExitNextCommand handles cleanup for commands that push context
func (l *FieldDiscoveryListener) ExitNextCommand(ctx *parser.NextCommandContext) {
	if ctx.Command() == nil {
		return
	}

	command := ctx.Command().GetText()
	if command == "append" || command == "join" || command == "multisearch" {
		l.popDerivedContext()
	}
}

// EnterSubquery and ExitSubquery handle subquery scoping
func (l *FieldDiscoveryListener) EnterSubquery(ctx *parser.SubqueryContext) {
	l.pushDerivedContext()
}

func (l *FieldDiscoveryListener) ExitSubquery(ctx *parser.SubqueryContext) {
	l.popDerivedContext()
}

// Command-specific handlers
func (l *FieldDiscoveryListener) handleLookupCommand(ctx *parser.NextCommandContext) {
	operations := ctx.AllOperation()
	if len(operations) == 0 {
		return
	}

	// First operation should be the lookup table name
	tableName := operations[0].GetText()
	l.addLookup(tableName)

	// Process the remaining operations to find field mappings
	inOutputSection := false

	for i := 1; i < len(operations); i++ {
		opText := operations[i].GetText()

		if strings.ToUpper(opText) == "OUTPUT" {
			inOutputSection = true
			continue
		}

		if inOutputSection {
			// Fields after OUTPUT are derived/created by the lookup
			l.markDerived(opText)
		} else {
			// Fields before OUTPUT are input fields used for matching
			// Skip obvious non-field tokens
			if l.isLikelyFieldReference(opText) {
				l.addInputField(opText)
			}
		}
	}
}

func (l *FieldDiscoveryListener) handleRenameCommand(ctx *parser.NextCommandContext) {
	// Parse rename operations to find "field AS newfield" patterns
	// In rename commands, the operations are typically "field AS newfield"
	operations := ctx.AllOperation()

	// Process operations in groups of 3: field, AS, newfield
	for i := 0; i < len(operations); i += 3 {
		if i+2 < len(operations) {
			originalField := operations[i].GetText()
			asKeyword := operations[i+1].GetText()
			newField := operations[i+2].GetText()

			// Verify it's actually a rename operation
			if strings.ToUpper(asKeyword) == "AS" {
				// Original field is an input field
				l.addInputField(originalField)
				// New field is derived
				l.markDerived(newField)
			}
		}
	}
}

func (l *FieldDiscoveryListener) handleFieldsCommand(ctx *parser.NextCommandContext) {
	// Fields in "fields" command are input field references
	for _, op := range ctx.AllOperation() {
		fieldName := op.GetText()
		l.addInputField(fieldName)
	}
}

func (l *FieldDiscoveryListener) handleInputLookupCommand(ctx *parser.NextCommandContext) {
	// First operation should be the lookup file name
	operations := ctx.AllOperation()
	if len(operations) > 0 {
		lookupFile := operations[0].GetText()
		// Clean up .csv extension if present
		lookupFile = strings.TrimSuffix(lookupFile, ".csv")
		l.addLookup(lookupFile)
	}
}

func (l *FieldDiscoveryListener) handleInputLookupInitCommand(ctx *parser.InitCommandContext) {
	// First operation should be the lookup file name
	operations := ctx.AllOperation()
	if len(operations) > 0 {
		lookupFile := operations[0].GetText()
		// Clean up .csv extension if present
		lookupFile = strings.TrimSuffix(lookupFile, ".csv")
		l.addLookup(lookupFile)
	}
}

func (l *FieldDiscoveryListener) handleDataModelInitCommand(ctx *parser.InitCommandContext) {
	// Expected format: | datamodel DataModel_Name Dataset_Name [search]
	operations := ctx.AllOperation()
	if len(operations) >= 2 {
		dataModelName := operations[0].GetText()
		datasetName := operations[1].GetText()

		// Add the datamodel
		l.addDataModel(dataModelName)

		// Add the dataset in DataModel.Dataset format
		datasetFullName := dataModelName + "." + datasetName
		l.addDataset(datasetFullName)

		// Process remaining operations for field references (skip "search" keyword)
		for i := 2; i < len(operations); i++ {
			opText := operations[i].GetText()
			if l.isLikelyFieldReference(opText) && !strings.EqualFold(opText, "search") {
				l.addInputField(opText)
			}
		}
	}
}

func (l *FieldDiscoveryListener) handleTstatsInitCommand(ctx *parser.InitCommandContext) {
	// Process operations to find datamodel references and field references
	operations := ctx.AllOperation()

	for _, op := range operations {
		opText := op.GetText()

		// Look for "from datamodel=ModelName" pattern
		if strings.Contains(strings.ToLower(opText), "datamodel=") {
			dataModelName := l.extractDataModelFromEquals(opText)
			if dataModelName != "" {
				l.addDataModel(dataModelName)
			}
		}

		// Look for "nodename=ObjectName" pattern
		if strings.Contains(strings.ToLower(opText), "nodename=") {
			objectName := l.extractValueAfterEquals(opText, "nodename")
			if objectName != "" {
				// We could track object names separately, but for now just note it
				// The object name will be used in field qualifications
			}
		}

		// Look for datamodel field references (Object.field or DataModel.Object.field)
		if strings.Contains(opText, ".") && l.isDataModelFieldReference(opText) {
			l.addInputField(opText)
		}

		// Look for fields inside function calls like sum(Web.bytes)
		l.extractFieldsFromFunctionCalls(opText)
	}
}

func (l *FieldDiscoveryListener) handleDataModelCommand(ctx *parser.NextCommandContext) {
	// Expected format: | datamodel DataModel_Name Dataset_Name [search]
	operations := ctx.AllOperation()
	if len(operations) >= 2 {
		dataModelName := operations[0].GetText()
		datasetName := operations[1].GetText()

		// Add the datamodel
		l.addDataModel(dataModelName)

		// Add the dataset in DataModel.Dataset format
		datasetFullName := dataModelName + "." + datasetName
		l.addDataset(datasetFullName)

		// Process remaining operations for field references (skip "search" keyword)
		for i := 2; i < len(operations); i++ {
			opText := operations[i].GetText()
			if l.isLikelyFieldReference(opText) && !strings.EqualFold(opText, "search") {
				l.addInputField(opText)
			}
		}
	}
}

func (l *FieldDiscoveryListener) handleTstatsCommand(ctx *parser.NextCommandContext) {
	// Process operations to find datamodel references and field references
	operations := ctx.AllOperation()

	for _, op := range operations {
		opText := op.GetText()

		// Look for "from datamodel=ModelName" pattern
		if strings.Contains(strings.ToLower(opText), "datamodel=") {
			dataModelName := l.extractDataModelFromEquals(opText)
			if dataModelName != "" {
				l.addDataModel(dataModelName)
			}
		}

		// Look for "nodename=ObjectName" pattern
		if strings.Contains(strings.ToLower(opText), "nodename=") {
			objectName := l.extractValueAfterEquals(opText, "nodename")
			if objectName != "" {
				// We could track object names separately, but for now just note it
				// The object name will be used in field qualifications
			}
		}

		// Look for datamodel field references (Object.field or DataModel.Object.field)
		if strings.Contains(opText, ".") && l.isDataModelFieldReference(opText) {
			l.addInputField(opText)
		}

		// Look for fields inside function calls like sum(Web.bytes)
		l.extractFieldsFromFunctionCalls(opText)
	}
}

// isDataModelFieldReference checks if a string looks like a datamodel field reference
func (l *FieldDiscoveryListener) isDataModelFieldReference(text string) bool {
	// Must contain a dot and look like Object.field or DataModel.Object.field
	if !strings.Contains(text, ".") {
		return false
	}

	parts := strings.Split(text, ".")
	// Support both Object.field (2 parts) and DataModel.Object.field (3 parts)
	if len(parts) != 2 && len(parts) != 3 {
		return false
	}

	// All parts should look like valid identifiers
	for _, part := range parts {
		if !l.isLikelyFieldReference(part) {
			return false
		}
	}

	return true
}

// EnterFieldUse handles general field references but needs to be context-aware
func (l *FieldDiscoveryListener) EnterFieldUse(ctx *parser.FieldUseContext) {
	if identifier := ctx.IDENTIFIER(); identifier != nil {
		fieldName := identifier.GetText()

		// Only add as input field if it's not in a value position
		// This is a basic heuristic - we could make this more sophisticated
		if l.isInFieldPosition(ctx) {
			l.addInputField(fieldName)
		}
	}
}

// EnterOUTPUTOP handles OUTPUT operations in lookup commands
func (l *FieldDiscoveryListener) EnterOUTPUTOP(ctx *parser.OUTPUTOPContext) {
	// Mark the entire concatenated context text as derived to prevent it from being added as a field
	l.markDerived(ctx.GetText())

	// In OUTPUT operations, the expression is the input field, the id is the output field
	if inputExpr := ctx.Expression(); inputExpr != nil {
		if value := inputExpr.Value(); value != nil {
			if id := value.Id(); id != nil {
				// This is an input field
				l.addInputField(id.GetText())
			}
		}
	}

	// The id after OUTPUT is a derived field
	if outputId := ctx.Id(); outputId != nil {
		l.markDerived(outputId.GetText())
	}
}

// EnterOUTPUTMULTIOP handles multiple OUTPUT operations
func (l *FieldDiscoveryListener) EnterOUTPUTMULTIOP(ctx *parser.OUTPUTMULTIOPContext) {
	// Mark the entire concatenated context text as derived to prevent it from being added as a field
	l.markDerived(ctx.GetText())

	// In multi OUTPUT operations, expressions are input fields, ids are output fields
	for _, expr := range ctx.AllExpression() {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				// This is an input field
				l.addInputField(id.GetText())
			}
		}
	}

	// All ids after OUTPUT are derived fields
	for _, outputId := range ctx.AllId() {
		l.markDerived(outputId.GetText())
	}
}

// EnterOUTPUTMULTIINOP handles multiple input/output operations
func (l *FieldDiscoveryListener) EnterOUTPUTMULTIINOP(ctx *parser.OUTPUTMULTIINOPContext) {
	// Mark the entire concatenated context text as derived to prevent it from being added as a field
	l.markDerived(ctx.GetText())

	// In multi input/output operations, expressions are input fields, ids are output fields
	for _, expr := range ctx.AllExpression() {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				// This is an input field
				l.addInputField(id.GetText())
			}
		}
	}

	// All ids after OUTPUT are derived fields
	for _, outputId := range ctx.AllId() {
		l.markDerived(outputId.GetText())
	}
}

// EnterRENAMEOP handles rename operations (expression AS id)
func (l *FieldDiscoveryListener) EnterRENAMEOP(ctx *parser.RENAMEOPContext) {
	// The expression is the original field (input field)
	if expr := ctx.Expression(); expr != nil {
		if value := expr.Value(); value != nil {
			if id := value.Id(); id != nil {
				// This is an input field
				l.addInputField(id.GetText())
			}
		}
	}

	// The id after AS is the new field name (derived field)
	if newId := ctx.Id(); newId != nil {
		l.markDerived(newId.GetText())
	}
}

// EnterEXPRESSIONOP handles general expression operations
func (l *FieldDiscoveryListener) EnterEXPRESSIONOP(ctx *parser.EXPRESSIONOPContext) {
	// Don't automatically add expression text as fields - let the specific context handlers
	// (EnterOUTPUTOP, EnterRENAMEOP, etc.) handle field extraction for their specific cases
	// This prevents treating operation text as field names
}

// isValidFieldReferenceInExpression determines if a field name from an expression context should be treated as an input field
func (l *FieldDiscoveryListener) isValidFieldReferenceInExpression(fieldName string) bool {
	// Skip concatenated strings that look like operations
	if strings.Contains(fieldName, "OUTPUT") || strings.Contains(fieldName, "AS") {
		return false
	}

	// Use the existing validation logic
	return l.isLikelyFieldReference(fieldName)
}

// isInFieldPosition determines if a field reference is in a position where it should be treated as an input field
func (l *FieldDiscoveryListener) isInFieldPosition(ctx *parser.FieldUseContext) bool {
	// Walk up the parent contexts to understand the position
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

// Additional helper methods for enhanced datamodel support

func (l *FieldDiscoveryListener) handleFromInitCommand(ctx *parser.InitCommandContext) {
	// Handle "from datamodel:DataModelName.ObjectName" syntax
	operations := ctx.AllOperation()

	for _, op := range operations {
		opText := op.GetText()

		// Look for "datamodel:ModelName.ObjectName" pattern
		if strings.Contains(strings.ToLower(opText), "datamodel:") {
			l.parseDataModelDatasetReference(opText)
		}
	}
}

func (l *FieldDiscoveryListener) handleFromCommand(ctx *parser.NextCommandContext) {
	// Handle "from datamodel:DataModelName.ObjectName" syntax
	operations := ctx.AllOperation()

	for _, op := range operations {
		opText := op.GetText()

		// Look for "datamodel:ModelName.ObjectName" pattern
		if strings.Contains(strings.ToLower(opText), "datamodel:") {
			l.parseDataModelDatasetReference(opText)
		}
	}
}

func (l *FieldDiscoveryListener) handlePivotInitCommand(ctx *parser.InitCommandContext) {
	// Handle "pivot DataModelName ObjectName ..." syntax
	operations := ctx.AllOperation()

	if len(operations) >= 2 {
		dataModelName := l.cleanQuotedName(operations[0].GetText())
		objectName := l.cleanQuotedName(operations[1].GetText())

		// Add the datamodel
		l.addDataModel(dataModelName)

		// Add the dataset in DataModel.Object format
		datasetFullName := dataModelName + "." + objectName
		l.addDataset(datasetFullName)
	}
}

func (l *FieldDiscoveryListener) handlePivotCommand(ctx *parser.NextCommandContext) {
	// Handle "pivot DataModelName ObjectName ..." syntax
	operations := ctx.AllOperation()

	if len(operations) >= 2 {
		dataModelName := l.cleanQuotedName(operations[0].GetText())
		objectName := l.cleanQuotedName(operations[1].GetText())

		// Add the datamodel
		l.addDataModel(dataModelName)

		// Add the dataset in DataModel.Object format
		datasetFullName := dataModelName + "." + objectName
		l.addDataset(datasetFullName)
	}
}

// Helper methods for parsing datamodel references

func (l *FieldDiscoveryListener) parseDataModelDatasetReference(text string) {
	// Parse "datamodel:DataModelName.ObjectName" or "datamodel:"DataModelName"."ObjectName""
	if colonIdx := strings.Index(strings.ToLower(text), "datamodel:"); colonIdx != -1 {
		datasetPart := text[colonIdx+10:] // Skip "datamodel:"

		// Handle different quoted formats
		// Case 1: "DataModel"."Object" - clean outer quotes first, then parse
		// Case 2: DataModel.Object - parse directly

		if strings.Contains(datasetPart, "\".\"") {
			// Handle "DataModel"."Object" format
			parts := strings.Split(datasetPart, "\".\"")
			if len(parts) == 2 {
				dataModelName := l.cleanQuotedName(parts[0])
				objectName := l.cleanQuotedName(parts[1])

				if dataModelName != "" && objectName != "" {
					l.addDataModel(dataModelName)
					l.addDataset(dataModelName + "." + objectName)
				}
			}
		} else if dotIdx := strings.Index(datasetPart, "."); dotIdx != -1 {
			// Handle regular DataModel.Object format
			dataModelName := strings.TrimSpace(datasetPart[:dotIdx])
			objectName := strings.TrimSpace(datasetPart[dotIdx+1:])

			dataModelName = l.cleanQuotedName(dataModelName)
			objectName = l.cleanQuotedName(objectName)

			if dataModelName != "" && objectName != "" {
				l.addDataModel(dataModelName)
				l.addDataset(dataModelName + "." + objectName)
			}
		}
	}
}

func (l *FieldDiscoveryListener) extractDataModelFromEquals(text string) string {
	if idx := strings.Index(strings.ToLower(text), "datamodel="); idx != -1 {
		value := text[idx+10:] // Skip "datamodel="
		return l.extractFirstToken(value)
	}
	return ""
}

func (l *FieldDiscoveryListener) extractValueAfterEquals(text string, key string) string {
	searchKey := strings.ToLower(key) + "="
	if idx := strings.Index(strings.ToLower(text), searchKey); idx != -1 {
		value := text[idx+len(searchKey):]
		return l.extractFirstToken(value)
	}
	return ""
}

func (l *FieldDiscoveryListener) extractFirstToken(text string) string {
	// Extract first token, handling quotes and stopping at spaces/operators
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Handle quoted strings
	if strings.HasPrefix(text, "\"") {
		if endIdx := strings.Index(text[1:], "\""); endIdx != -1 {
			return text[1 : endIdx+1]
		}
	}

	// Find first space, pipe, or other delimiter
	for i, r := range text {
		if r == ' ' || r == '|' || r == '\t' || r == '\n' {
			return text[:i]
		}
	}

	return text
}

func (l *FieldDiscoveryListener) cleanQuotedName(name string) string {
	name = strings.TrimSpace(name)
	// Remove surrounding quotes if present
	if len(name) >= 2 && strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
		return name[1 : len(name)-1]
	}
	return name
}

// extractFieldsFromFunctionCalls extracts field references from function calls like sum(Web.bytes)
func (l *FieldDiscoveryListener) extractFieldsFromFunctionCalls(text string) {
	// Look for patterns like function(field) where field contains dots
	if strings.Contains(text, "(") && strings.Contains(text, ")") {
		start := strings.Index(text, "(")
		end := strings.LastIndex(text, ")")
		if start != -1 && end != -1 && end > start {
			// Extract content between parentheses
			content := text[start+1 : end]

			// Check if this looks like a datamodel field reference OR a regular field
			if (strings.Contains(content, ".") && l.isDataModelFieldReference(content)) || l.isLikelyFieldReference(content) {
				l.addInputField(content)
			}
		}
	}
}

// EnterExpression handles expressions which may contain function calls like avg(bytes_in)
func (l *FieldDiscoveryListener) EnterExpression(ctx *parser.ExpressionContext) {
	// Check if this is a function call expression (function LPAREN ... RPAREN)
	if ctx.Function() != nil && len(ctx.AllExpression()) > 0 {
		// This is a function call - extract field references from the arguments
		for _, argExpr := range ctx.AllExpression() {
			l.extractFieldReferencesFromExpression(argExpr)
		}
	}
}
