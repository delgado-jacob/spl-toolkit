---
title: "Architecture Overview"
layout: page
---

# Architecture Overview

SPL Toolkit is built on a **Grammar-First Architecture** that ensures robust and accurate SPL query processing. This document provides a deep dive into the system's design principles, components, and implementation details.

## Design Principles

### 1. Grammar-First Architecture
- **ANTLR4 Foundation**: All SPL parsing uses the complete ANTLR4 grammar definition
- **No Regex Parsing**: Avoids fragile pattern matching approaches
- **AST-Based Processing**: Leverages Abstract Syntax Tree traversal for language-aware analysis
- **Grammar Evolution**: Supports SPL language extensions without breaking existing functionality

### 2. Context-Aware Processing
- **Field Classification**: Distinguishes input fields from derived fields
- **Scope Tracking**: Maintains context stacks for subqueries and joins
- **Hierarchical Analysis**: Handles nested operations and complex expressions
- **Semantic Understanding**: Respects SPL semantics in all transformations

### 3. Performance-Oriented Design
- **Zero-Copy Operations**: Minimizes memory allocations where possible
- **Efficient AST Traversal**: Uses visitor and listener patterns optimally
- **Concurrent Safety**: Thread-safe operations after initialization
- **Memory Management**: Careful resource handling for production workloads

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         SPL Toolkit                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐    │
│  │ Python API  │    │   Go Library │    │  CLI Interface  │    │
│  │             │    │              │    │                 │    │
│  │ ┌─────────┐ │    │ ┌──────────┐ │    │ ┌─────────────┐ │    │
│  │ │ Mapper  │ │◄──►│ │  Mapper  │ │◄──►│ │   Commands  │ │    │
│  │ │ Config  │ │    │ │  Config  │ │    │ │   Args      │ │    │
│  │ │ Query   │ │    │ │  Query   │ │    │ │   Output    │ │    │
│  │ └─────────┘ │    │ └──────────┘ │    │ └─────────────┘ │    │
│  └─────────────┘    └──────────────┘    └─────────────────┘    │
│         │                   │                     │             │
│         ▼                   ▼                     ▼             │
├─────────────────────────────────────────────────────────────────┤
│                      Core Go Library                            │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐  │
│  │ Mapping Engine  │  │Discovery Engine │  │ Config Manager │  │
│  │                 │  │                 │  │                │  │
│  │ • Rule Engine   │  │ • Field Finder  │  │ • Validation   │  │
│  │ • Token Rewrite │  │ • Resource Ext. │  │ • Rule Loading │  │
│  │ • Context Track │  │ • AST Analysis  │  │ • Schema Check │  │
│  └─────────────────┘  └─────────────────┘  └────────────────┘  │
│         │                       │                   │           │
│         └───────────────────────┼───────────────────┘           │
│                                 ▼                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                ANTLR4 Parser Core                      │   │
│  │                                                         │   │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │   │
│  │ │SPL Grammar  │ │ Lexer/      │ │ AST Builder     │   │   │
│  │ │Definition   │ │ Parser      │ │ & Visitors      │   │   │
│  │ └─────────────┘ └─────────────┘ └─────────────────┘   │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. ANTLR4 Parser Core

**SPL Grammar Definition**
- Complete SPL language specification in ANTLR4 format
- Located in `grammar/SPLLexer.g4` and `grammar/SPLParser.g4`
- Covers all SPL commands, functions, operators, and syntax constructs
- Extensible for custom SPL additions

**Lexer/Parser Generation**
```go
// Generated from ANTLR4 grammar
type SPLLexer struct {
    *antlr.BaseLexer
    // Tokenization logic
}

type SPLParser struct {
    *antlr.BaseParser
    // Parsing logic with grammar rules
}
```

**AST Builder & Visitors**
```go
// Base visitor for AST traversal
type BaseSPLParserVisitor struct {
    antlr.ParseTreeVisitor
}

// Base listener for event-driven parsing
type BaseSPLParserListener struct {
    antlr.ParseTreeListener
}
```

### 2. Mapping Engine

**Rule Engine**
```go
type RuleEngine struct {
    rules     []MappingRule
    evaluator *ConditionEvaluator
    priority  *PriorityManager
}

func (r *RuleEngine) EvaluateRules(query string, context Context) []MappingRule {
    // Evaluate conditions and return applicable rules
}
```

**Token Stream Rewriting**
```go
type TokenRewriter struct {
    stream   *antlr.CommonTokenStream
    mappings map[string]string
    context  *FieldContext
}

func (tr *TokenRewriter) ApplyMappings(mappings []FieldMapping) error {
    // Rewrite token stream while preserving syntax
}
```

