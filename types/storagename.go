package types

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrInvalidStorageName is returned for a program or run name that cannot be
// used as a filename.
var ErrInvalidStorageName = errors.New("invalid storage name")

// ValidateStorageName reports whether name may be joined to a storage
// directory to form a file path.
//
// Program and run names arrive over the API and are turned into filenames
// under base_path. filepath.Join cleans the result but does not confine it, so
// a name carrying a separator or a parent reference escapes the directory it
// was meant to land in. Rather than filter the ways of spelling that, the rule
// is a whitelist: a name must be one plain path segment, which is all any
// legitimate name has ever needed to be.
func ValidateStorageName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("%w: name is empty", ErrInvalidStorageName)
	case name == "." || name == "..":
		return fmt.Errorf("%w: %q refers to a directory", ErrInvalidStorageName, name)
	case strings.ContainsRune(name, 0):
		return fmt.Errorf("%w: name contains a NUL byte", ErrInvalidStorageName)
	case strings.ContainsAny(name, `/\`):
		return fmt.Errorf("%w: %q contains a path separator", ErrInvalidStorageName, name)
	case name != filepath.Base(name):
		return fmt.Errorf("%w: %q is not a single path element", ErrInvalidStorageName, name)
	}
	return nil
}
