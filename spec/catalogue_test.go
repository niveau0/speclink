package spec_test

import (
	"encoding/json"
	"testing"

	"github.com/worldiety/speclink/spec"
)

// The catalogue is filled at package initialisation, exactly as it is in a
// target project. These two are declared here rather than reusing the
// requirement in registry_test.go, so that what this file asserts about the
// catalogue does not depend on what that one asserts about the registry.

var rPlanned = spec.Declare(spec.Requirement{
	ID:     "R-CATALOGUE-PLANNED",
	Kind:   spec.Functional,
	Status: spec.Planned,
	Title:  "Something accepted but not built",
	Text:   "The system MUST eventually do this.",
})

var rInternal = spec.Declare(spec.Requirement{
	ID:         "R-CATALOGUE-INTERNAL",
	Kind:       spec.NonFunctional,
	Status:     spec.Normative,
	Disclosure: spec.Internal,
	Title:      "Something the operator knows",
	Text:       "The system MUST do this, and only the team needs to hear about it.",
})

// TestCatalogueSeesWhatTheRegistryCannot is the whole reason Declare exists.
//
// The binding registry knows only requirements something is bound to, and
// speclink demands a binding for normative ones alone. A planned requirement is
// deliberately bound to nothing. Reading the catalogue out of the registry would
// therefore answer "there is no such requirement" for it, which is a different
// and far worse statement than "not implemented yet".
func TestCatalogueSeesWhatTheRegistryCannot(t *testing.T) {
	if got := lookup(t, "R-CATALOGUE-PLANNED").Title; got != rPlanned.Title {
		t.Errorf("planned requirement missing from the catalogue, got title %q", got)
	}

	for _, e := range spec.Entries() {
		for _, a := range e.Assertions {
			for _, id := range a.Requirements {
				if id == "R-CATALOGUE-PLANNED" {
					t.Fatal("the fixture binds the planned requirement, so this test proves nothing")
				}
			}
		}
	}
}

// Declare returns its argument, so the var still holds the requirement and can
// be referenced by DerivedFrom, Supersedes and spec.Satisfies. Without that the
// call would be a second line to write and to forget.
func TestDeclarePassesTheValueThrough(t *testing.T) {
	if rInternal.ID != "R-CATALOGUE-INTERNAL" || rInternal.Disclosure != spec.Internal {
		t.Errorf("Declare did not return its argument unchanged: %+v", rInternal)
	}
}

// The zero value of Disclosure is Public. Every other enum here starts at
// iota + 1 so that zero means "not stated"; this one does not, and a change of
// mind about that would silently reclassify every requirement ever written.
func TestUnclassifiedIsPublic(t *testing.T) {
	if lookup(t, "R-CATALOGUE-PLANNED").Disclosure != spec.Public {
		t.Error("a requirement that says nothing about disclosure must be public")
	}
	if spec.Disclosure(0) != spec.Public {
		t.Error("the zero value of Disclosure is not Public")
	}
}

// A duplicate identifier is a programming error that would otherwise make the
// catalogue answer one identifier two different ways, depending on which entry
// a lookup reached first.
func TestDuplicateIdentifierPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("declaring the same ID twice was accepted")
		}
	}()
	spec.Declare(spec.Requirement{ID: "R-CATALOGUE-PLANNED"})
}

// TestVocabularyHasNames covers the three enums a consumer has to show to a
// human. Without names every one of them writes the same switch, and the
// spellings drift apart between reports of the same catalogue.
func TestVocabularyHasNames(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{spec.Functional.String(), "functional"},
		{spec.NonFunctional.String(), "nonFunctional"},
		{spec.Constraint.String(), "constraint"},
		{spec.Decision.String(), "decision"},
		{spec.Business.String(), "business"},
		{spec.Technical.String(), "technical"},
		{spec.Mixed.String(), "mixed"},
		{spec.Normative.String(), "normative"},
		{spec.Abstract.String(), "abstract"},
		{spec.Planned.String(), "planned"},
		{spec.OutOfScope.String(), "outOfScope"},
		{spec.Informative.String(), "informative"},
		{spec.Superseded.String(), "superseded"},
		{spec.Mockup.String(), "mockup"},
		{spec.Scribble.String(), "scribble"},
		{spec.Diagram.String(), "diagram"},
		{spec.AcceptanceCriteria.String(), "acceptanceCriteria"},
		{spec.Protocol.String(), "protocol"},
		{spec.Document.String(), "document"},
		{spec.Public.String(), "public"},
		{spec.Internal.String(), "internal"},
		{spec.Confidential.String(), "confidential"},
		{spec.Secret.String(), "secret"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("named %q, want %q", c.got, c.want)
		}
	}

	// A value from a newer spec than the reader is the same version skew
	// DumpVersion exists for. "unknown" is readable; a formatting error is a
	// bug report.
	if got := spec.Status(99).String(); got != "unknown" {
		t.Errorf("an unknown status named itself %q", got)
	}
}

// A requirement handed to a consumer as JSON must carry words, not the integers
// of an enum whose numbering is an implementation detail.
func TestRequirementMarshalsWithWords(t *testing.T) {
	b, err := json.Marshal(lookup(t, "R-CATALOGUE-INTERNAL"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"Status":"normative"`, `"Disclosure":"internal"`, `"Kind":"nonFunctional"`} {
		if !contains(string(b), want) {
			t.Errorf("expected %s in %s", want, b)
		}
	}
}

func lookup(t *testing.T, id spec.RequirementID) spec.Requirement {
	t.Helper()
	for _, r := range spec.Requirements() {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("requirement %s is not in the catalogue", id)
	return spec.Requirement{}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
