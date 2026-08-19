package types

import (
	"regexp"
	"testing"
)

// The release workflow turns Version straight into a git tag, so it has to be
// a plain MAJOR.MINOR.PATCH with no leading zeros and no pre-release suffix.
var semverPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

func TestVersionIsSemver(t *testing.T) {
	if !semverPattern.MatchString(Version) {
		t.Errorf("Version is %q, want MAJOR.MINOR.PATCH", Version)
	}
}
