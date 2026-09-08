package analysis

import (
	"fmt"
	"sort"
)

func spl2Capabilities() CapabilityManifest {
	m := CapabilityManifest{SchemaVersion: 1, Language: "spl2", Profile: "splunkd", Version: "current", Commands: []Capability{}, Functions: []Capability{}}
	for _, name := range []string{"search", "from", "eval", "where", "fields", "table", "rename", "stats", "eventstats", "streamstats", "lookup", "sort", "dedup", "head", "reverse"} {
		m.Commands = append(m.Commands, Capability{Name: name, SyntaxSupported: true, SemanticSupported: true, Limitations: []string{"Approved ordinary forms only; unsupported options, source/path membership and unknown output labels retain incomplete coverage."}})
	}
	for _, name := range []string{"select", "join", "append", "appendpipe", "appendcols", "union", "if", "spl1", "bin", "rex", "spath", "makeresults", "loadjob", "tstats", "mstats", "timechart", "timewrap", "makemv", "mvexpand", "mvcombine", "fillnull"} {
		m.Commands = append(m.Commands, Capability{Name: name, SyntaxSupported: true, Limitations: []string{"Dedicated grammar; field effects are not yet modeled."}})
	}
	for name, spec := range spl2Functions {
		limit := fmt.Sprintf("%d to %d positional arguments", spec.min, spec.max)
		if spec.max < 0 {
			limit = fmt.Sprintf("at least %d positional arguments", spec.min)
		}
		if name == "case" {
			limit += " in condition/value pairs"
		}
		if spec.aggregate {
			limit += "; aggregate context only"
		} else {
			limit += "; expression context only"
		}
		limit += "; conditional availability follows expression evidence; named arguments remain incomplete"
		m.Functions = append(m.Functions, Capability{Name: name, SyntaxSupported: true, SemanticSupported: true, Limitations: []string{limit}})
	}
	sort.Slice(m.Commands, func(i, j int) bool { return m.Commands[i].Name < m.Commands[j].Name })
	sort.Slice(m.Functions, func(i, j int) bool { return m.Functions[i].Name < m.Functions[j].Name })
	return m
}
