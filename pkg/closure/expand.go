package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxExpansionDepth = 32
const maxExpansionCalls = 4096
const maxExpansionBytes = 1 << 20

type macroExpander struct {
	request Request
	calls   int
}

// expandMacros preserves opaque invocations in Text and records why each was
// held. A byte limit truncates output at the limit and records a zero-width gap.
func expandMacros(req Request) expansion {
	input := directExpansion(req.Document.Text, sourceInterval{Kind: "query", SourceID: req.Document.SourceID, Start: 0, End: len(req.Document.Text)})
	if req.Document.Language != "" && req.Document.Language != "spl" {
		return input
	}
	return (&macroExpander{request: req}).expand(input, map[string]bool{}, 0)
}

func (x *macroExpander) expand(input expansion, active map[string]bool, depth int) expansion {
	calls := scanMacroInvocations(input.Text)
	var out expansion
	cursor := 0
	for _, call := range calls {
		if call.Span.Start > cursor {
			if !appendLimited(&out, input.slice(cursor, call.Span.Start)) {
				return out
			}
		}
		site := input.slice(call.Span.Start, call.Span.End)
		part, reason := x.expandCall(input, call, active, depth)
		if reason != "" {
			part = expansion{}
			part.addGap(reason, site)
		}
		retainSiteZeroGaps(&part, site)
		if !appendLimited(&out, part) {
			return out
		}
		cursor = call.Span.End
	}
	if cursor < len(input.Text) || len(input.Text) == 0 {
		appendLimited(&out, input.slice(cursor, len(input.Text)))
	}
	return out
}

// A prior expansion may have left a zero-width gap inside this call. The
// replacement must carry that evidence even if the call discards its text.
func retainSiteZeroGaps(part *expansion, site expansion) {
	for _, gap := range site.Gaps {
		if gap.EffectiveStart != gap.EffectiveEnd {
			continue
		}
		duplicate := false
		for _, existing := range part.Gaps {
			if existing.Reason != gap.Reason || len(existing.Origins) != len(gap.Origins) {
				continue
			}
			same := true
			for i := range gap.Origins {
				if existing.Origins[i] != gap.Origins[i] {
					same = false
					break
				}
			}
			if same {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		// Replacement changes the call's length; anchor prior uncertainty
		// at its start so output truncation cannot discard it.
		gap.EffectiveStart, gap.EffectiveEnd = 0, 0
		part.Gaps = append(part.Gaps, gap)
	}
}

func appendLimited(out *expansion, part expansion) bool {
	remaining := maxExpansionBytes - len(out.Text)
	if len(part.Text) <= remaining {
		out.append(part)
		return true
	}
	for remaining > 0 && remaining < len(part.Text) && !utf8.RuneStart(part.Text[remaining]) {
		remaining--
	}
	if remaining > 0 {
		out.append(part.slice(0, remaining))
	}
	lost := part.Origins(remaining, len(part.Text))
	origins := make([]sourceInterval, 0, len(lost))
	for _, segment := range lost {
		origins = append(origins, segment.Source)
	}
	out.Gaps = append(out.Gaps, opaqueGap{EffectiveStart: len(out.Text), EffectiveEnd: len(out.Text), Reason: "limit", Origins: origins})
	return false
}

func (x *macroExpander) expandCall(input expansion, call macroInvocation, active map[string]bool, depth int) (expansion, string) {
	if call.Unsupported != "" {
		return expansion{}, "dynamic"
	}
	nameOrigins := input.Origins(call.NameSpan.Start, call.NameSpan.End)
	if len(nameOrigins) != 1 || nameOrigins[0].Placeholder != nil || nameOrigins[0].Source.End-nameOrigins[0].Source.Start != call.NameSpan.End-call.NameSpan.Start {
		return expansion{}, "dynamic"
	}
	if x.calls >= maxExpansionCalls || depth >= maxExpansionDepth {
		return expansion{}, "limit"
	}
	x.calls++
	def, reason := x.resolve(input, call)
	if reason != "" {
		return expansion{}, reason
	}
	if active[def.ID] {
		return expansion{}, "cycle"
	}
	if def.EvalBased != nil && *def.EvalBased {
		return expansion{}, "eval_based"
	}
	if def.Validation != nil && *def.Validation != "" {
		return expansion{}, "validation_held"
	}
	if def.Document == nil {
		return expansion{}, "missing"
	}
	if def.Document.Language != "" && def.Document.Language != "spl" {
		return expansion{}, "dynamic"
	}
	origin := input.Origins(call.Span.Start, call.Span.End)
	frame := invocationFrame{ObjectID: def.ID, InstanceID: fmt.Sprintf("expansion-%d", x.calls), Invocation: make([]sourceInterval, 0, len(origin))}
	for _, s := range origin {
		frame.Invocation = append(frame.Invocation, s.Source)
	}
	values, ok := macroValues(call, def)
	if !ok {
		return expansion{}, "dynamic"
	}
	active[def.ID] = true
	defer delete(active, def.ID)
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return values[names[i]].Span.Start < values[names[j]].Span.Start })
	definition := directExpansion(def.Document.Text, sourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID, Start: 0, End: len(def.Document.Text)})
	for _, ancestor := range nameOrigins[0].InvocationChain {
		definition = definition.withFrame(ancestor)
	}
	definition = definition.withFrame(frame)
	expanded := make(map[string]expansion)
	argument := func(name string) expansion {
		if value, found := expanded[name]; found {
			return value
		}
		arg := values[name]
		raw := input.slice(arg.ValueSpan.Start, arg.ValueSpan.End).withFrame(frame)
		value := x.expand(raw, active, depth+1)
		expanded[name] = value
		return value
	}
	substituted := substituteWith(definition, names, argument)
	for _, name := range names {
		if _, used := expanded[name]; used {
			continue
		}
		arg := values[name]
		raw := input.slice(arg.ValueSpan.Start, arg.ValueSpan.End).withFrame(frame)
		unused := x.expand(raw, active, depth+1)
		for _, gap := range unused.Gaps {
			gap.EffectiveStart, gap.EffectiveEnd = 0, 0
			substituted.Gaps = append(substituted.Gaps, gap)
		}
	}
	return x.expand(substituted, active, depth+1), ""
}

