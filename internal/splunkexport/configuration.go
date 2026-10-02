package splunkexport

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// ConfigurationResult contains detached, allowlisted knowledge-object evidence
// and all twelve knowledge collections. Coverage describes identity enumeration,
// independently of definition diagnostics and offline closure expansion. Err is
// a coded fatal artifact-budget error: the assembler must stop export when nonnil.
// Acquisition failures are normal coverage gaps, with no raw response retained.
// The assembler must also enforce the combined budget with inventory and headers.
type ConfigurationResult struct {
	Objects     []environment.Object
	Collections []environment.Collection
	Diagnostics []Diagnostic
	Err         error
}

type configurationAdapter struct {
	kind, family, namePrefix string
	unpaged                  bool
}

var configurationAdapters = []configurationAdapter{
	{kind: "macro", family: "configs/conf-macros"},
	{kind: "saved_search", family: "saved/searches"},
	{kind: "event_type", family: "saved/eventtypes"},
	{kind: "lookup", family: "data/transforms/lookups"},
	{kind: "data_model", family: "datamodel/model"},
	{kind: "tag", family: "search/tags", unpaged: true},
	{kind: "calculated_field", family: "data/props/calcfields"},
	// Producer convention: these distinct endpoint families have separate typed
	// names, props:<entry.name> and transforms:<entry.name>, even when names match.
	{kind: "field_extraction", family: "data/props/extractions", namePrefix: "props:"},
	{kind: "field_extraction", family: "data/transforms/extractions", namePrefix: "transforms:"},
}

// Decode only fields with a supported mapping. RawMessage here distinguishes
// absent/null/malformed advertised fields; arbitrary content fields are ignored.
type configurationContent struct {
	ACL        json.RawMessage `json:"eai:acl"`
	Definition json.RawMessage `json:"definition"`
	Arguments  json.RawMessage `json:"args"`
	EvalBased  json.RawMessage `json:"iseval"`
	Validation json.RawMessage `json:"validation"`
	Search     json.RawMessage `json:"search"`
}
type configurationEntryValue struct {
	Name    json.RawMessage      `json:"name"`
	ACL     json.RawMessage      `json:"acl"`
	Content configurationContent `json:"content"`
}
type configurationPage struct {
	Entries []configurationEntryValue `json:"entry"`
	Paging  *struct {
		Total   *int `json:"total"`
		Offset  *int `json:"offset"`
		PerPage *int `json:"perPage"`
	} `json:"paging"`
	Messages []struct {
		Type string `json:"type"`
	} `json:"messages"`
}
type configurationContext struct{ owner, app, sharing string }

