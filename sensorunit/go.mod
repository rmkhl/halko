module github.com/rmkhl/halko/sensorunit

go 1.24.0

toolchain go1.24.3

require (
	github.com/rmkhl/halko/types v0.0.0-20250925152202-3475d41465c7
	github.com/rmkhl/halko/types/log v0.0.0-20250607062522-bc4262653186
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require golang.org/x/sys v0.39.0 // indirect

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
