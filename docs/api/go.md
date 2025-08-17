# Go API Reference

Complete reference for the SPL Toolkit Go library.

## Package Import

```go
import "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
```

## Core Types

### Mapper

The main interface for SPL field mapping operations.

```go
type Mapper struct {
    // Private fields
}

// New creates a new Mapper instance
func New() *Mapper

// NewWithConfig creates a new Mapper with configuration
func NewWithConfig(config *Config) (*Mapper, error)
```

### Config

Configuration structure for field mappings and rules.

```go
type Config struct {
    Version     string        `json:"version"`
    Name        string        `json:"name,omitempty"`
    Description string        `json:"description,omitempty"`
    Mappings    []FieldMapping `json:"mappings"`
    Rules       []MappingRule  `json:"rules,omitempty"`
}
```

### FieldMapping

Basic field mapping definition.

```go
type FieldMapping struct {
    Source  string `json:"source"`
    Target  string `json:"target"`
    Comment string `json:"comment,omitempty"`
}
```

### MappingRule

Conditional mapping rule with conditions and priority.

```go
type MappingRule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name,omitempty"`
    Description string        `json:"description,omitempty"`
    Conditions  []Condition   `json:"conditions"`
    Mappings    []FieldMapping `json:"mappings"`
    Priority    int           `json:"priority"`
    Enabled     bool          `json:"enabled"`
}
```

### Condition

Rule condition for conditional mapping.

```go
type Condition struct {
    Type     string      `json:"type"`
    Operator string      `json:"operator"`
    Value    string      `json:"value,omitempty"`
    Field    string      `json:"field,omitempty"`
    Children []Condition `json:"children,omitempty"`
}
```

### QueryInfo

Information discovered from SPL query analysis.

```go
type QueryInfo struct {
    InputFields   []string          `json:"input_fields"`
    DerivedFields map[string]string `json:"derived_fields"`
    Sources       []string          `json:"sources"`
    Sourcetypes   []string          `json:"sourcetypes"`
    Lookups       []string          `json:"lookups"`
    Macros        []string          `json:"macros"`
    Datamodels    []string          `json:"datamodels"`
    Commands      []string          `json:"commands"`
}
```

## Core Methods

### Mapping Operations

#### MapQuery

Applies field mappings to an SPL query.

```go
func (m *Mapper) MapQuery(query string) (string, error)
```

**Parameters:**
- `query`: SPL query string to transform

**Returns:**
- Mapped query string
- Error if parsing or mapping fails

**Example:**
```go
mapper := mapper.New()
mappings := `[{"source": "src_ip", "target": "source_ip"}]`
err := mapper.LoadMappings([]byte(mappings))
if err != nil {
    return err
}

result, err := mapper.MapQuery("search src_ip=192.168.1.1")
if err != nil {
    return err
}
// result: "search source_ip=192.168.1.1"
```

#### MapQueryWithContext

Applies field mappings with additional context for rule evaluation.

```go
func (m *Mapper) MapQueryWithContext(query string, context map[string]interface{}) (string, error)
```

**Parameters:**
- `query`: SPL query string to transform
- `context`: Additional context for rule evaluation

**Returns:**
- Mapped query string
- Error if parsing or mapping fails

**Example:**
```go
context := map[string]interface{}{
    "sourcetype": "access_combined",
    "index": "web_logs",
}

result, err := mapper.MapQueryWithContext(query, context)
```

#### MapQueries

Batch mapping of multiple queries.

```go
func (m *Mapper) MapQueries(queries []string) ([]string, error)
```

**Parameters:**
- `queries`: Array of SPL query strings

**Returns:**
- Array of mapped query strings
- Error if any query fails

### Discovery Operations

#### DiscoverQuery

Analyzes an SPL query to extract components and metadata.

```go
func (m *Mapper) DiscoverQuery(query string) (*QueryInfo, error)
```

**Parameters:**
- `query`: SPL query string to analyze

**Returns:**
- QueryInfo with discovered components
- Error if parsing fails

**Example:**
```go
info, err := mapper.DiscoverQuery("search sourcetype=apache clientip=192.168.1.1 | stats count by clientip")
if err != nil {
    return err
}

fmt.Printf("Input fields: %v\n", info.InputFields)
fmt.Printf("Sourcetypes: %v\n", info.Sourcetypes)
```

#### DiscoverFields

Extracts only field information from a query.

```go
func (m *Mapper) DiscoverFields(query string) ([]string, error)
```

**Parameters:**
- `query`: SPL query string

**Returns:**
- Array of input field names
- Error if parsing fails

### Configuration Management

#### LoadMappings

Loads field mappings from JSON bytes.

```go
func (m *Mapper) LoadMappings(data []byte) error
```

**Parameters:**
- `data`: JSON-encoded mapping configuration

**Returns:**
- Error if configuration is invalid

**Example:**
```go
configJSON := `{
  "version": "1.0",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"}
  ]
}`