func configurationString(raw json.RawMessage) (string, bool) {
	var value string
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	return value, true
}
func configurationContextValue(raw json.RawMessage) (configurationContext, bool, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return configurationContext{}, false, true
	}
	var acl struct {
		Owner   json.RawMessage `json:"owner"`
		App     json.RawMessage `json:"app"`
		Sharing json.RawMessage `json:"sharing"`
	}
	if json.Unmarshal(raw, &acl) != nil {
		return configurationContext{}, true, false
	}
	var value configurationContext
	for _, field := range []struct {
		raw    json.RawMessage
		target *string
	}{{acl.Owner, &value.owner}, {acl.App, &value.app}, {acl.Sharing, &value.sharing}} {
		if len(field.raw) == 0 || string(field.raw) == "null" {
			continue
		}
		text, ok := configurationString(field.raw)
		if !ok {
			return configurationContext{}, true, false
		}
		*field.target = text
	}
	return value, true, true
}
func configurationACL(entry configurationEntryValue) (configurationContext, string) {
	primary, present, valid := configurationContextValue(entry.ACL)
	nested, nestedPresent, nestedValid := configurationContextValue(entry.Content.ACL)
	if !valid || !nestedValid {
		return configurationContext{}, "context_invalid"
	}
	if present && nestedPresent && primary != nested {
		return configurationContext{}, "context_conflict"
	}
	if !present {
		primary = nested
	}
	if strings.TrimSpace(primary.owner) == "" || strings.TrimSpace(primary.app) == "" {
		return primary, "context_unavailable"
	}
	return primary, ""
}
func configurationNamespace(ctx configurationContext) string {
	if strings.TrimSpace(ctx.owner) == "" || strings.TrimSpace(ctx.app) == "" {
		return ""
	}
	return url.PathEscape(ctx.owner) + "/" + url.PathEscape(ctx.app)
}
func configurationSelected(selector environment.Selector, value string) bool {
	if selector.All != nil && *selector.All {
		return true
	}
	for _, selected := range selector.Values {
		if value == selected {
			return true
		}
	}
	return false
}
func configurationScopeMatches(scope environment.CaptureScope, ctx configurationContext) bool {
	return configurationSelected(scope.Namespace, configurationNamespace(ctx)) && configurationSelected(scope.App, ctx.app) && configurationSelected(scope.Owner, ctx.owner)
}
func configurationRequestContexts(scope environment.CaptureScope) []configurationContext {
	values := func(selector environment.Selector) []string {
		if selector.All != nil && *selector.All {
			return []string{"-"}
		}
		out := append([]string{}, selector.Values...)
		sort.Strings(out)
		return out
	}
	contexts := []configurationContext{}
	for _, owner := range values(scope.Owner) {
		for _, app := range values(scope.App) {
			contexts = append(contexts, configurationContext{owner: owner, app: app})
		}
	}
	return contexts
}
func configurationPath(ctx configurationContext, family string) string {
	return "/servicesNS/" + url.PathEscape(ctx.owner) + "/" + url.PathEscape(ctx.app) + "/" + family
}
func configurationDiagnostic(code, kind string) Diagnostic {
	return Diagnostic{Code: code, Severity: "warning", Kind: kind, Message: "Configuration acquisition did not establish supported evidence for the declared scope or definition."}
}
func configurationPageValid(page configurationPage, offset, total int) bool {
	p := page.Paging
	if page.Entries == nil || p == nil || p.Total == nil || p.Offset == nil || p.PerPage == nil || *p.Total < 0 || *p.Offset != offset || *p.PerPage <= 0 || *p.PerPage > configurationPageSize || offset > *p.Total || total >= 0 && *p.Total != total {
		return false
	}
	remaining := *p.Total - offset
	expected := *p.PerPage
	if remaining < expected {
		expected = remaining
	}
	return len(page.Entries) == expected
}
func configurationPageWarning(page configurationPage) bool {
	for _, message := range page.Messages {
		if message.Type != "INFO" && message.Type != "DEBUG" {
			return true
		}
	}
	return false
}
func configurationNeedsDetail(kind, name string, content configurationContent) bool {
	switch kind {
	case "macro":
		// A zero-argument stanza needs neither args nor optional iseval.
		// An advertised invalid args value cannot be repaired by inheritance.
		base, arity, _, valid := configurationMacroIdentity(name, content.Arguments)
		if base == "" || !valid && len(content.Arguments) > 0 {
			return false
		}
		definition, ok := configurationString(content.Definition)
		return arity > 0 && len(content.Arguments) == 0 || !ok || strings.TrimSpace(definition) == ""
	case "saved_search", "event_type":
		search, ok := configurationString(content.Search)
		return !ok || strings.TrimSpace(search) == ""
	}
	return false
}

// Detail is fetched only from the same fixed family and a locally escaped name,
// never from a response URL. Returned context must agree with the list evidence.
func (c *Client) configurationDetail(ctx context.Context, adapter configurationAdapter, requestContext, objectContext configurationContext, name string) (configurationEntryValue, bool) {
	detailContext := requestContext
	if configurationNamespace(objectContext) != "" {
		detailContext = objectContext
	}
	var detail configurationPage
	if c.getJSON(ctx, configurationPath(detailContext, adapter.family)+"/"+url.PathEscape(name), nil, &detail) != nil || len(detail.Entries) != 1 || configurationPageWarning(detail) {
		return configurationEntryValue{}, false
	}
	value := detail.Entries[0]
	detailName, ok := identityName(value.Name)
	if !ok || detailName != name {
		return configurationEntryValue{}, false
	}
	actual, code := configurationACL(value)
	if code != "" || actual != objectContext {
		return configurationEntryValue{}, false
	}
	return value, true
}