**Context Tracking**
```go
type FieldContext struct {
    inputFields   map[string]bool
    derivedFields map[string]FieldDerivation
    scopeStack    []*Scope
}

type FieldDerivation struct {
    Command   string   // eval, rename, lookup, etc.
    Expression string  // original derivation expression
    Dependencies []string // fields this derived field depends on
}
```

### 3. Discovery Engine

**Field Discovery Listener**
```go
type FieldDiscoveryListener struct {
    *BaseSPLParserListener
    
    inputFields   []string
    derivedFields map[string]FieldDerivation
    resources     *ResourceInfo
    context       *DiscoveryContext
}

func (l *FieldDiscoveryListener) EnterFieldExpression(ctx *FieldExpressionContext) {
    // Extract field references from expressions
}
```

**Resource Extraction**
```go
type ResourceInfo struct {
    Sources      []string
    Sourcetypes  []string
    Lookups      []string
    Macros       []string
    Datamodels   []string
    Commands     []string
}

func (di *DiscoveryEngine) ExtractResources(query string) (*ResourceInfo, error) {
    // Use AST listeners to extract all resources
}
```

**Expression Analysis**
```go
type ExpressionAnalyzer struct {
    parser    *SPLParser
    evaluator *ExpressionEvaluator
}

func (ea *ExpressionAnalyzer) AnalyzeExpression(expr string) (*ExpressionInfo, error) {
    // Parse and analyze complex expressions for field references
}
```

### 4. Configuration Manager

**Schema Validation**
```go
type ConfigValidator struct {
    schema   *JSONSchema
    ruleset  *ValidationRules
}

func (cv *ConfigValidator) ValidateConfig(config []byte) (*ValidationResult, error) {
    // Validate JSON structure and semantic rules
}
```

**Rule Loading**
```go
type RuleLoader struct {
    parser    *JSONParser
    validator *ConfigValidator
    cache     *RuleCache
}

func (rl *RuleLoader) LoadRules(configData []byte) ([]MappingRule, error) {
    // Parse, validate, and cache mapping rules
}
```

## Processing Pipeline

### 1. Query Parsing

```
SPL Query String
       ↓
  ANTLR4 Lexer (Tokenization)
       ↓
  ANTLR4 Parser (AST Generation)
       ↓
   AST Tree Structure
```

### 2. Discovery Phase

```
AST Tree
    ↓
Field Discovery Listener
    ↓
Resource Extraction Visitor
    ↓
Context Analysis
    ↓
Discovery Result
```

### 3. Mapping Phase

```
AST Tree + Mapping Rules
         ↓
    Rule Evaluation
         ↓
  Context Assessment
         ↓
   Token Stream Rewriting
         ↓
    Modified Query
```

## Memory Management

### Go Memory Layout

```go
type Mapper struct {
    parser     *SPLParser        // ~500KB
    rules      []MappingRule     // ~10KB per rule
    cache      *QueryCache       // Configurable size
    listeners  []ParseListener   // ~50KB per listener
    context    *GlobalContext    // ~100KB
}
```

### Memory Optimization Strategies

1. **Object Pooling**: Reuse parser instances
2. **Lazy Loading**: Load rules on demand
3. **Query Caching**: Cache parsed ASTs
4. **String Interning**: Deduplicate field names

```go
type ObjectPool struct {
    parsers chan *SPLParser
    maxSize int
}

func (p *ObjectPool) Get() *SPLParser {
    select {
    case parser := <-p.parsers:
        return parser
    default:
        return createNewParser()
    }
}
```

## Concurrency Model

### Thread Safety

```go
type ConcurrentMapper struct {
    mu       sync.RWMutex
    config   *Config          // Protected by mutex
    parsers  *sync.Pool       // Thread-safe parser pool
    cache    *sync.Map        // Thread-safe cache
}

func (cm *ConcurrentMapper) MapQuery(query string) (string, error) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    parser := cm.parsers.Get().(*SPLParser)
    defer cm.parsers.Put(parser)
    
    return cm.mapWithParser(parser, query)
}
```

### Python Bindings

```go
//export MapQueryPython
func MapQueryPython(query *C.char, config *C.char) *C.char {
    // CGO bridge function
    queryStr := C.GoString(query)
    configStr := C.GoString(config)
    
    result, err := mapperInstance.MapQuery(queryStr)
    if err != nil {
        return C.CString(fmt.Sprintf("ERROR: %s", err.Error()))
    }
    
    return C.CString(result)
}
```

## Performance Characteristics

### Parsing Performance

