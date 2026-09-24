package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/reqtree"
)

// evidence records which tests actually demonstrated which requirements.
//
// It is the second half of spec.Verified, and the half that makes the first one
// worth anything. Reading the call out of the source proves that somebody wrote
// it down; only running it proves that anything happened. A call can sit behind
// a condition that never holds, or in a test that fails long before reaching
// it, and neither is distinguishable from a working test by any amount of
// static analysis.
//
// So the record here is a record of a moment. Every other entry in
// speclink.lock is a hash of text that can be read again at any time; this one
// cannot be reconstructed from the working tree at all, which is precisely why
// it has to be written down.
//
// It reads what the build wrote rather than running anything. speclink does not
// run tests: the build order is compiler, then speclink, then tests, and a
// command that invoked the suite would either violate that or duplicate it. It
// also makes the evidence something CI hands over rather than something
// speclink produces, which is the right way round for evidence.
//
// Where that comes from differs by language, and the difference is instructive.
// Go writes a line from inside the test, which proves control reached the
// statement — at the cost of putting code in the test. The JVM reads the claim
// from an annotation and the result from the report the build already wrote,
// which costs nothing at all and proves the same thing, because the annotation
// is on the method and a method that passed ran to its end.
//
// Only passing tests are recorded either way. A test that claimed something and
// then failed showed nothing, and recording it would make the failure
// invisible.
func evidence(args []string) error {
	fs := flag.NewFlagSet("evidence", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, holding "+baseline.FileName)
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	in := fs.String("in", "", "`file` holding the test output the frontend reads, e.g. \"go test -json\"; standard input by default")
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	cover := fs.String("coverprofile", "", "`file` written by \"go test -coverprofile\"; records how much of each declaration a run executed")
	dry := fs.Bool("n", false, "report what would be recorded, write nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	// The requirement tree is loaded to turn references back into requirements:
	// the record is bound to the wording a test ran against, and the wording
	// lives in the tree.
	discard := &diag.Set{}
	model, _, p, err := open(absRoot, *cfgPath, *prof, fs.Args(), false)
	if err != nil {
		return err
	}
	reader, ok := model.(lang.EvidenceReader)
	if !ok {
		return fmt.Errorf("profile %s reads no test results, so there is no evidence to record", p.Name)
	}
	source := io.Reader(os.Stdin)
	if *in != "" {
		f, openErr := os.Open(*in)
		if openErr != nil {
			return openErr
		}
		defer f.Close()
		source = f
	}

	tree := reqtree.Build(absRoot, model.Requirements(discard), discard)
	demonstrated, err := reader.Demonstrations(source, treeLookup{tree})
	if err != nil {
		return err
	}

	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}

	// The coverage of a declaration is recorded beside the verifications, and
	// for the same reason: both are what one run showed, and both are worth
	// nothing unless they are tied to the text they were measured against.
	measured, err := recordCoverage(base, model, *cover, discard)
	if err != nil {
		return err
	}

	changed, unknown := check.RecordVerifications(base, tree, demonstrated)
	for _, id := range unknown {
		// A record naming a requirement the tree does not have is a defect
		// worth saying out loud rather than skipping: it means a test is
		// verifying something that has been renamed or removed, and dropping it
		// silently would leave the test looking useful.
		fmt.Fprintf(os.Stderr, "ignored   %s: no such requirement\n", id)
	}
	if len(changed) == 0 && measured == 0 {
		fmt.Fprintln(os.Stderr, "nothing to record; the record already matches this run.")
		return nil
	}
	if measured > 0 {
		fmt.Fprintf(os.Stderr, "measured  %s\n", plural(measured, "declaration", "declarations"))
	}
	for _, line := range changed {
		fmt.Fprintln(os.Stderr, line)
	}
	if *dry {
		fmt.Fprintln(os.Stderr, "\nnothing written (-n).")
		return nil
	}
	if err := base.Save(absRoot); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\nrecorded in %s.\n", baseline.FileName)
	return nil
}

// treeLookup is the one question a frontend has to ask of the tree, narrowed
// so that it does not depend on the whole of it.
type treeLookup struct{ tree *reqtree.Tree }

func (t treeLookup) IDOf(ref string) (string, bool) {
	if r := t.tree.BySymbol(ref); r != nil {
		return r.ID, true
	}
	if r := t.tree.ByID[ref]; r != nil {
		return r.ID, true
	}
	return "", false
}

// recordCoverage attributes a coverage profile to the declarations.
//
// It writes into the same record attest does, because both say something about
// one text: the fingerprint is carried along so that a figure can never outlive
// the declaration it was measured on. A profile taken before a rewrite is not
// evidence about what is there now.
func recordCoverage(base *baseline.File, model lang.Model, profilePath string, discard *diag.Set) (int, error) {
	if profilePath == "" {
		return 0, nil
	}
	inferrer, ok := model.(lang.ConstructInferrer)
	if !ok {
		return 0, nil
	}
	blocks, err := loadCoverProfile(profilePath)
	if err != nil {
		return 0, err
	}

	n := 0
	for _, c := range inferrer.Constructs(discard) {
		if c.Fingerprint == "" {
			continue
		}
		statements, covered := coverageOf(c, blocks)
		if statements == 0 {
			continue
		}

		rec := base.Constructs[c.Name]
		if rec.Fingerprint != c.Fingerprint {
			// A figure about older text is not a figure about this one.
			rec = baseline.Construct{Fingerprint: c.Fingerprint}
		}
		rec.Statements, rec.Covered = statements, covered
		base.Constructs[c.Name] = rec
		n++
	}
	return n, nil
}
