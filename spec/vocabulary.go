package spec

// The vocabulary of a requirement in words.
//
// Every one of these types is an int, and every consumer that shows a
// requirement to a human needs a name for it. Without these methods each of
// them writes the same switch, the spellings drift apart, and two reports of
// the same catalogue disagree about what "Kind 3" is called.
//
// The names are English and stable: they are identifiers, not prose. A German
// user interface translates them, and translating a known token is a smaller
// job than guessing what an integer meant. Nothing here is localised, because a
// library that picks a language picks the wrong one for somebody.
//
// An unknown value answers "unknown" rather than panicking or returning the
// number. A catalogue written against a newer spec than the reader is exactly
// the version skew [DumpVersion] exists for, and a report that says "unknown"
// is readable, while one that says "%!Kind(7)" is a bug report.

// String returns the stable identifier of the kind.
func (k Kind) String() string {
	switch k {
	case Functional:
		return "functional"
	case NonFunctional:
		return "nonFunctional"
	case Constraint:
		return "constraint"
	case Decision:
		return "decision"
	}
	return "unknown"
}

// MarshalText writes the identifier rather than the number, so that JSON
// produced from a requirement can be read without this package at hand.
func (k Kind) MarshalText() ([]byte, error) { return []byte(k.String()), nil }

// String returns the stable identifier of the discipline.
func (d Discipline) String() string {
	switch d {
	case Business:
		return "business"
	case Technical:
		return "technical"
	case Mixed:
		return "mixed"
	}
	return "unknown"
}

// MarshalText writes the identifier rather than the number.
func (d Discipline) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// String returns the stable identifier of the status.
func (s Status) String() string {
	switch s {
	case Normative:
		return "normative"
	case Abstract:
		return "abstract"
	case Planned:
		return "planned"
	case OutOfScope:
		return "outOfScope"
	case Informative:
		return "informative"
	case Superseded:
		return "superseded"
	}
	return "unknown"
}

// MarshalText writes the identifier rather than the number.
func (s Status) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// String returns the stable identifier of the role.
func (r Role) String() string {
	switch r {
	case Mockup:
		return "mockup"
	case Scribble:
		return "scribble"
	case Diagram:
		return "diagram"
	case AcceptanceCriteria:
		return "acceptanceCriteria"
	case Protocol:
		return "protocol"
	case Document:
		return "document"
	}
	return "unknown"
}

// MarshalText writes the identifier rather than the number.
func (r Role) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

// String returns the stable identifier of the classification.
func (d Disclosure) String() string {
	switch d {
	case Public:
		return "public"
	case Internal:
		return "internal"
	case Confidential:
		return "confidential"
	case Secret:
		return "secret"
	}
	return "unknown"
}

// MarshalText writes the identifier rather than the number.
func (d Disclosure) MarshalText() ([]byte, error) { return []byte(d.String()), nil }
