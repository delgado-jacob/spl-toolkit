package workflow

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

// Compatibility owns closure evidence. Original proved defects always remain findings.
func compatibilityStatus(original *analysis.Result, evidence *compatibility.Report) analysis.Status {
	if original.Status == analysis.Invalid {
		return analysis.Invalid
	}
	switch evidence.Outcome {
	case "unsatisfied":
		return analysis.Invalid
	case "satisfied":
		return analysis.Valid
	default:
		return analysis.Incomplete
	}
}

// Resolution owns variant proof, closure, and publication decisions.
func resolutionStatus(original *analysis.Result, evidence *resolution.Report) analysis.Status {
	if original.Status == analysis.Invalid || evidence.Counts.Failed > 0 {
		return analysis.Invalid
	}
	if evidence.Counts.Incomplete > 0 || evidence.Counts.Verified == 0 {
		return analysis.Incomplete
	}
	return analysis.Valid
}

func finalize(report *Report) {
	for _, entry := range report.Entries {
		if entry.Failure == nil {
			report.Counts.Assessed++
		} else {
			report.ExecutionComplete = false
			switch entry.Failure.Phase {
			case "acquisition":
				report.Counts.AcquisitionFailed++
			case "configuration":
				report.Counts.ConfigurationFailed++
			default:
				report.Counts.InternalFailed++
			}
		}
		switch entry.Status {
		case analysis.Valid:
			report.Counts.Valid++
		case analysis.Invalid:
			report.Counts.Invalid++
		default:
			report.Counts.Incomplete++
		}
	}
	if report.Counts.Invalid > 0 {
		report.Status = analysis.Invalid
	} else if report.Counts.Incomplete > 0 || !report.ExecutionComplete {
		report.Status = analysis.Incomplete
	}
	switch {
	case !report.ExecutionComplete:
		report.CIExitCode = 2
	case report.Counts.Invalid > 0:
		report.CIExitCode = 1
	case report.Counts.Incomplete > 0:
		report.CIExitCode = 3
	default:
		report.CIExitCode = 0
	}
}
