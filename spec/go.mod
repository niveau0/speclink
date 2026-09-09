// The directive catalogue is a module of its own, and that is not tidiness.
//
// A target project imports this package and nothing else of speclink: the
// annotation and requirement files are ordinary Go and are part of its normal
// build. The tool around it is a compiler frontend and depends on
// golang.org/x/tools and golang.org/x/mod, which have no business in the
// module graph of an application that only writes down what it promises.
//
// The boundary was already described before it existed: DumpVersion and
// VerifiedVersion exist precisely because the project pins this package while
// the developer runs an arbitrary speclink binary. This makes it the version
// boundary it was documented as.
//
// Nothing here imports anything outside the standard library, and nothing here
// may start to.
module github.com/worldiety/speclink/spec

go 1.27.0
