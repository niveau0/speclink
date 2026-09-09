module github.com/worldiety/speclink

go 1.27.0

// The catalogue is required as a published module, not replaced with the
// directory next to it, and that is not a preference.
//
// `go install .../cmd/speclink@version` refuses a module whose go.mod carries
// any replace directive — it must be interpretable exactly as if it were the
// main module. One replace here therefore costs the only distribution channel
// this tool has. What makes a working tree build against its own spec instead
// is go.work, which go install does not read.
//
// The consequence to keep in mind: the version below is what a released binary
// is built against. Change something in spec/ and the pseudo-version has to be
// raised after pushing, or an installed speclink reads a spec that no longer
// matches the one projects are pinning. That is the version skew DumpVersion and
// VerifiedVersion exist to catch rather than to prevent.
require (
	github.com/worldiety/speclink/spec v0.0.0-20260909124605-a91968ddccb4
	golang.org/x/mod v0.38.0
	golang.org/x/tools v0.48.0
)

require golang.org/x/sync v0.22.0 // indirect