// Only supported overlapping properties participate in detail agreement. An
// inherited detail response cannot silently replace an already observed body.
func configurationMergeContent(kind, name string, list, detail configurationContent) (configurationContent, bool) {
	type field struct {
		old   json.RawMessage
		newer *json.RawMessage
		query bool
	}
	fields := []field{}
	if kind == "macro" {
		fields = append(fields, field{list.Definition, &detail.Definition, true}, field{list.Validation, &detail.Validation, false})
		if len(list.Arguments) > 0 && string(list.Arguments) != "null" {
			_, _, oldArgs, oldOK := configurationMacroIdentity(name, list.Arguments)
			_, _, newArgs, newOK := configurationMacroIdentity(name, detail.Arguments)
			if oldOK && newOK && inventoryHash([]any{oldArgs}) != inventoryHash([]any{newArgs}) {
				return list, true
			}
			detail.Arguments = list.Arguments
		}
		oldEval, oldOK := configurationBoolean(list.EvalBased)
		newEval, newOK := configurationBoolean(detail.EvalBased)
		if oldOK && newOK && oldEval != nil && newEval != nil && *oldEval != *newEval {
			return list, true
		}
		if len(list.EvalBased) > 0 {
			detail.EvalBased = list.EvalBased
		}
	} else {
		fields = append(fields, field{list.Search, &detail.Search, true})
	}
	for _, field := range fields {
		old, oldOK := configurationString(field.old)
		newer, newOK := configurationString(*field.newer)
		if field.query && strings.TrimSpace(old) == "" {
			continue
		}
		if oldOK && newOK && old != newer {
			return list, true
		}
		if len(field.old) > 0 && string(field.old) != "null" {
			*field.newer = field.old
		}
	}
	return detail, false
}

var configurationMacroSuffix = regexp.MustCompile(`^(.*)\(([0-9]+)\)$`)

