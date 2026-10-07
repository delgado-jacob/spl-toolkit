package sarif

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

// ExportAdditional extends the canonical corpus projection without re-analysis.
// Located findings must refer to a retained, hash-checked source document.
func ExportAdditional(report *corpus.Report, findings []AdditionalFinding) (*Log, error) {
	log, err := Export(report)
	if err != nil {
		return nil, err
	}
	run := &log.Runs[0]
	entries := map[string]corpus.ReportEntry{}
	for _, e := range report.Entries {
		entries[e.ID] = e
	}
	for _, f := range findings {
		e, ok := entries[f.DocumentID]
		if !ok {
			return nil, fmt.Errorf("sarif: unknown additional finding document %q", f.DocumentID)
		}
		if f.RuleID == "" || f.Message == "" {
			return nil, fmt.Errorf("sarif: incomplete additional finding identity")
		}
		if f.Level != "error" && f.Level != "warning" && f.Level != "note" && f.Level != "none" {
			return nil, fmt.Errorf("sarif: invalid additional finding level %q", f.Level)
		}
		r := Result{RuleID: f.RuleID, Level: f.Level, Message: Message{Text: f.Message}, Properties: f.Properties}
		if r.Properties == nil {
			r.Properties = map[string]any{}
		} else {
			raw, err := json.Marshal(r.Properties)
			if err != nil {
				return nil, err
			}
			r.Properties = nil
			if err := json.Unmarshal(raw, &r.Properties); err != nil {
				return nil, err
			}
		}
		r.Properties["document_id"] = f.DocumentID
		if f.Location != nil {
			a, _, _, _, err := evaluated(e.Evaluation)
			if err != nil {
				return nil, err
			}
			positions := sourceCoordinates(a.Document.Text, []analysis.Diagnostic{{Location: *f.Location}})
			region, err := sourceRegion(positions, *f.Location)
			if err != nil {
				return nil, err
			}
			if region != nil {
				loc, _, err := originLocation(e.ID, e.Origin)
				if err != nil {
					return nil, err
				}
				for i, artifact := range run.Artifacts {
					if artifact.Location.URI == loc.URI && artifact.Location.URIBaseID == loc.URIBaseID {
						index := i
						loc.Index = &index
						break
					}
				}
				if loc.Index == nil {
					return nil, fmt.Errorf("sarif: additional finding artifact missing")
				}
				r.Locations = []Location{{PhysicalLocation: PhysicalLocation{ArtifactLocation: loc, Region: region}}}
			}
		}
		run.Results = append(run.Results, r)
	}
	ids := map[string]bool{}
	for _, r := range run.Results {
		ids[r.RuleID] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	run.Tool.Driver.Rules = []Rule{}
	indices := map[string]int{}
	for i, id := range ordered {
		indices[id] = i
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, Rule{ID: id})
	}
	for i := range run.Results {
		run.Results[i].RuleIndex = indices[run.Results[i].RuleID]
	}
	return log, nil
}
