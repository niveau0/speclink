package golang

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// TestInferKinds pins what the nago recognisers find in the conformant fixture.
//
// The summary line of the command only counts constructs, so a recogniser that
// returns the wrong kind would still add up. This asserts the kinds themselves,
// keyed by name, because that is what forward coverage and every diagnostic
// downstream depend on.
func TestInferKinds(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	got := map[string]ir.ConstructKind{}
	for _, p := range pkgs {
		for _, c := range p.Infer() {
			got[identOf(c.Name)] = c.Kind
		}
	}

	want := map[string]ir.ConstructKind{
		"SubmitQuote":       ConstructUseCase,
		"ApproveQuote":      ConstructUseCase,
		"FindQuoteOverview": ConstructQuery,
		"SubmitQuoteCmd":    ConstructCommand,
		"QuoteSubmitted":    ConstructEvent,
		// Recognised through Evolve rather than Identity: an event sourced
		// aggregate is a plain struct rebuilt by replaying events, and carries
		// no marker of its own.
		"QuoteAggregate": ConstructAggregate,
		// The read model is call driven: its type carries only Clone, so the
		// fact that it is a projection exists solely at construction.
		"QuoteOverview": ConstructProjection,
		// The repository is a named type standing for the framework interface.
		"QuoteRepository": ConstructRepository,
	}
	for name, kind := range want {
		if got[name] != kind {
			t.Errorf("%s: inferred %v, want %v", name, got[name], kind)
		}
	}
}

// TestCoverageObligations pins which construct kinds have to name a
// requirement. The set is a policy decision, and the one it replaces was never
// argued: a query fell out of it by omission, although every architecture rule
// already treats a query as a use case. Reading is a promise too.
func TestCoverageObligations(t *testing.T) {
	must := []ir.ConstructKind{
		ConstructUseCase,
		ConstructQuery,
		ConstructCommand,
		ConstructEvent,
		ConstructProjection,
	}
	for _, k := range must {
		if !k.NeedsRequirement() {
			t.Errorf("%v carries business meaning and must name a requirement", k)
		}
	}

	// Structural kinds are covered through the use case that guards, writes or
	// holds them; demanding a binding for each would only produce noise.
	mustNot := []ir.ConstructKind{
		ConstructAggregate,
		ConstructPermission,
		ConstructRepository,
	}
	for _, k := range mustNot {
		if k.NeedsRequirement() {
			t.Errorf("%v is structural and must not demand a requirement of its own", k)
		}
	}
}

// identOf reduces a fully qualified construct name to its identifier.
func identOf(qualified string) string {
	for i := len(qualified) - 1; i >= 0; i-- {
		if qualified[i] == '.' {
			return qualified[i+1:]
		}
	}
	return qualified
}

// TestAggregateWithoutIdentity guards the recogniser that closed the largest
// blind spot found so far.
//
// data.Aggregate describes a stored entity and announces itself with Identity.
// An event sourced aggregate announces nothing: it is a plain struct rebuilt by
// folding events, and against the reference ERP that meant every single one of
// its sixty aggregates went unseen. What does state the relationship is the
// framework's own event contract — evs.Evt is generic over the aggregate, and
// Evolve names it in the signature.
func TestAggregateWithoutIdentity(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/bad")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./billing/...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	var kinds []ir.ConstructKind
	for _, p := range pkgs {
		for _, c := range p.Infer() {
			if identOf(c.Name) == "Aggregate" {
				kinds = append(kinds, c.Kind)
			}
		}
	}

	if len(kinds) != 1 {
		t.Fatalf("expected the fold target to be recognised once, got %d", len(kinds))
	}
	if kinds[0] != ConstructAggregate {
		t.Errorf("the type an event folds into is an aggregate, got %v", kinds[0])
	}
}

// TestStreamingUseCaseIsAQueryThatReportsErrors covers the shape a lazily
// evaluated read has to take.
//
// `func(auth.Subject) iter.Seq2[Quote, error]` has no trailing error, and until
// this was recognised it was wrong on both counts at once: the signature rule
// demanded an error the type already carries, and the classifier read a use
// case with a single result as a write. Asking for `(iter.Seq2[Quote, error],
// error)` instead would have been asking for a failure channel that can only
// ever be nil — a permission is audited when the sequence is pulled, not when
// it is handed over — and a caller who checks the empty one believes it has
// checked.
func TestStreamingUseCaseIsAQueryThatReportsErrors(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	var kinds []ir.ConstructKind
	for _, p := range pkgs {
		for _, c := range p.Infer() {
			if identOf(c.Name) == "StreamQuoteOverviews" {
				kinds = append(kinds, c.Kind)
			}
		}
	}
	if len(kinds) != 1 {
		t.Fatalf("expected the streaming use case to be recognised once, got %d", len(kinds))
	}
	if kinds[0] != ConstructQuery {
		t.Errorf("a use case yielding a sequence reads; it was classified %v", kinds[0])
	}

	cfg := config.Config{ContextRoot: "app", CmdRoot: "cmd", InfraRoots: []string{"pkg", "foundation"}}
	found := &diag.Set{}
	CheckUseCases(pkgs, cfg, root, DDD1, ir.Waivers{}, found)
	for _, f := range found.Findings() {
		if f.Rule == RuleUCSignature {
			var buf bytes.Buffer
			_ = found.WriteText(&buf)
			t.Fatalf("the sequence already carries the error, yet the signature rule fired:\n%s", buf.String())
		}
	}
}
