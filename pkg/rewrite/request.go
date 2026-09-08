package rewrite

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func checkJSON(data []byte) error {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return inputError("%v", err)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	var raw json.RawMessage
	if err := d.Decode(&raw); err != nil {
		return inputError("invalid JSON: %v", err)
	}
	if err := d.Decode(&raw); err != io.EOF {
		return inputError("expected exactly one JSON value")
	}
	return nil
}

// object enforces duplicate, unknown, and required keys before typed decoding.
func object(data []byte, required []string, allowed ...string) (map[string]json.RawMessage, error) {
	if err := checkJSON(data); err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	start, _ := d.Token()
	if start != json.Delim('{') {
		return nil, inputError("expected an object")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		token, _ := d.Token() // checkJSON already established valid JSON syntax.
		key := token.(string)
		if _, exists := fields[key]; exists {
			return nil, inputError("duplicate property %q", key)
		}
		known := false
		for _, name := range allowed {
			known = known || name == key
		}
		if !known {
			return nil, inputError("unknown property %q", key)
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return nil, inputError("invalid property %q: %v", key, err)
		}
		fields[key] = raw
	}
	for _, key := range required {
		if _, exists := fields[key]; !exists {
			return nil, inputError("missing property %q", key)
		}
	}
	return fields, nil
}

func stringValue(raw []byte) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '"' {
		return "", inputError("expected a string")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", inputError("invalid string: %v", err)
	}
	return value, nil
}

