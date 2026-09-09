module github.com/worldiety/speclink

go 1.27.0

require (
	github.com/worldiety/speclink/spec v0.0.0
	golang.org/x/mod v0.38.0
	golang.org/x/tools v0.48.0
)

require golang.org/x/sync v0.22.0 // indirect

// The catalogue lives in this repository and is released from it, so the tool
// always builds against the spec it reads. The tool needs it for exactly two
// constants — the marker and the version spec.Verified writes — and duplicating
// those here to avoid the dependency is precisely how the two would stop
// agreeing.
//
// Before the first release the required version must become a real tag
// (spec/vX.Y.Z). A replace applies to the main module only, and `go install
// .../cmd/speclink@version` treats no module as the main one, so it ignores this
// line and would fail to resolve v0.0.0. The replace stays regardless: it is
// what makes a working tree build against its own spec.
replace github.com/worldiety/speclink/spec => ./spec