func configurationMacroIdentity(name string, rawArguments json.RawMessage) (string, int, []string, bool) {
	arity := 0
	base := name
	if match := configurationMacroSuffix.FindStringSubmatch(name); match != nil {
		parsed, err := strconv.Atoi(match[2])
		if err != nil || strings.TrimSpace(match[1]) == "" {
			return "", 0, nil, false
		}
		arity = parsed
		base = match[1]
	} else if strings.ContainsAny(name, "()") {
		return "", 0, nil, false
	}
	// macros.conf defines a bare stanza as zero-arity and ignores its args
	// setting. Only omission or a JSON string is supported evidence here.
	if len(rawArguments) == 0 {
		return base, arity, []string{}, arity == 0
	}
	raw, ok := configurationString(rawArguments)
	if !ok {
		return "", 0, nil, false
	}
	args := []string{}
	if arity == 0 {
		return base, arity, args, true
	}
	seen := map[string]bool{}
	if strings.TrimSpace(raw) != "" {
		for _, value := range strings.Split(raw, ",") {
			arg := strings.TrimSpace(value)
			if arg == "" || seen[arg] {
				return "", 0, nil, false
			}
			seen[arg] = true
			args = append(args, arg)
		}
	}
	return base, arity, args, len(args) == arity
}
func configurationBoolean(raw json.RawMessage) (*bool, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, true
	}
	var value bool
	if json.Unmarshal(raw, &value) == nil {
		return &value, true
	}
	text := string(raw)
	if str, ok := configurationString(raw); ok {
		text = str
	}
	switch text {
	case "0", "false":
		value = false
	case "1", "true":
		value = true
	default:
		return nil, false
	}
	return &value, true
}
func configurationMappedObject(instanceID string, adapter configurationAdapter, name string, ctx configurationContext, content configurationContent, provenance environment.Provenance) (environment.Object, []string, bool) {
	object := environment.Object{Kind: adapter.kind, Name: adapter.namePrefix + name, Namespace: configurationNamespace(ctx), Owner: ctx.owner, App: ctx.app, Sharing: ctx.sharing, Provenance: provenance}
	diagnostics := []string{}
	rawDefinition := content.Search
	definitionSupported := true
	switch adapter.kind {
	case "macro":
		base, arity, args, ok := configurationMacroIdentity(name, content.Arguments)
		if !ok {
			return object, []string{"macro_arity_invalid"}, false
		}
		object.Name = base
		object.Arity = &arity
		object.Arguments = args
		eval, valid := configurationBoolean(content.EvalBased)
		if len(content.EvalBased) == 0 {
			// macros.conf defaults iseval to false after any necessary detail
			// merge. Explicit null remains unknown rather than a default.
			value := false
			eval = &value
		}
		if !valid {
			diagnostics = append(diagnostics, "macro_metadata_invalid")
		}
		object.EvalBased = eval
		if eval == nil {
			definitionSupported = false
			if valid {
				diagnostics = append(diagnostics, "macro_metadata_unavailable")
			}
		}
		if len(content.Validation) > 0 && string(content.Validation) != "null" {
			validation, ok := configurationString(content.Validation)
			if ok {
				object.Validation = &validation
			} else {
				diagnostics = append(diagnostics, "macro_metadata_invalid")
				definitionSupported = false
			}
		}
		rawDefinition = content.Definition
	case "saved_search", "event_type":
	default:
		diagnostics = append(diagnostics, "definition_unsupported")
		object.ID = stableObjectID(instanceID, object.Kind, object.Name, object.Namespace, object.App, object.Owner, nil)
		return object, diagnostics, true
	}
	definition, ok := configurationString(rawDefinition)
	if definitionSupported && ok && strings.TrimSpace(definition) != "" {
		object.Document = &analysis.QueryDocument{Text: definition, Language: "spl", Profile: "splunkd", Version: "current", SourceID: provenance.SourceID}
	} else {
		diagnostics = append(diagnostics, "definition_unavailable")
	}
	object.ID = stableObjectID(instanceID, object.Kind, object.Name, object.Namespace, object.App, object.Owner, object.Arity)
	return object, diagnostics, true
}
func configurationObjectContent(object environment.Object) string {
	object.Provenance = environment.Provenance{}
	if object.Document != nil {
		document := *object.Document
		document.SourceID = ""
		object.Document = &document
	}
	return inventoryHash([]any{object})
}
func (c *Client) collectConfiguration(ctx context.Context, instanceID string) ConfigurationResult {
	result := ConfigurationResult{Objects: []environment.Object{}, Collections: []environment.Collection{}, Diagnostics: []Diagnostic{}}
	objects := map[string]environment.Object{}
	safeContent := map[string]string{}
	gaps := map[string]string{}
	kinds := []string{}
	bytesRetained := 0
	addDiagnostic := func(code, kind string) {
		result.Diagnostics = append(result.Diagnostics, configurationDiagnostic(code, kind))
	}
	gap := func(code, kind string) {
		if gaps[kind] == "" {
			gaps[kind] = code
		}
		addDiagnostic(code, kind)
	}
	contexts := configurationRequestContexts(c.options.Scope)
	for _, adapter := range configurationAdapters {
		if _, found := gaps[adapter.kind]; !found {
			kinds = append(kinds, adapter.kind)
			gaps[adapter.kind] = ""
		}
		for _, requestContext := range contexts {
			path := configurationPath(requestContext, adapter.family)
			offset, total := 0, -1
			for {
				var page configurationPage
				query := url.Values{}
				if !adapter.unpaged {
					query.Set("count", strconv.Itoa(configurationPageSize))
					query.Set("offset", strconv.Itoa(offset))
				}
				if ctx.Err() != nil {
					gap("configuration_not_attempted", adapter.kind)
					break
				}
				if err := c.getJSON(ctx, path, query, &page); err != nil {
					gap(err.Error(), adapter.kind)
					break
				}
				if adapter.unpaged {
					// search/tags is a documented identity-only feed with no paging contract.
					if page.Entries == nil {
						gap("configuration_page_invalid", adapter.kind)
						break
					}
				} else if !configurationPageValid(page, offset, total) {
					gap("configuration_page_invalid", adapter.kind)
					break
				} else {
					total = *page.Paging.Total
				}
				if configurationPageWarning(page) {
					gap("configuration_messages", adapter.kind)
				}
				for _, entry := range page.Entries {
					name, valid := identityName(entry.Name)
					if !valid {
						gap("configuration_identity_invalid", adapter.kind)
						continue
					}
					objectContext, contextCode := configurationACL(entry)
					if contextCode != "" {
						gap(contextCode, adapter.kind)
					}
					if contextCode == "context_conflict" || contextCode == "context_invalid" || !configurationScopeMatches(c.options.Scope, objectContext) {
						continue
					}
					content := entry.Content
					if configurationNeedsDetail(adapter.kind, name, content) {
						if detail, ok := c.configurationDetail(ctx, adapter, requestContext, objectContext, name); ok {
							merged, conflict := configurationMergeContent(adapter.kind, name, content, detail.Content)
							if conflict {
								gap("definition_conflict", adapter.kind)
							} else {
								content = merged
							}
						} else {
							gap("configuration_detail_unavailable", adapter.kind)
						}
					}
					acquisition := path + "/" + url.PathEscape(name)
					provenance := acquisitionProvenance(instanceID, "splunk_rest", acquisition, time.Now().UTC().Format(time.RFC3339Nano))
					object, diagnostics, usable := configurationMappedObject(instanceID, adapter, name, objectContext, content, provenance)
					for _, code := range diagnostics {
						if code == "macro_arity_invalid" || code == "macro_metadata_invalid" {
							gap(code, adapter.kind)
						} else {
							addDiagnostic(code, adapter.kind)
						}
					}
					if !usable {
						continue
					}
					safe := configurationObjectContent(object)
					if existing, found := objects[object.ID]; found {
						if safeContent[object.ID] != safe {
							gap("object_conflict", adapter.kind)
						} else if object.Provenance.SourceID < existing.Provenance.SourceID {
							objects[object.ID] = object
						}
						continue
					}
					encoded, _ := json.Marshal(object)
					if len(encoded) > assembledLimit-bytesRetained {
						result.Err = failure("artifact_too_large")
						return result
					}
					bytesRetained += len(encoded)
					objects[object.ID] = object
					safeContent[object.ID] = safe
				}
				if adapter.unpaged {
					break
				}
				offset += len(page.Entries)
				if offset == total {
					break
				}
			}
		}
	}
	for _, object := range objects {
		result.Objects = append(result.Objects, object)
	}
	sort.Slice(result.Objects, func(i, j int) bool { return result.Objects[i].ID < result.Objects[j].ID })
	for _, kind := range kinds {
		collection := environment.Collection{Kind: kind, Coverage: "complete"}
		if code := gaps[kind]; code != "" {
			collection.Reason = code
			collection.Coverage = "unavailable"
			for _, object := range result.Objects {
				if object.Kind == kind {
					collection.Coverage = "partial"
					break
				}
			}
		}
		result.Collections = append(result.Collections, collection)
	}
	for _, kind := range []string{"dataset", "module", "function", "external_command"} {
		result.Collections = append(result.Collections, environment.Collection{Kind: kind, Coverage: "unavailable", Reason: "adapter_unsupported"})
		addDiagnostic("adapter_unsupported", kind)
	}
	sort.Slice(result.Collections, func(i, j int) bool { return result.Collections[i].Kind < result.Collections[j].Kind })
	return result
}