err := mapper.LoadMappings([]byte(configJSON))
```

#### LoadConfig

Loads configuration from a Config struct.

```go
func (m *Mapper) LoadConfig(config *Config) error
```

**Parameters:**
- `config`: Configuration struct

**Returns:**
- Error if configuration is invalid

#### LoadConfigFromFile

Loads configuration from a file path.

```go
func (m *Mapper) LoadConfigFromFile(filepath string) error
```

**Parameters:**
- `filepath`: Path to JSON configuration file

**Returns:**
- Error if file cannot be read or configuration is invalid

#### ValidateConfig

Validates configuration without loading it.

```go
func ValidateConfig(data []byte) (*ValidationResult, error)
```

**Parameters:**
- `data`: JSON-encoded configuration

**Returns:**
- ValidationResult with details
- Error if validation fails

### Parser Operations

#### ParseQuery

Parses an SPL query into an AST without applying mappings.

```go
func (m *Mapper) ParseQuery(query string) (*AST, error)
```

**Parameters:**
- `query`: SPL query string

**Returns:**
- Abstract Syntax Tree representation
- Error if parsing fails

#### ValidateQuery

Validates SPL query syntax.

```go
func (m *Mapper) ValidateQuery(query string) error
```

**Parameters:**
- `query`: SPL query string

**Returns:**
- Error if query has syntax errors

## Advanced Features

### Custom Rule Evaluation

#### RegisterConditionEvaluator

Registers a custom condition evaluator.

```go
func (m *Mapper) RegisterConditionEvaluator(conditionType string, evaluator ConditionEvaluator) error
```

**Parameters:**
- `conditionType`: Name of the condition type
- `evaluator`: Custom evaluator implementation

### Performance Tuning

#### SetCacheSize

Sets the internal query cache size.

```go
func (m *Mapper) SetCacheSize(size int)
```

**Parameters:**
- `size`: Maximum number of cached queries (0 to disable)

#### SetParserPoolSize

Sets the parser pool size for concurrent operations.

```go
func (m *Mapper) SetParserPoolSize(size int)
```

**Parameters:**
- `size`: Number of parser instances in pool

### Statistics

#### GetStats

Returns mapping operation statistics.

```go
func (m *Mapper) GetStats() *MapperStats
```

**Returns:**
- Statistics about mapper performance

```go
type MapperStats struct {
    QueriesMapped    int64
    CacheHits        int64
    CacheMisses      int64
    AverageMapTime   time.Duration
    RulesEvaluated   int64
    ParseErrors      int64
}
```

## Error Types

### ParseError

SPL syntax parsing error.

```go
type ParseError struct {
    Line     int
    Position int
    Message  string
    Context  string
}

func (e *ParseError) Error() string
```

### MappingError

Field mapping application error.

```go
type MappingError struct {
    Field   string
    Rule    string
    Reason  string
}

func (e *MappingError) Error() string
```

### ConfigError

Configuration validation error.

```go
type ConfigError struct {
    Field   string
    Value   interface{}
    Reason  string
}

func (e *ConfigError) Error() string
```

## Utility Functions

### Field Utilities

#### IsValidFieldName

Checks if a string is a valid SPL field name.

```go
func IsValidFieldName(name string) bool
```

#### NormalizeFieldName

Normalizes field name according to SPL conventions.

```go
func NormalizeFieldName(name string) string
```

### Query Utilities

#### ExtractCommands

Extracts command names from an SPL query.

```go
func ExtractCommands(query string) ([]string, error)
```

#### SimplifyQuery

Removes comments and normalizes whitespace in a query.

```go
func SimplifyQuery(query string) string
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    // Create mapper
    m := mapper.New()
    
    // Load configuration
    config := `{
        "version": "1.0",
        "name": "Network Traffic Mapping",
        "mappings": [
            {"source": "src_ip", "target": "source_ip"},
            {"source": "dst_ip", "target": "destination_ip"},
            {"source": "src_port", "target": "source_port"}
        ],
        "rules": [
            {
                "id": "firewall_logs",
                "conditions": [
                    {"type": "sourcetype", "operator": "contains", "value": "firewall"}
                ],
                "mappings": [
                    {"source": "action", "target": "firewall_action"}
                ],
                "priority": 1,
                "enabled": true
            }
        ]
    }`
    
    err := m.LoadMappings([]byte(config))
    if err != nil {
        log.Fatal(err)
    }
    
    // Map a query
    query := "search src_ip=192.168.1.1 dst_port=80 | stats count by src_ip, dst_ip"
    mappedQuery, err := m.MapQuery(query)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Original: %s\n", query)
    fmt.Printf("Mapped:   %s\n", mappedQuery)
    
    // Discover query information
    info, err := m.DiscoverQuery(mappedQuery)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Input fields: %v\n", info.InputFields)
    fmt.Printf("Commands: %v\n", info.Commands)
    
    // Get statistics
    stats := m.GetStats()
    fmt.Printf("Queries mapped: %d\n", stats.QueriesMapped)
    fmt.Printf("Cache hits: %d\n", stats.CacheHits)
}
```

## Thread Safety

The Mapper is thread-safe for read operations after configuration is loaded:

```go
mapper := mapper.New()
mapper.LoadMappings(config) // Not thread-safe

// After configuration, safe for concurrent use
go func() {
    result1, _ := mapper.MapQuery(query1)
}()

go func() {
    result2, _ := mapper.MapQuery(query2)
}()
```

## Performance Considerations

- **Parser Pooling**: Reuses parser instances for better performance
- **Query Caching**: Caches parsed ASTs to avoid re-parsing
- **Memory Management**: Uses object pools to reduce garbage collection
- **Batch Operations**: Use `MapQueries` for processing multiple queries

## Best Practices

1. **Reuse Mapper Instances**: Create once, use many times
2. **Load Configuration Once**: Avoid reloading configuration frequently
3. **Use Context**: Provide context for better rule evaluation
4. **Handle Errors**: Always check for parsing and mapping errors
5. **Monitor Performance**: Use `GetStats()` to track performance metrics