| Query Complexity | Parse Time | Memory Usage |
|------------------|------------|--------------|
| Simple search    | ~50μs      | ~100KB       |
| Medium complexity| ~200μs     | ~500KB       |
| Complex query    | ~1ms       | ~2MB         |

### Mapping Performance

| Rule Count | Evaluation Time | Memory Overhead |
|------------|----------------|-----------------|
| 10 rules   | ~10μs          | ~100KB          |
| 100 rules  | ~50μs          | ~1MB            |
| 1000 rules | ~200μs         | ~10MB           |

### Scalability Metrics

- **Throughput**: 10,000+ queries/second (Go API)
- **Latency**: P95 < 5ms for complex queries
- **Memory**: Linear scaling with rule count
- **Concurrent Performance**: Near-linear scaling up to CPU core count

## Error Handling Strategy

### Error Types

```go
type SPLError interface {
    error
    Code() string
    Context() map[string]interface{}
}

type ParseError struct {
    Line     int
    Position int
    Message  string
    Context  string
}

type MappingError struct {
    Field   string
    Rule    string
    Reason  string
}
```

### Error Recovery

```go
type ErrorRecovery struct {
    strategy RecoveryStrategy
    fallback FallbackHandler
}

type RecoveryStrategy int

const (
    SkipInvalidFields RecoveryStrategy = iota
    UseOriginalQuery
    ReturnError
)
```

## Testing Architecture

### Test Structure

```
tests/
├── unit/                   # Unit tests
│   ├── parser/            # Parser component tests
│   ├── mapper/            # Mapping engine tests
│   └── discovery/         # Discovery engine tests
├── integration/           # Integration tests
│   ├── scenarios/         # End-to-end scenarios
│   └── performance/       # Performance benchmarks
└── fixtures/             # Test data
    ├── queries/          # Sample SPL queries
    ├── configs/          # Test configurations
    └── expected/         # Expected results
```

### Grammar Testing Strategy

```go
func TestGrammarCoverage(t *testing.T) {
    testCases := []struct {
        name     string
        query    string
        expected *AST
    }{
        {"basic_search", "search src_ip=192.168.1.1", expectedAST1},
        {"complex_eval", "search * | eval new_field=if(field1>100, \"high\", \"low\")", expectedAST2},
        // ... comprehensive test cases
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ast, err := parser.Parse(tc.query)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, ast)
        })
    }
}
```

## Extensibility Points

### Custom Commands

```go
type CustomCommand interface {
    Name() string
    Parse(ctx *ParseContext) (*CommandNode, error)
    Map(node *CommandNode, mappings []FieldMapping) error
}

func (m *Mapper) RegisterCommand(cmd CustomCommand) {
    m.customCommands[cmd.Name()] = cmd
}
```

### Custom Functions

```go
type CustomFunction interface {
    Name() string
    Signature() FunctionSignature
    Evaluate(args []interface{}) (interface{}, error)
}
```

### Grammar Extensions

```antlr
// In SPLParser.g4 - extensible grammar points
customCommand
    : CUSTOM_COMMAND_TOKEN customCommandArgs
    ;

customCommandArgs
    : customArg (COMMA customArg)*
    ;
```

## Security Considerations

### Input Validation

```go
type SecurityValidator struct {
    maxQueryLength int
    maxRuleCount   int
    allowedCommands map[string]bool
}

func (sv *SecurityValidator) ValidateQuery(query string) error {
    if len(query) > sv.maxQueryLength {
        return ErrQueryTooLong
    }
    
    // Check for potentially dangerous commands
    ast, _ := parser.Parse(query)
    for _, cmd := range ast.Commands {
        if !sv.allowedCommands[cmd.Name] {
            return ErrDisallowedCommand
        }
    }
    
    return nil
}
```

### Resource Limits

```go
type ResourceLimits struct {
    MaxParseTime    time.Duration
    MaxMemoryUsage  int64
    MaxCacheSize    int
    MaxRuleDepth    int
}
```

## Future Architecture Considerations

### Phase 2: Enhanced Features
- **Advanced Rule Engine**: Support for complex conditional logic
- **DataModel Mapping**: Full datamodel structure transformation
- **Query Optimization**: AST-level query optimization

### Phase 3: Query Translation
- **Dual-Mode Queries**: Support both raw and datamodel modes
- **Translation Engine**: Convert between search modes
- **Optimization Hints**: Query performance suggestions

### Phase 4: Machine Learning
- **Auto-mapping Detection**: Learn mappings from data patterns
- **Query Classification**: Automatic rule application
- **Performance Prediction**: Query execution time estimation

This architecture provides a solid foundation for current functionality while maintaining extensibility for future enhancements.