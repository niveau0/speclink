// Package ir holds the language neutral intermediate model.
//
// The ir is not an external interface. It has no serialisation format, no
// version and needs no round trip tests: one binary, one process, no boundary.
//
// Its only purpose is type isolation:
//
//	go/types.Type, token.Pos and ast.Node must never reach
//	internal/check, internal/diag or internal/backend.
//
// Without that boundary the rules and backends would depend directly on Go type
// information, and a second language frontend could not be added without
// rewriting them. The boundary costs almost nothing; it only decides where the
// data types live.
//
// A direct consequence: positions here are File/Line/Col, not token.Pos, which
// is meaningless without a FileSet and therefore language bound.
package ir

import "fmt"

// Position is a source location. Column is 1 based, 0 when unknown.
type Position struct {
	File string
	Line int
	Col  int
}

// Less orders two positions by file and line, so that findings come out in the
// order somebody reads the source rather than in map order.
func (p Position) Less(q Position) bool {
	if p.File != q.File {
		return p.File < q.File
	}
	if p.Line != q.Line {
		return p.Line < q.Line
	}
	return p.Col < q.Col
}

func (p Position) String() string {
	if p.File == "" {
		return "<unknown>"
	}
	if p.Col > 0 {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
	}
	return fmt.Sprintf("%s:%d", p.File, p.Line)
}

// TargetKind is the sort of construct a binding attaches to.
//
// The kinds are categories, not the declaration forms of one language. Each
// frontend maps its own forms onto them and the rules only ever ask which
// category a target is in.
//
// Func, Var and Const are not chosen by the author. The binding names a
// declaration and the frontend decides which of the three it is, so the
// annotation cannot disagree with the code.
type TargetKind int

const (
	// TargetType is a named type: a struct, an interface, a class, a trait.
	TargetType TargetKind = iota + 1
	// TargetFunc is something callable: a function, a method, an associated
	// function.
	TargetFunc
	// TargetVar is a mutable value at the top level of a unit.
	TargetVar
	// TargetConst is an immutable, statically known value.
	TargetConst
	// TargetField is one field of a type; Target.Field names it.
	TargetField
	// TargetPackage is a unit of code organisation as the frontend knows it:
	// a Go package, a JVM package, a Rust module.
	TargetPackage
	// TargetProcess is a course of business rather than a place in the code.
	//
	// It exists so that a process can satisfy a requirement through the same
	// machinery a construct does. A requirement about the course of business
	// is answered by the course, not by whichever use case happens to be named
	// in it, and without this it would read as covered by nothing.
	TargetProcess
	// TargetChannel is a way across a boundary rather than a place in the code.
	TargetChannel
)

func (k TargetKind) String() string {
	switch k {
	case TargetType:
		return "type"
	case TargetFunc:
		return "func"
	case TargetVar:
		return "var"
	case TargetConst:
		return "constant"
	case TargetField:
		return "field"
	case TargetPackage:
		return "package"
	case TargetProcess:
		return "process"
	case TargetChannel:
		return "channel"
	}
	return "unknown"
}

// Target names the construct a binding attaches to.
//
// Name is fully qualified in the frontend's own spelling, e.g.
// "example.com/m/sales.SubmitQuoteUC" in Go. Field is set for TargetField only.
type Target struct {
	Kind    TargetKind
	Package string
	Name    string
	Field   string
}

func (t Target) String() string {
	if t.Kind == TargetField {
		return t.Name + "." + t.Field
	}
	if t.Name == "" {
		return t.Package
	}
	return t.Name
}

// AssertionKind discriminates the payload of an [Assertion].
type AssertionKind int

const (
	AssertSatisfies AssertionKind = iota + 1
	AssertTransition
	AssertExternal
	AssertHelp
	AssertTerm
	AssertRationale
	AssertWaive
	AssertDraft
	AssertOptional
	// AssertPersistence marks a type as storage where the framework does not
	// say so by itself: an interface as a port, a struct as a written shape.
	AssertPersistence
	// AssertStoredAs marks a type as the written form of a domain type, which
	// moves the promise onto it and leaves the domain type free.
	AssertStoredAs
	// AssertRestrict states in prose the rule a value of a type must satisfy,
	// which the type system cannot carry.
	AssertRestrict
	// AssertValid gives examples a conforming implementation must accept.
	AssertValid
	// AssertInvalid gives examples it must reject.
	AssertInvalid
	// AssertClaim marks a field as asserted by the sender and not to be taken
	// as true by the receiver.
	AssertClaim
	// AssertVerified is the only assertion that is not read from an annotation
	// file. It comes from a spec.Verified call inside a test, whose target is
	// the test function.
	AssertVerified
)

func (k AssertionKind) String() string {
	switch k {
	case AssertSatisfies:
		return "satisfies"
	case AssertTransition:
		return "transition"
	case AssertExternal:
		return "external"
	case AssertHelp:
		return "help"
	case AssertTerm:
		return "term"
	case AssertRationale:
		return "rationale"
	case AssertWaive:
		return "waive"
	case AssertDraft:
		return "draft"
	case AssertOptional:
		return "optional"
	case AssertPersistence:
		return "persistence"
	case AssertStoredAs:
		return "storedAs"
	case AssertRestrict:
		return "restrict"
	case AssertValid:
		return "valid"
	case AssertInvalid:
		return "invalid"
	case AssertClaim:
		return "claim"
	case AssertVerified:
		return "verified"
	}
	return "unknown"
}

// Assertion is one statement about the construct named by its binding.
// Only the fields relevant for Kind are populated.
type Assertion struct {
	Kind AssertionKind
	Pos  Position

	// Requirements holds the requirement IDs of an AssertSatisfies.
	Requirements []string
	// EventType is the fully qualified event type of an AssertTransition.
	EventType string
	// DomainType is the fully qualified type an AssertStoredAs writes down.
	DomainType string
	// State is the target lifecycle state of an AssertTransition.
	State string
	// Text carries help, rationale or waiver reason.
	Text string
	// Term is the glossary ID of an AssertTerm.
	Term string
	// Rule is the suspended rule ID of an AssertWaive.
	Rule string
	// Vectors carries the examples of an AssertValid or AssertInvalid.
	Vectors []string
}

// Binding attaches a set of assertions to one construct.
type Binding struct {
	Target     Target
	Assertions []Assertion
	Pos        Position
}
