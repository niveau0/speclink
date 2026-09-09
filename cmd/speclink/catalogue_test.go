package main

import (
	"strings"
	"testing"
)

// TestARequirementReachesTheRunningProgram covers the rule that keeps the
// runtime catalogue whole.
//
// speclink reads the requirement tree statically and needs no help. The program
// built from that tree does: Go offers no reflection over package level
// variables, so an application that wants to explain why it behaves as it does
// cannot enumerate its own requirements. spec.Declare is what registers them,
// and a rule is what stops it from being the line everybody forgets.
func TestARequirementReachesTheRunningProgram(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "requirements/fun/quote/R-QUOTE-SUBMIT.spec.go",
		"spec.Declare(spec.Requirement{", "spec.Requirement{")
	// Unwrapping the call leaves the closing "})" behind, which no longer
	// parses. The literal ends with the brace alone.
	rewrite(t, dir, "requirements/fun/quote/R-QUOTE-SUBMIT.spec.go", "\n})\n", "\n}\n")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a requirement invisible to its own program was accepted:\n%s", out)
	}
	if !strings.Contains(out, "RQuoteSubmit is not declared through spec.Declare") {
		t.Errorf("expected K1-REQ-UNDECLARED:\n%s", out)
	}
	// One finding, not two. The requirement is still read, so nothing it covers
	// is reported as suddenly uncovered — a second wave of findings about the
	// wrong subject is how a small mistake becomes an unreadable run.
	if strings.Contains(out, "R-QUOTE-SUBMIT is satisfied by nothing") {
		t.Errorf("the requirement was dropped rather than reported:\n%s", out)
	}
}

// TestASplitSentenceIsStillRefusedButExplained is about what a diagnostic is
// for.
//
// Two string literals glued with a + fold to a constant before speclink ever
// sees them, so the usual reason for refusing an expression — that a computed
// value is invisible to the verifier — does not apply here. The refusal stands
// on a different ground: a sentence cut at a seam cannot be found again. grep
// for the words either side of the + comes back empty while the requirement
// sits right there in the tree.
//
// Which means the finding has to teach the raw string, because "state the fact
// directly" is exactly what the author was trying to do when the line got too
// long.
func TestASplitSentenceIsStillRefusedButExplained(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "requirements/dec/R-DEC-NUMBERING.spec.go",
		`Rationale:    "Gapless numbering is a bookkeeping obligation`,
		`Rationale:    "Gapless numbering is a bookkeeping " + "obligation`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a sentence assembled from literals was accepted:\n%s", out)
	}
	if !strings.Contains(out, "a text is assembled from several literals") {
		t.Errorf("expected the concatenation finding:\n%s", out)
	}
	if !strings.Contains(out, "raw string in backticks") {
		t.Errorf("the finding does not name the form that works:\n%s", out)
	}
}

// And the form it names has to be one the reader accepts, or the advice sends
// the author into the next finding. The conformant fixture carries a multi-line
// rationale as a raw string; this reads it back through the export.
func TestARawStringSurvivesTheReader(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "requirements", "../../testdata/example", "-format", "json", "./requirements/...")
	if code != 0 {
		t.Fatalf("expected a clean tree, got exit %d:\n%s", code, out)
	}
	// The line break inside the literal is part of the value and arrives as
	// \n in the JSON. What matters is that the sentence is whole: a reader
	// searching for words either side of the break finds them.
	if !strings.Contains(out, "the history is not evidence but noise") {
		t.Errorf("the first line of the raw rationale was lost:\n%s", out)
	}
	if !strings.Contains(out, "the expensive one.") {
		t.Errorf("the last line of the raw rationale was lost:\n%s", out)
	}
}
