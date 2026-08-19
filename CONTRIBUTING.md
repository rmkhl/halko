# Contributing to Halko

## Versioning

Halko is versioned with [semantic versioning](https://semver.org/). The single
source of truth is the `Version` constant in `types/version.go`. Every service
reports it on its status endpoint and logs it at startup, `halkoctl version`
prints it, and the webapp shows it per service on the System page.

### Bump the version in the PR

Decide the bump before opening a pull request and include it in the PR:

| Bump      | When                                                                    |
| --------- | ----------------------------------------------------------------------- |
| **Patch** | Bug fixes only. Nothing about the interface or behaviour changes.        |
| **Minor** | Functionality is added or changed, in a way existing users survive.      |
| **Major** | An API endpoint, a configuration key, or the program format is invalidated. |

Patch is the default: if the change only fixes bugs, bump the patch component.

A "major" change is specifically one that breaks something already deployed —
removing or renaming an API endpoint, requiring a configuration key that older
`halko.cfg` files do not have, or changing the drying program format so that a
stored program stops loading. Adding an optional field to any of the three is a
minor bump, not a major one.

The `Version bump` CI job fails a pull request whose `types/version.go` is
unchanged. For a PR that ships nothing runnable — documentation, CI, editor
config — add the **`no-release`** label and the check is skipped. The check
re-runs whenever the label is added or removed, so adding it to a PR that has
already gone red is enough; there is no need to push a commit to clear it.

### The tag is created for you

Do not create release tags by hand. When a change lands on master, the
`Release tag` workflow reads `types/version.go` and pushes the matching
`vMAJOR.MINOR.PATCH` tag. A merge that did not bump the version finds the tag
already present and does nothing, so the workflow is safe to re-run.

Tags are created after the merge on purpose. A tag pushed from a PR branch
would point at a commit that may never land if the branch is rejected or
force-pushed, and retracting a pushed tag is disruptive for anyone who has
already fetched it.

### Go module tags

The repository has no root `go.mod`, and the services resolve each other
through `replace` directives, so the `vX.Y.Z` tag is a release marker rather
than a Go module version. Per-module tags (`types/v1.0.0` and friends) are
deliberately not created: nothing consumes these modules through the module
proxy.
