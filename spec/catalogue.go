package spec

import (
	"sort"
	"sync"
)

var (
	catalogueMu sync.Mutex
	catalogue   []Requirement
	declaredAt  = map[RequirementID]bool{}
)

// Declare records a requirement in the runtime catalogue and returns it
// unchanged.
//
//	var RQuoteSubmit = spec.Declare(spec.Requirement{
//		ID: "R-QUOTE-SUBMIT",
//		…
//	})
//
// # Why this exists
//
// speclink reads the requirement tree statically and needs no help doing it.
// The program built from that tree does need help: a running application that
// wants to tell somebody *why* it behaves as it does has no way to enumerate
// the requirement values, because Go offers no reflection over package level
// variables. Without this, every project that wants its requirements at run
// time writes a go/ast parser and a code generator — a partial reimplementation
// of speclink's own frontend, once per project, each one subtly different.
//
// [Entries] is not a substitute, and the reason is the trap worth naming. The
// binding registry knows only requirements that something is bound to, and
// speclink demands a binding solely for [Normative] ones. A [Planned] or
// [Informative] requirement is deliberately bound to nothing, would be missing
// from that view, and an assistant reading it would answer "there is no such
// requirement" — which is a different and much worse statement than "it is not
// implemented yet".
//
// # What the catalogue is
//
// It is the requirements of *this binary*, not of the repository. Go
// initialises the package level variables of packages that are linked in, so a
// requirement package no main package imports is simply absent. For a program
// explaining itself that is arguably the right set — it should speak about what
// it is — but it is not the whole tree, and it must not be mistaken for it. A
// project that wants everything imports a package that pulls the tree in.
//
// # Duplicate identifiers
//
// A repeated ID panics. It is a programming error, it is found deterministically
// at process start rather than at some later request, and the alternative is a
// catalogue that answers one identifier two different ways depending on which
// entry the lookup happens to reach.
func Declare(r Requirement) Requirement {
	catalogueMu.Lock()
	defer catalogueMu.Unlock()
	if declaredAt[r.ID] {
		panic("spec: requirement " + string(r.ID) + " is declared twice")
	}
	declaredAt[r.ID] = true
	catalogue = append(catalogue, r)
	return r
}

// Requirements returns the requirements of this binary, sorted by ID.
//
// The sort makes the output stable; it carries no other meaning. See [Declare]
// for why this is the catalogue of the binary rather than of the repository.
func Requirements() []Requirement {
	catalogueMu.Lock()
	out := make([]Requirement, len(catalogue))
	copy(out, catalogue)
	catalogueMu.Unlock()

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
