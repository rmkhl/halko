package main

import (
	"strings"
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestVersionLineReportsSharedVersion(t *testing.T) {
	got := versionLine()

	if !strings.HasPrefix(got, "halkoctl ") {
		t.Errorf("version line is %q, want it to name the tool", got)
	}
	if !strings.HasSuffix(got, types.Version) {
		t.Errorf("version line is %q, want it to end with version %q", got, types.Version)
	}
}
