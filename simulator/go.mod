module github.com/rmkhl/halko/simulator

go 1.23.0

toolchain go1.24.3

require (
	github.com/rmkhl/halko/types v0.0.0-20250925152202-3475d41465c7
	github.com/rmkhl/halko/types/log v0.0.0-20250607062522-bc4262653186
)

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
