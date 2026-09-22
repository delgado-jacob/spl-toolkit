package analysis

import (
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strings"
)

type functionSpec struct {
	min, max           int
	aggregate, dynamic bool
}

var functions = map[string]functionSpec{
	"abs":            {1, 1, false, false},
	"round":          {1, 2, false, false},
	"ceil":           {1, 1, false, false},
	"ceiling":        {1, 1, false, false},
	"floor":          {1, 1, false, false},
	"len":            {1, 1, false, false},
	"lower":          {1, 1, false, false},
	"upper":          {1, 1, false, false},
	"trim":           {1, 2, false, false},
	"ltrim":          {1, 2, false, false},
	"rtrim":          {1, 2, false, false},
	"substr":         {2, 3, false, false},
	"replace":        {3, 3, false, false},
	"coalesce":       {2, -1, false, false},
	"if":             {3, 3, false, false},
	"case":           {2, -1, false, false},
	"isnull":         {1, 1, false, false},
	"isnotnull":      {1, 1, false, false},
	"tonumber":       {1, 2, false, false},
	"tostring":       {1, 2, false, false},
	"mvcount":        {1, 1, false, false},
	"mvindex":        {2, 3, false, false},
	"mvfind":         {2, 2, false, false},
	"split":          {2, 2, false, false},
	"match":          {2, 2, false, false},
	"true":           {0, 0, false, false},
	"null":           {0, 0, false, false},
	"now":            {0, 0, false, false},
	"relative_time":  {2, 2, false, false},
	"strftime":       {2, 2, false, false},
	"like":           {2, 2, false, false},
	"count":          {0, 1, true, false},
	"sum":            {1, 1, true, false},
	"avg":            {1, 1, true, false},
	"min":            {1, 1, true, false},
	"max":            {1, 1, true, false},
	"values":         {1, 1, true, false},
	"list":           {1, 1, true, false},
	"dc":             {1, 1, true, false},
	"distinct_count": {1, 1, true, false},
	"first":          {1, 1, true, false},
	"last":           {1, 1, true, false},
	"earliest":       {1, 1, true, false},
	"latest":         {1, 1, true, false},
	"stdev":          {1, 1, true, false},
	"searchmatch":    {1, 1, false, true},
}

func (s *semanticStage) function(ctx parser.IAnalysisFunctionCallContext, aggregate bool) bool {
	name := strings.ToLower(ctx.AnalysisFunctionName().GetText())
	spec, ok := functions[name]
	if !ok {
		s.diagnostic(CodeUnsupportedFunction, fmt.Sprintf("function %q is unsupported", name), ctx)
		return false
	}
	if spec.dynamic {
		s.diagnostic(CodeDynamicReference, fmt.Sprintf("function %q has unresolved dynamic query semantics", name), ctx)
		return false
	}
	n := 0
	if ctx.AnalysisArgumentList() != nil {
		n = len(ctx.AnalysisArgumentList().AllAnalysisExpression())
	}
	if n < spec.min || (spec.max >= 0 && n > spec.max) || (name == "case" && n%2 != 0) || spec.aggregate != aggregate {
		s.diagnostic(CodeUnsupportedSemantics, fmt.Sprintf("function %q has unsupported arguments or context", name), ctx)
		return false
	}
	return true
}