func macroValues(call macroInvocation, def Definition) (map[string]macroArgument, bool) {
	if len(call.Arguments) != len(def.Arguments) {
		return nil, false
	}
	values := make(map[string]macroArgument, len(call.Arguments))
	var positional []macroArgument
	known := make(map[string]bool, len(def.Arguments))
	for _, name := range def.Arguments {
		known[name] = true
	}
	for _, arg := range call.Arguments {
		if arg.Name == "" {
			positional = append(positional, arg)
			continue
		}
		if !known[arg.Name] {
			return nil, false
		}
		if _, found := values[arg.Name]; found {
			return nil, false
		}
		values[arg.Name] = arg
	}
	i := 0
	for _, name := range def.Arguments {
		if _, found := values[name]; found {
			continue
		}
		if i >= len(positional) {
			return nil, false
		}
		values[name] = positional[i]
		i++
	}
	return values, i == len(positional)
}

func substitute(definition expansion, values map[string]expansion) expansion {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	return substituteWith(definition, names, func(name string) expansion { return values[name] })
}

// substituteWith requests an argument only when its placeholder is reached.
// The caller may cache used arguments for repeated placeholders and inspect
// unused arguments separately without retaining their expanded text.
func substituteWith(definition expansion, names []string, valueFor func(string) expansion) expansion {
	var out expansion
	cursor := 0
	for cursor < len(definition.Text) {
		next, end, name := len(definition.Text), len(definition.Text), ""
		for _, candidate := range names {
			pattern := "$" + candidate + "$"
			at := strings.Index(definition.Text[cursor:], pattern)
			if at < 0 {
				continue
			}
			position := cursor + at
			if position < next || (position == next && (len(pattern) > end-next || (len(pattern) == end-next && candidate < name))) {
				next, end, name = position, position+len(pattern), candidate
			}
		}
		if name == "" {
			break
		}
		if !appendLimited(&out, definition.slice(cursor, next)) {
			return out
		}
		placeholder := definition.Origins(next, end)
		value := valueFor(name)
		if len(placeholder) == 1 {
			value = value.withPlaceholder(placeholder[0].Source)
		}
		if !appendLimited(&out, value) {
			return out
		}
		cursor = end
	}
	appendLimited(&out, definition.slice(cursor, len(definition.Text)))
	return out
}

func (x *macroExpander) resolve(input expansion, call macroInvocation) (Definition, string) {
	var matches []Definition
	for _, def := range x.request.Bundle.Objects {
		if def.Kind == "macro" && def.Name == call.Name && def.Arity != nil && *def.Arity == len(call.Arguments) {
			matches = append(matches, def)
		}
	}
	if len(matches) == 0 {
		return Definition{}, "missing"
	}
	if len(matches) == 1 {
		return matches[0], ""
	}
	site, ok := originalCallSite(input, call)
	if !ok {
		return Definition{}, "ambiguous"
	}
	var text string
	switch site.Kind {
	case "query":
		text = x.request.Document.Text
	case "definition":
		for _, def := range x.request.Bundle.Objects {
			if def.ID == site.ObjectID && def.Document != nil {
				text = def.Document.Text
				break
			}
		}
	}
	if site.End > len(text) {
		return Definition{}, "ambiguous"
	}
	sum := sha256.Sum256([]byte(text))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	var chosen *Definition
	for _, binding := range x.request.Bindings {
		if binding.Kind != "macro" || binding.DocumentDigest != digest || binding.Start != site.Start || binding.End != site.End {
			continue
		}
		for i := range matches {
			if matches[i].ID == binding.ObjectID {
				if chosen != nil && chosen.ID != matches[i].ID {
					return Definition{}, "ambiguous"
				}
				chosen = &matches[i]
			}
		}
	}
	if chosen == nil {
		return Definition{}, "ambiguous"
	}
	return *chosen, ""
}

// originalCallSite reconstructs a definition occurrence after argument
// substitution. Placeholder bytes point back to the interval they replaced.
func originalCallSite(input expansion, call macroInvocation) (sourceInterval, bool) {
	origins := input.Origins(call.Span.Start, call.Span.End)
	if len(origins) == 0 {
		return sourceInterval{}, false
	}
	first := origins[0].Source
	if origins[0].Placeholder != nil {
		first = *origins[0].Placeholder
	}
	site := first
	for _, seg := range origins[1:] {
		source := seg.Source
		if seg.Placeholder != nil {
			source = *seg.Placeholder
		}
		if source.Kind != site.Kind || source.SourceID != site.SourceID || source.ObjectID != site.ObjectID {
			return sourceInterval{}, false
		}
		if source.Start == site.End {
			site.End = source.End
			continue
		}
		// One placeholder may yield several segments after nested expansion.
		if source.Start >= site.Start && source.End <= site.End {
			continue
		}
		return sourceInterval{}, false
	}
	return site, true
}
