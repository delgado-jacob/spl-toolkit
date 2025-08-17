# Contributing Guide

Thank you for your interest in contributing to SPL Toolkit! This guide will help you get started with contributing to the project.

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and inclusive in all interactions.

## Getting Started

### Prerequisites

- Go 1.22+
- Python 3.8+
- Make
- Git

### Setting Up Development Environment

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/spl-toolkit.git
   cd spl-toolkit
   ```

2. **Setup Development Environment**
   ```bash
   make dev-setup
   ```

3. **Run Tests**
   ```bash
   make dev-test
   ```

4. **Build Everything**
   ```bash
   make build-all
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

Follow the project's coding standards and architecture principles:

- **Grammar-First**: Use ANTLR4 grammar for all SPL parsing
- **AST-Based**: Leverage Abstract Syntax Tree traversal
- **Context-Aware**: Maintain field derivation context
- **Test-Driven**: Write tests for new functionality

### 3. Testing

#### Run All Tests
```bash
make dev-test
```

#### Run Specific Test Suites
```bash
# Go tests only
make test

# Python tests only  
make python-test

# Performance benchmarks
make benchmark

# Integration tests
make integration-test
```

#### Test Coverage
```bash
make test-coverage
```

Target coverage: 80%+ for new code

### 4. Code Quality

#### Format Code
```bash
make fmt
```

#### Lint Code
```bash
make lint
```

#### Security Scan
```bash
make security
```

### 5. Documentation

Update documentation for any user-facing changes:

- Update relevant markdown files in `docs/`
- Add code examples for new features
- Update API documentation
- Add integration examples

### 6. Commit Guidelines

Use conventional commit format:

```
type(scope): subject

body

footer
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```bash
git commit -m "feat(mapper): add conditional rule support"
git commit -m "fix(parser): handle malformed SPL queries gracefully"
git commit -m "docs(api): add Python API examples"
```

### 7. Pull Request Process

1. **Ensure CI Passes**
   - All tests pass
   - Code coverage meets requirements
   - Linting passes
   - Documentation builds successfully

2. **Create Pull Request**
   - Use descriptive title
   - Include detailed description
   - Reference related issues
   - Add screenshots for UI changes

3. **Code Review**
   - Address reviewer feedback
   - Keep discussions constructive
   - Be responsive to comments

## Architecture Guidelines

### Grammar-First Development

All SPL processing must use the ANTLR4 grammar:

```go
// ✅ Good: Grammar-aware processing
func (l *FieldListener) EnterFieldExpression(ctx *FieldExpressionContext) {
    field := ctx.FIELD_NAME().GetText()
    l.fields = append(l.fields, field)
}

// ❌ Bad: Regex-based processing
func extractFields(query string) []string {
    re := regexp.MustCompile(`\b\w+\s*=`)
    // ... fragile pattern matching
}
```

### Context-Aware Processing

Maintain awareness of field derivation and scope:

```go
// ✅ Good: Context tracking
type FieldContext struct {
    inputFields   map[string]bool
    derivedFields map[string]FieldDerivation
    scopeStack    []*Scope
}

// ❌ Bad: Treating all fields the same
func getAllFields(query string) []string {
    // ... no distinction between input and derived fields
}
```

### Testing Best Practices

#### Grammar Testing

Test against the actual grammar, not hardcoded expectations:

```go
// ✅ Good: Grammar-based testing
func TestQueryParsing(t *testing.T) {
    query := "search src_ip=192.168.1.1 | eval new_field=src_ip+\"_test\""
    ast, err := parser.Parse(query)
    assert.NoError(t, err)
    assert.NotNil(t, ast)
    
    // Validate AST structure matches grammar expectations
    commands := ast.GetCommands()
    assert.Len(t, commands, 2)
    assert.Equal(t, "search", commands[0].GetName())
    assert.Equal(t, "eval", commands[1].GetName())
}

// ❌ Bad: Hardcoded pattern testing
func TestFieldExtraction(t *testing.T) {
    fields := extractFieldsWithRegex("search src_ip=192.168.1.1")
    assert.Equal(t, []string{"src_ip"}, fields)
    // Brittle test that doesn't validate grammar compliance
}
```

#### Test Structure

Organize tests by functionality:

```
pkg/mapper/
├── mapper_test.go              # Core mapper functionality
├── discovery_test.go           # Discovery engine tests
├── field_mapping_test.go       # Field mapping tests
├── parser_test.go             # Parser tests
├── testdata/                  # Test fixtures
│   ├── queries/              # Sample SPL queries
│   ├── configs/              # Test configurations
│   └── expected/             # Expected results
└── benchmarks/               # Performance benchmarks
    └── mapping_bench_test.go
