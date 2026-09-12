---
title: "Go API"
layout: page
---

# Go API

For legacy mapping, import `github.com/delgado-jacob/spl-toolkit/pkg/mapper`. Its exported operations include:

```go
func New() *Mapper
func NewWithConfig(*MappingConfig) *Mapper
func (*Mapper) LoadMappings([]byte) error
func (*Mapper) MapQuery(string) (string, error)
func (*Mapper) MapQueryWithContext(string, map[string]interface{}) (string, error)
func (*Mapper) DiscoverQuery(string) (*QueryInfo, error)
func (*Mapper) ValidateQuery(string) error
func (*Parser) Parse(string) (*ASTNode, error)
```

Runnable mapping example:

```go
package main

import (
    "fmt"
    "log"

    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    m := mapper.New()
    if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil {
        log.Fatal(err)
    }
    result, err := m.MapQuery("search src_ip=1")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}
```

The checked-in copy is `examples/go/basic/main.go` from the repository root.

`NewWithConfig` preserves the existing constructor signature. If its configuration is invalid, mapping methods return the configuration error. Discovery remains available because it does not consume mapping configuration.

`QueryInfo` contains `DataModels`, `Datasets`, `Lookups`, `Macros`, `Sources`, `SourceTypes`, and `InputFields`. Input fields are flat and are not a lineage model. Macro-only recovery can return only `Macros` after unsupported macro syntax prevents normal parsing.

Legacy mappings replace supported field tokens and preserve surrounding text byte-for-byte. This is not a general semantic rewrite guarantee. Legacy mapping has no batch operation, custom-command registration API or public performance cache contract.

For canonical structured analysis and offline field/schema validation, use `pkg/analysis` and `pkg/validation`; see the [structured API](../API.md). Safe rewriting uses `pkg/rewrite`:

```go
func DecodeRequest([]byte) (Request, error)
func DecodeBatchRequest([]byte) (BatchRequest, error)
func Rewrite(Request) (*Result, error)
func RewriteBatch(BatchRequest) (*BatchResult, error)
```

Requests carry explicit rules and preview/apply mode independently of legacy mapper configuration. Query findings return reports; malformed requests return errors. The [rewrite contract](../rewrite.md) defines original/candidate/returned text, the single commit gate, ordered batches, supported identities and strict optional destination targets.
