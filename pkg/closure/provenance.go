package closure

// sourceInterval is a half-open UTF-8 byte range in a caller-owned document.
// ObjectID distinguishes definitions that share a source file.
type sourceInterval struct {
	Kind     string
	SourceID string
	ObjectID string
	Start    int
	End      int
}

type invocationFrame struct {
	ObjectID   string
	InstanceID string
	Invocation []sourceInterval
}

type provenanceSegment struct {
	EffectiveStart  int
	EffectiveEnd    int
	Source          sourceInterval
	InvocationChain []invocationFrame
	Placeholder     *sourceInterval  // outermost placeholder, retained for occurrence matching
	Placeholders    []sourceInterval // outermost to innermost substitution sites
}

type opaqueGap struct {
	EffectiveStart int
	EffectiveEnd   int
	Reason         string
	Origins        []sourceInterval
}

type expansion struct {
	Text     string
	Segments []provenanceSegment
	Gaps     []opaqueGap
}

// Origins returns every provenance segment intersecting an effective byte range.
// Direct copies are clipped to the intersection; substitutions retain their
// placeholder evidence. A range spanning sources returns all of them in order.
func (e expansion) Origins(start, end int) []provenanceSegment {
	if start < 0 || end <= start || start >= len(e.Text) {
		return nil
	}
	if end > len(e.Text) {
		end = len(e.Text)
	}
	var out []provenanceSegment
	for _, s := range e.Segments {
		lo, hi := max(start, s.EffectiveStart), min(end, s.EffectiveEnd)
		if lo >= hi {
			continue
		}
		s.Source.Start += lo - s.EffectiveStart
		s.Source.End -= s.EffectiveEnd - hi
		s.EffectiveStart, s.EffectiveEnd = lo, hi
		out = append(out, s)
	}
	return out
}

func directExpansion(text string, source sourceInterval) expansion {
	out := expansion{Text: text}
	if len(text) > 0 {
		out.Segments = []provenanceSegment{{EffectiveEnd: len(text), Source: source}}
	}
	return out
}

func (e expansion) slice(start, end int) expansion {
	if start < 0 {
		start = 0
	}
	if end > len(e.Text) {
		end = len(e.Text)
	}
	if end < start {
		return expansion{}
	}
	out := expansion{Text: e.Text[start:end]}
	for _, s := range e.Origins(start, end) {
		s.EffectiveStart -= start
		s.EffectiveEnd -= start
		out.Segments = append(out.Segments, s)
	}
	for _, g := range e.Gaps {
		// A point gap belongs to the slice ending at that byte. The first
		// slice also owns a gap at byte zero.
		if g.EffectiveStart == g.EffectiveEnd {
			point := g.EffectiveStart
			if !((point > start && point <= end) || (point == 0 && start == 0) || (start == end && point == start)) {
				continue
			}
			g.EffectiveStart, g.EffectiveEnd = point-start, point-start
			out.Gaps = append(out.Gaps, g)
			continue
		}
		if g.EffectiveEnd <= start || g.EffectiveStart >= end {
			continue
		}
		g.EffectiveStart = max(g.EffectiveStart, start) - start
		g.EffectiveEnd = min(g.EffectiveEnd, end) - start
		out.Gaps = append(out.Gaps, g)
	}
	return out
}

func (e *expansion) append(part expansion) {
	offset := len(e.Text)
	e.Text += part.Text
	for _, s := range part.Segments {
		s.EffectiveStart += offset
		s.EffectiveEnd += offset
		e.Segments = append(e.Segments, s)
	}
	for _, g := range part.Gaps {
		g.EffectiveStart += offset
		g.EffectiveEnd += offset
		e.Gaps = append(e.Gaps, g)
	}
}

func (e expansion) withFrame(frame invocationFrame) expansion {
	for i := range e.Segments {
		chain := make([]invocationFrame, 0, len(e.Segments[i].InvocationChain)+1)
		chain = append(chain, e.Segments[i].InvocationChain...)
		chain = append(chain, frame)
		e.Segments[i].InvocationChain = chain
	}
	return e
}
func (e expansion) withPlaceholder(source sourceInterval) expansion {
	e.Segments = append([]provenanceSegment(nil), e.Segments...)
	for i := range e.Segments {
		prior := e.Segments[i].Placeholders
		if len(prior) == 0 && e.Segments[i].Placeholder != nil {
			prior = []sourceInterval{*e.Segments[i].Placeholder}
		}
		all := make([]sourceInterval, 0, len(prior)+1)
		all = append(all, source)
		all = append(all, prior...)
		e.Segments[i].Placeholders = all
		s := source
		e.Segments[i].Placeholder = &s
	}
	return e
}
func (e *expansion) addGap(reason string, source expansion) {
	origins := make([]sourceInterval, 0, len(source.Segments))
	for _, segment := range source.Segments {
		origins = append(origins, segment.Source)
	}
	e.Gaps = append(e.Gaps, opaqueGap{EffectiveStart: len(e.Text), EffectiveEnd: len(e.Text) + len(source.Text), Reason: reason, Origins: origins})
	e.append(source)
}
