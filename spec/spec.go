// Package spec is the public directive catalogue of speclink.
//
// Target projects import this package from two kinds of files:
//
//   - <id>.spec.go            declares requirements (the requirement tree)
//   - <source>.annotation.go  asserts facts about the constructs of <source>.go
//
// Both file kinds are ordinary Go and are part of the normal build. The Go
// compiler therefore checks arity, argument types, field names, enum values and
// every identifier reference. speclink itself never executes these terms; it
// reads the typed AST.
//
// The declarations in this file are the schema. There is no separate schema
// language and no code generation.
package spec

// RequirementID is the stable identity of a requirement, e.g. "R-QUOTE-SUBMIT".
//
// The ID is path independent: it lives in the term, not in the directory
// layout. File name and attachment folder repeat it and are cross-checked by
// speclink. References from code use the Go identifier, never the ID string,
// so moving a requirement never breaks a reference.
type RequirementID string

// RuleID identifies a single speclink rule, e.g. "K4-QUERY-SUBJECT".
// Used by [Waive] to suspend that rule for one construct.
type RuleID string

// TermID identifies a glossary entry within its context.
type TermID string

// State is a coarse aggregate lifecycle state, e.g. "submitted".
// Deliberately not an event field: the state follows from the event type.
type State string

// Kind classifies what a requirement is. It determines the first directory
// level of the requirement tree: dec/, nfr/, cst/ and fun/<domain>/.
//
// Boundaries:
//
//   - Constraint vs NonFunctional: a constraint is imposed from outside and is
//     not negotiable (law, mandated platform). A non-functional requirement is
//     a quality goal we must actively meet.
//   - Decision vs Functional: a decision justifies a ruling, including
//     deliberate exclusions ("what we will not build"). The concrete functional
//     requirements that follow from it are carried separately.
type Kind int

const (
	// Functional describes concrete system behaviour.
	Functional Kind = iota + 1
	// NonFunctional is a quality goal: performance, security, auditability.
	NonFunctional
	// Constraint is imposed from outside: law, platform, technology.
	Constraint
	// Decision is a normative ruling and requires a Rationale.
	Decision
)

// Discipline records who owns a requirement.
type Discipline int

const (
	// Business is owned by the domain side.
	Business Discipline = iota + 1
	// Technical is owned by engineering.
	Technical
	// Mixed requires both.
	Mixed
)

// Status drives the backward direction of the coverage check (K3).
//
// Only [Normative] must be covered by at least one construct. All other values
// are the explicit, justified exemptions. This is what turns "I hope everything
// is implemented" into a number.
type Status int

const (
	// Normative must be covered by at least one construct.
	Normative Status = iota + 1
	// Abstract is a pure derivation node and must NOT be covered directly.
	Abstract
	// Planned is accepted but not yet implemented.
	Planned
	// OutOfScope is deliberately excluded.
	OutOfScope
	// Informative carries context only and states no obligation.
	Informative
	// Superseded has been replaced; the successor points here via Supersedes.
	Superseded
)

// Disclosure classifies who a requirement may be shown to, on the scale an
// information security management system uses.
//
// It exists because the catalogue is no longer only read by the team. A
// requirement that reaches an end user through an assistant, a help page or a
// support desk is disclosed, and somebody has to have decided that it may be.
//
// # Why the zero value is Public and not the safe end
//
// Every other enum in this package starts at iota + 1, so that the zero value
// means "not stated" and can be reported as an omission. This one does not, and
// the deviation is deliberate.
//
// A requirement says what the system must do. That is the thing the people
// using it are entitled to know, and material that must not be seen has no
// business being in a requirement in the first place — it belongs in the
// document the requirement points at through [Source], which speclink never
// copies. Reading silence as secrecy would make the catalogue mute on the day
// the field is introduced, and the predictable answer to a mute catalogue is
// somebody setting several hundred requirements to Public in one sitting
// without reading any of them. A classification arrived at that way is worth
// less than none, because it looks like a decision.
//
// So the default is the honest description of what these texts are, and the
// field is for the exception.
//
// # What it does not do
//
// Nothing enforces it. speclink cannot check a disclosure decision, and it does
// not filter its own output by it: the generated documentation still contains
// every requirement, whatever it is classified as. The field states an intent
// for the consumer that shows the text to somebody; it is not a control.
type Disclosure int

const (
	// Public may be shown to anyone, including end users. The zero value.
	Public Disclosure = iota
	// Internal is for the organisation building or operating the system.
	Internal
	// Confidential is for a named circle within it.
	Confidential
	// Secret is disclosed individually and never in bulk.
	Secret
)

// Role classifies accompanying material of a requirement.
type Role int

const (
	// Mockup is a rendered design of a screen or document.
	Mockup Role = iota + 1
	// Scribble is an informal sketch.
	Scribble
	// Diagram is a process or structure drawing.
	Diagram
	// AcceptanceCriteria spells out how fulfilment is judged.
	AcceptanceCriteria
	// Protocol records a decision or an acceptance meeting.
	Protocol
	// Document is any other accompanying file.
	Document
)
