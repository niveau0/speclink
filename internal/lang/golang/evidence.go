package golang

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/spec"
)

var _ lang.EvidenceReader = (*Model)(nil)

// Demonstrations reads a go test -json stream.
//
// The tree is not consulted: spec.Verified writes requirement IDs, because the
// line is written by the running test and the test knows the requirement it
// holds, not the identifier it was reached through.
func (m *Model) Demonstrations(in io.Reader, _ lang.RequirementIndex) (map[string][]string, error) {
	return readTestOutput(in)
}

// testEvent is the part of the `go test -json` stream this needs.
type testEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
	Output string `json:"Output"`
}

// readTestOutput collects the requirements each passing test demonstrated.
//
// Attribution comes from the stream rather than from the line, which is why
// spec.Verified writes through the test's own logger: go test tags every output
// event with the test that produced it, and without that a record could not be
// tied to a pass or a failure.
func readTestOutput(r io.Reader) (map[string][]string, error) {
	var (
		claimed = map[string][]string{}
		passed  = map[string]bool{}
		seen    bool
	)

	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for s.Scan() {
		var e testEvent
		if err := json.Unmarshal(s.Bytes(), &e); err != nil {
			// Not every line of the stream is ours to understand; a build
			// failure prints plain text into it.
			continue
		}
		seen = true

		switch e.Action {
		case "pass":
			if e.Test != "" {
				passed[e.Test] = true
			}
		case "output":
			if e.Test == "" {
				continue
			}
			ids, err := parseVerifiedLine(e.Output)
			if err != nil {
				return nil, err
			}
			claimed[e.Test] = append(claimed[e.Test], ids...)
		}
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("read test output: %w", err)
	}
	if !seen {
		return nil, errors.New(`no "go test -json" events on the input; pipe "go test -json ./..." into this command, or point -in at its output`)
	}

	out := map[string][]string{}
	for test, ids := range claimed {
		// A test that claimed something and then failed showed nothing. The
		// claim is still in the source, so K14-VERIFICATION-STALE will report
		// it; recording it here would make the failure invisible instead.
		if !passed[test] {
			continue
		}
		for _, id := range ids {
			out[id] = appendUnique(out[id], test)
		}
	}
	return out, nil
}

// parseVerifiedLine extracts the requirement IDs from one output line.
//
// The marker is looked for anywhere in the line, not at its start, because
// testing prefixes its output with the file and line it came from.
func parseVerifiedLine(line string) ([]string, error) {
	i := strings.Index(line, spec.VerifiedMarker)
	if i < 0 {
		return nil, nil
	}
	payload := strings.TrimSpace(line[i+len(spec.VerifiedMarker):])

	var record struct {
		Version int      `json:"v"`
		Reqs    []string `json:"reqs"`
	}
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return nil, fmt.Errorf("unreadable verification line %q: %w", payload, err)
	}
	// The project pins speclink/spec in its go.mod while the developer runs an
	// arbitrary speclink binary, so this is one of the few places where genuine
	// version skew is possible. Refusing is the only safe answer: recording
	// nothing looks exactly like a test that was never written.
	if record.Version != spec.VerifiedVersion {
		return nil, fmt.Errorf("the tests were built against spec.Verified version %d, this speclink reads version %d; align the speclink/spec requirement in go.mod with the binary",
			record.Version, spec.VerifiedVersion)
	}
	return record.Reqs, nil
}
