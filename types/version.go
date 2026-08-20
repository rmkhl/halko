package types

// Version is the Halko release version.
//
// This constant is the single source of truth for the version. Every service
// reports it on its status endpoint and logs it at startup, and the
// release-tag workflow creates the matching git tag once a change lands on
// master. Releasing therefore means bumping this line -- see CONTRIBUTING.md
// for which component to bump.
const Version = "1.1.0"
