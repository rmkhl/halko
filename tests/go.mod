module github.com/rmkhl/halko/tests

go 1.22.0

require (
	github.com/rmkhl/halko/types v0.0.0-00010101000000-000000000000
	github.com/rmkhl/halko/types/log v0.0.0-20250607062522-bc4262653186
)

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