```

## Project Structure

```
spl-toolkit/
├── cmd/                      # CLI application
├── pkg/                      # Go library packages
│   ├── mapper/              # Core mapping functionality
│   └── bindings/            # Language bindings
├── python/                   # Python bindings
├── grammar/                  # ANTLR4 grammar files
├── parser/                   # Generated parser code
├── docs/                     # Documentation
├── examples/                 # Usage examples
└── scripts/                  # Build and utility scripts
```

## Adding New Features

### 1. SPL Command Support

To add support for a new SPL command:

1. **Update Grammar** (if needed)
   ```antlr
   // In SPLParser.g4
   newCommand
       : NEW_COMMAND_TOKEN newCommandArgs
       ;
   ```

2. **Implement Listener**
   ```go
   func (l *DiscoveryListener) EnterNewCommand(ctx *NewCommandContext) {
       // Extract fields and resources
   }
   ```

3. **Add Mapping Support**
   ```go
   func (m *MappingListener) ExitNewCommand(ctx *NewCommandContext) {
       // Apply field mappings
   }
   ```

4. **Write Tests**
   ```go
   func TestNewCommandDiscovery(t *testing.T) {
       query := "search * | newcommand field1 field2"
       info, err := mapper.DiscoverQuery(query)
       assert.NoError(t, err)
       assert.Contains(t, info.InputFields, "field1")
   }
   ```

### 2. Mapping Rules

To add new conditional rule types:

1. **Define Condition Type**
   ```go
   type NewConditionType struct {
       Type     string `json:"type"`
       Operator string `json:"operator"`
       Value    string `json:"value"`
   }
   ```

2. **Implement Evaluator**
   ```go
   func (e *ConditionEvaluator) EvaluateNewCondition(
       condition *NewConditionType, 
       context *QueryContext,
   ) bool {
       // Evaluation logic
   }
   ```

3. **Update Schema**
   ```json
   {
     "type": "object",
     "properties": {
       "type": {"const": "new_condition_type"},
       "operator": {"enum": ["equals", "contains"]},
       "value": {"type": "string"}
     }
   }
   ```

### 3. Language Bindings

To add bindings for a new language:

1. **Create Binding Interface**
   ```go
   //export NewLanguageMapQuery  
   func NewLanguageMapQuery(query *C.char) *C.char {
       // CGO bridge implementation
   }
   ```

2. **Implement Language-Specific API**
   ```typescript
   // For TypeScript bindings
   export class SPLMapper {
       constructor(config: MapperConfig) { }
       mapQuery(query: string): string { }
   }
   ```

3. **Add Build Support**
   ```makefile
   # In Makefile
   new-language-build:
       # Build commands for new language
   ```

## Performance Guidelines

### Memory Management

- Use object pooling for frequently allocated objects
- Implement proper cleanup in defer statements
- Monitor memory usage with benchmarks

```go
func BenchmarkMapping(b *testing.B) {
    mapper := mapper.New()
    query := "search src_ip=192.168.1.1"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := mapper.MapQuery(query)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Concurrency

- Ensure thread safety without unnecessary locking
- Use read locks for configuration access
- Implement parser pools for concurrent usage

## Documentation Standards

### Code Documentation

```go
// MapQuery applies field mappings to the given SPL query based on the
// loaded configuration rules. It returns the transformed query string
// or an error if parsing or mapping fails.
//
// The mapping process:
//   1. Parses the query using ANTLR4 grammar
//   2. Evaluates conditional rules against query context
//   3. Applies applicable field mappings via token stream rewriting
//   4. Returns the modified query string
//
// Example:
//   mapper.LoadMappings([]byte(`[{"source": "src_ip", "target": "source_ip"}]`))
//   result, err := mapper.MapQuery("search src_ip=192.168.1.1")
//   // result: "search source_ip=192.168.1.1"
func (m *Mapper) MapQuery(query string) (string, error) {
    // Implementation
}
```

### API Documentation

- Include examples for all public functions
- Document error conditions
- Provide usage patterns
- Show integration examples

## Release Process

### Version Management

We use semantic versioning (SemVer):
- `MAJOR.MINOR.PATCH`
- Breaking changes increment MAJOR
- New features increment MINOR  
- Bug fixes increment PATCH

### Release Checklist

1. Update version in relevant files
2. Update CHANGELOG.md
3. Run full test suite
4. Build and test all artifacts
5. Create GitHub release
6. Deploy documentation
7. Announce release

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Code Reviews**: Detailed technical discussions

### Maintainer Response

- Issue acknowledgment: Within 48 hours
- Pull request review: Within 1 week
- Security issues: Within 24 hours

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Documentation credits

Thank you for contributing to SPL Toolkit! 🎉