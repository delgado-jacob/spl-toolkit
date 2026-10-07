package workflow

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
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