func array(raw []byte) ([]json.RawMessage, error) {
	if err := checkJSON(raw); err != nil {
		return nil, err
	}
	raw = bytes.TrimSpace(raw)
	if raw[0] != '[' {
		return nil, inputError("expected an array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, inputError("invalid array: %v", err)
	}
	return values, nil
}

func decodeVersion(raw []byte) (int, error) {
	var version int
	if json.Unmarshal(raw, &version) != nil || version != 1 {
		return 0, inputError("schema_version must be integer 1")
	}
	return version, nil
}

func decodeMode(raw []byte) (Mode, error) {
	if raw == nil {
		return Preview, nil
	}
	mode, err := stringValue(raw)
	if err != nil || (mode != string(Preview) && mode != string(Apply)) {
		return "", inputError("mode must be preview or apply")
	}
	return Mode(mode), nil
}

// DecodeRequest checks the strict versioned wrapper and returns owned values.
func DecodeRequest(data []byte) (Request, error) {
	f, err := object(data, []string{"schema_version", "document", "rules"}, "schema_version", "mode", "document", "rules", "validation_target")
	if err != nil {
		return Request{}, err
	}
	version, mode, rules, target, err := decodeShared(f)
	if err != nil {
		return Request{}, err
	}
	documents, err := validation.DecodeDocuments(append(append([]byte{'['}, f["document"]...), ']'))
	if err != nil {
		return Request{}, inputError("document: %v", err)
	}
	return prepareRequest(Request{SchemaVersion: version, Mode: mode, Document: documents[0], Rules: rules, ValidationTarget: target})
}

// DecodeBatchRequest rejects the whole input if any document shape is invalid.
func DecodeBatchRequest(data []byte) (BatchRequest, error) {
	f, err := object(data, []string{"schema_version", "documents", "rules"}, "schema_version", "mode", "documents", "rules", "validation_target")
	if err != nil {
		return BatchRequest{}, err
	}
	version, mode, rules, target, err := decodeShared(f)
	if err != nil {
		return BatchRequest{}, err
	}
	documents, err := validation.DecodeDocuments(f["documents"])
	if err != nil {
		return BatchRequest{}, inputError("documents: %v", err)
	}
	return prepareBatchRequest(BatchRequest{SchemaVersion: version, Mode: mode, Documents: documents, Rules: rules, ValidationTarget: target})
}

func DecodeRuleSet(data []byte) (RuleSet, error) {
	f, err := object(data, []string{"schema_version", "rules"}, "schema_version", "rules")
	if err != nil {
		return RuleSet{}, err
	}
	version, _, rules, _, err := decodeShared(f)
	if err != nil {
		return RuleSet{}, err
	}
	rules, err = prepareRules(rules)
	if err != nil {
		return RuleSet{}, err
	}
	return RuleSet{SchemaVersion: version, Rules: rules}, nil
}

func decodeShared(f map[string]json.RawMessage) (int, Mode, []Rule, *ValidationTarget, error) {
	version, err := decodeVersion(f["schema_version"])
	if err != nil {
		return 0, "", nil, nil, err
	}
	mode, err := decodeMode(f["mode"])
	if err != nil {
		return 0, "", nil, nil, err
	}
	items, err := array(f["rules"])
	if err != nil {
		return 0, "", nil, nil, err
	}
	rules := make([]Rule, 0, len(items))
	for i, raw := range items {
		rule, err := decodeRule(raw)
		if err != nil {
			return 0, "", nil, nil, inputError("rule %d: %v", i, err)
		}
		rules = append(rules, rule)
	}
	var target *ValidationTarget
	if raw, exists := f["validation_target"]; exists {
		target, err = decodeTarget(raw)
		if err != nil {
			return 0, "", nil, nil, err
		}
	}
	return version, mode, rules, target, nil
}

func decodeRule(raw []byte) (Rule, error) {
	f, err := object(raw, []string{"id", "kind", "source", "target"}, "id", "kind", "source", "target", "when")
	if err != nil {
		return Rule{}, err
	}
	var r Rule
	r.ID, err = stringValue(f["id"])
	if err != nil {
		return Rule{}, err
	}
	r.Kind, err = stringValue(f["kind"])
	if err != nil {
		return Rule{}, err
	}
	r.Source, err = decodeIdentity(f["source"])
	if err != nil {
		return Rule{}, err
	}
	r.Target, err = decodeIdentity(f["target"])
	if err != nil {
		return Rule{}, err
	}
	if value, exists := f["when"]; exists {
		condition, err := decodeCondition(value)
		if err != nil {
			return Rule{}, err
		}
		r.When = &condition
	}
	return r, nil
}

func decodeIdentity(raw []byte) (Identity, error) {
	f, err := object(raw, nil, "name", "path")
	if err != nil {
		return Identity{}, err
	}
	if len(f) != 1 {
		return Identity{}, inputError("identity requires exactly one name or path")
	}
	if value, exists := f["name"]; exists {
		name, err := stringValue(value)
		if err != nil {
			return Identity{}, err
		}
		return Identity{Name: &name}, nil
	}
	items, err := array(f["path"])
	if err != nil {
		return Identity{}, err
	}
	path := make([]string, 0, len(items))
	for _, raw := range items {
		segment, err := stringValue(raw)
		if err != nil {
			return Identity{}, err
		}
		path = append(path, segment)
	}
	return Identity{Path: path}, nil
}

func decodeCondition(raw []byte) (Condition, error) {
	f, err := object(raw, nil, "all", "any", "fact", "kind", "identity", "operator", "value")
	if err != nil {
		return Condition{}, err
	}
	for _, key := range []string{"all", "any"} {
		if value, exists := f[key]; exists {
			if len(f) != 1 {
				return Condition{}, inputError("condition requires exactly one leaf or combinator")
			}
			items, err := array(value)
			if err != nil {
				return Condition{}, err
			}
			children := make([]Condition, 0, len(items))
			for _, item := range items {
				child, err := decodeCondition(item)
				if err != nil {
					return Condition{}, err
				}
				children = append(children, child)
			}
			if key == "all" {
				return Condition{All: children}, nil
			}
			return Condition{Any: children}, nil
		}
	}
	fact, err := stringValue(f["fact"])
	if err != nil {
		return Condition{}, err
	}
	keys := []string{"fact", "kind", "identity"}
	if fact == "literal" {
		keys = append(keys, "operator", "value")
	}
	f, err = object(raw, keys, keys...)
	if err != nil {
		return Condition{}, err
	}
	c := Condition{Fact: fact}
	c.Kind, err = stringValue(f["kind"])
	if err != nil {
		return Condition{}, err
	}
	identity, err := decodeIdentity(f["identity"])
	if err != nil {
		return Condition{}, err
	}
	c.Identity = &identity
	if fact == "literal" {
		c.Operator, err = stringValue(f["operator"])
		if err != nil {
			return Condition{}, err
		}
		c.Value = append(json.RawMessage(nil), f["value"]...)
	}
	return c, nil
}

// Preparation is shared by decoders and the later canonical rewrite operations.
// No parser, rewrite, or candidate validation runs at this request boundary.
func prepareRequest(r Request) (Request, error) {
	mode, rules, target, err := prepareShared(r.SchemaVersion, r.Mode, r.Rules, r.ValidationTarget)
	if err != nil {
		return Request{}, err
	}
	documents, err := prepareDocuments([]analysis.QueryDocument{r.Document})
	if err != nil {
		return Request{}, err
	}
	return Request{SchemaVersion: 1, Mode: mode, Document: documents[0], Rules: rules, ValidationTarget: target}, nil
}

func prepareBatchRequest(r BatchRequest) (BatchRequest, error) {
	mode, rules, target, err := prepareShared(r.SchemaVersion, r.Mode, r.Rules, r.ValidationTarget)
	if err != nil {
		return BatchRequest{}, err
	}
	documents, err := prepareDocuments(r.Documents)
	if err != nil {
		return BatchRequest{}, err
	}
	return BatchRequest{SchemaVersion: 1, Mode: mode, Documents: documents, Rules: rules, ValidationTarget: target}, nil
}

func prepareShared(version int, mode Mode, rules []Rule, target *ValidationTarget) (Mode, []Rule, *ValidationTarget, error) {
	if version != 1 {
		return "", nil, nil, inputError("schema_version must be integer 1")
	}
	if mode == "" {
		mode = Preview
	}
	if mode != Preview && mode != Apply {
		return "", nil, nil, inputError("mode must be preview or apply")
	}
	prepared, err := prepareRules(rules)
	if err != nil {
		return "", nil, nil, err
	}
	target, err = prepareTarget(target)
	if err != nil {
		return "", nil, nil, err
	}
	return mode, prepared, target, nil
}

func prepareDocuments(documents []analysis.QueryDocument) ([]analysis.QueryDocument, error) {
	for _, doc := range documents {
		for _, s := range []string{doc.Text, doc.Language, doc.Profile, doc.Version, doc.SourceID} {
			if !utf8.ValidString(s) {
				return nil, inputError("document strings must be valid UTF-8")
			}
		}
	}
	raw, err := json.Marshal(documents)
	if err != nil {
		return nil, inputError("documents: %v", err)
	}
	result, err := validation.DecodeDocuments(raw)
	if err != nil {
		return nil, inputError("documents: %v", err)
	}
	return result, nil
}

func validName(s string) bool { return utf8.ValidString(s) && strings.TrimSpace(s) != "" }

func prepareRules(rules []Rule) ([]Rule, error) {
	prepared := make([]Rule, 0, len(rules))
	seen := map[string]bool{}
	for i, r := range rules {
		if !validName(r.ID) || seen[r.ID] {
			return nil, inputError("rule %d: id must be unique, nonblank UTF-8", i)
		}
		seen[r.ID] = true
		switch r.Kind {
		case "field", "index", "source", "sourcetype", "lookup", "dataset", "data_model":
		default:
			return nil, inputError("rule %q: unsupported kind %q", r.ID, r.Kind)
		}
		var err error
		r.Source, err = prepareIdentity(r.Source, r.Kind)
		if err != nil {
			return nil, inputError("rule %q source: %v", r.ID, err)
		}
		r.Target, err = prepareIdentity(r.Target, r.Kind)
		if err != nil {
			return nil, inputError("rule %q target: %v", r.ID, err)
		}
		if r.When != nil {
			condition, err := prepareCondition(*r.When)
			if err != nil {
				return nil, inputError("rule %q condition: %v", r.ID, err)
			}
			r.When = &condition
		}
		prepared = append(prepared, r)
	}
	return prepared, nil
}

func prepareIdentity(id Identity, kind string) (Identity, error) {
	if (id.Name == nil) == (id.Path == nil) {
		return Identity{}, inputError("identity requires exactly one name or path")
	}
	if id.Name != nil {
		if !validName(*id.Name) {
			return Identity{}, inputError("name must be nonblank UTF-8")
		}
		name := *id.Name
		return Identity{Name: &name}, nil
	}
	if kind != "field" || len(id.Path) == 0 {
		return Identity{}, inputError("nonempty paths are field-only")
	}
	for _, segment := range id.Path {
		if !validName(segment) {
			return Identity{}, inputError("path segments must be nonblank UTF-8")
		}
	}
	return Identity{Path: append([]string{}, id.Path...)}, nil
}

func prepareCondition(c Condition) (Condition, error) {
	if c.All != nil || c.Any != nil {
		if (c.All != nil && c.Any != nil) || c.Fact != "" || c.Kind != "" || c.Identity != nil || c.Operator != "" || c.Value != nil {
			return Condition{}, inputError("condition requires exactly one leaf or combinator")
		}
		children := c.All
		if c.Any != nil {
			children = c.Any
		}
		if len(children) == 0 {
			return Condition{}, inputError("condition combinators must be nonempty")
		}
		prepared := make([]Condition, 0, len(children))
		for _, child := range children {
			p, err := prepareCondition(child)
			if err != nil {
				return Condition{}, err
			}
			prepared = append(prepared, p)
		}
		if c.All != nil {
			return Condition{All: prepared}, nil
		}
		return Condition{Any: prepared}, nil
	}
	switch c.Fact {
	case "literal":
		switch c.Kind {
		case "field", "index", "source", "sourcetype":
		default:
			return Condition{}, inputError("unsupported literal kind %q", c.Kind)
		}
		if c.Operator != "equals" && c.Operator != "contains" {
			return Condition{}, inputError("unsupported literal operator %q", c.Operator)
		}
		if err := checkJSON(c.Value); err != nil {
			return Condition{}, inputError("literal value: %v", err)
		}
		d := json.NewDecoder(bytes.NewReader(c.Value))
		d.UseNumber()
		var value any
		if err := d.Decode(&value); err != nil {
			return Condition{}, inputError("literal value: %v", err)
		}
		switch value.(type) {
		case nil, string, bool, json.Number:
		default:
			return Condition{}, inputError("literal value must be a scalar")
		}
		if _, ok := value.(string); c.Operator == "contains" && !ok {
			return Condition{}, inputError("contains requires a string value")
		}
		c.Value = append(json.RawMessage(nil), bytes.TrimSpace(c.Value)...)
	case "source_reference_present":
		if c.Kind != "field" || c.Operator != "" || c.Value != nil {
			return Condition{}, inputError("source_reference_present requires only a field identity")
		}
	default:
		return Condition{}, inputError("unsupported condition fact %q", c.Fact)
	}
	if c.Identity == nil {
		return Condition{}, inputError("condition identity is required")
	}
	identity, err := prepareIdentity(*c.Identity, c.Kind)
	if err != nil {
		return Condition{}, err
	}
	c.Identity = &identity
	return c, nil
}
