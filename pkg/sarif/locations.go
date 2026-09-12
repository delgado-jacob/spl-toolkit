package sarif

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

const sourceRootID = "SRCROOT"

func rootURI(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Scheme != "file" || !strings.HasPrefix(raw, "file://") || u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") || strings.Contains(u.Path, "\\") || !strings.HasPrefix(u.Path, "/") || !strings.HasSuffix(u.Path, "/") {
		return "", fmt.Errorf("sarif: file origin requires a trailing-slash absolute file URI")
	}
	return u.String(), nil
}

func relativeURI(raw string) (string, error) {
	if raw == "" || !utf8.ValidString(raw) || strings.HasPrefix(raw, "/") || strings.ContainsAny(raw, "\\\x00") {
		return "", fmt.Errorf("sarif: invalid relative file path %q", raw)
	}
	parts := strings.Split(raw, "/")
	for i, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("sarif: invalid relative file path %q", raw)
		}
		// PathEscape leaves ':' unescaped, but a colon in the first segment
		// would make the result parse as a URI scheme instead of a relative URI.
		parts[i] = strings.ReplaceAll(url.PathEscape(part), ":", "%3A")
	}
	return strings.Join(parts, "/"), nil
}

func originLocation(id string, origin corpus.Origin) (ArtifactLocation, string, error) {
	switch origin.Kind {
	case "file":
		root, err := rootURI(origin.BaseURI)
		if err != nil {
			return ArtifactLocation{}, "", err
		}
		path, err := relativeURI(origin.RelativePath)
		if err != nil {
			return ArtifactLocation{}, "", err
		}
		return ArtifactLocation{URI: path, URIBaseID: sourceRootID}, root, nil
	case "inline":
		if id == "" || !utf8.ValidString(id) {
			return ArtifactLocation{}, "", fmt.Errorf("sarif: invalid inline document ID")
		}
		return ArtifactLocation{URI: "spl-toolkit:///inline/" + url.PathEscape(id)}, "", nil
	default:
		return ArtifactLocation{}, "", fmt.Errorf("sarif: unsupported origin kind %q", origin.Kind)
	}
}

type coordinate struct{ line, column int }

type sourcePositions struct {
	length       int
	points       map[int]coordinate
	eofInsertion coordinate
	bomLength    int
}

// sourceCoordinates uses Unicode code points and treats CRLF as one newline.
// The analysis contract's offsets are UTF-8 bytes; its line/column values are
// deliberately not reused as SARIF positions.
func sourceCoordinates(text string, diagnostics []analysis.Diagnostic) sourcePositions {
	wanted := make(map[int]bool, 2*len(diagnostics)+1)
	for _, d := range diagnostics {
		wanted[d.Location.Start.Offset] = true
		wanted[d.Location.End.Offset] = true
	}
	points := make(map[int]coordinate, len(wanted)+1)
	line, column := 1, 1
	lastTerminated := coordinate{}
	bomLength := 0
	if strings.HasPrefix(text, "\ufeff") {
		bomLength = len("\ufeff")
	}
	for offset, r := range text {
		if wanted[offset] {
			points[offset] = coordinate{line, column}
		}
		if offset == 0 && bomLength != 0 {
			continue
		}
		switch r {
		case '\r':
			if offset+1 < len(text) && text[offset+1] == '\n' {
				column++
				continue
			}
			lastTerminated = coordinate{line, column + 1}
			line++
			column = 1
		case '\n':
			lastTerminated = coordinate{line, column + 1}
			line++
			column = 1
		default:
			column++
		}
	}
	points[len(text)] = coordinate{line, column}
	if strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\r") {
		return sourcePositions{len(text), points, lastTerminated, bomLength}
	}
	return sourcePositions{len(text), points, points[len(text)], bomLength}
}

func sourceRegion(source sourcePositions, location analysis.Location) (*Region, error) {
	start, end := location.Start, location.End
	if start == (analysis.Position{}) && end == (analysis.Position{}) {
		return nil, nil
	}
	if start.Line < 1 || start.Column < 1 || end.Line < 1 || end.Column < 1 || start.Offset < 0 || end.Offset < start.Offset || end.Offset > source.length {
		return nil, fmt.Errorf("sarif: invalid canonical source location")
	}
	a, okA := source.points[start.Offset]
	b, okB := source.points[end.Offset]
	if !okA || !okB {
		return nil, fmt.Errorf("sarif: source location splits a UTF-8 code point")
	}
	if source.bomLength != 0 && start.Offset < source.bomLength && end.Offset > start.Offset {
		offset, length := start.Offset, end.Offset-start.Offset
		return &Region{ByteOffset: &offset, ByteLength: &length}, nil
	}
	if start.Offset == end.Offset && end.Offset == source.length {
		a, b = source.eofInsertion, source.eofInsertion
	}
	return &Region{StartLine: a.line, StartColumn: a.column, EndLine: b.line, EndColumn: b.column}, nil
}
