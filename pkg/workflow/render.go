package workflow

import "fmt"

// RenderAssessment projects retained canonical evidence without evaluating it.
// An empty format selects the canonical JSON report.
func RenderAssessment(report *Report, format string) (any, error) {
	if report == nil {
		return nil, fmt.Errorf("workflow: report is required")
	}
	switch format {
	case "", "json":
		return detachedExport(report)
	case "text":
		return FormatReport(report), nil
	case "sarif":
		return ExportSARIF(report)
	case "graph":
		return ExportGraph(report)
	case "bom":
		return ExportBOM(report)
	default:
		return nil, requestErrorAt("request_invalid", "/format", "unsupported format")
	}
}

// AssessOutput returns the selected transport payload alongside its canonical
// report, whose CI result remains authoritative for every projection.
func AssessOutput(request Request) (any, *Report, error) {
	switch request.Format {
	case "", "text", "json", "sarif", "graph", "bom":
	default:
		return nil, nil, requestErrorAt("request_invalid", "/format", "unsupported format")
	}
	report, err := Assess(request)
	if err != nil {
		return nil, nil, err
	}
	payload, err := RenderAssessment(report, request.Format)
	return payload, report, err
}

// AssessOutputJSON strictly admits an inline request and honors its format.
// Text output is a string and remains a JSON string for owned native results.
func AssessOutputJSON(raw []byte) (any, error) {
	request, err := DecodeRequest(raw)
	if err != nil {
		return nil, err
	}
	payload, _, err := AssessOutput(request)
	return payload, err
}
