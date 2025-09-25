module github.com/rmkhl/halko/sensorunit

go 1.23.0

toolchain go1.24.3

require (
	github.com/rmkhl/halko/types v0.0.0-20250607062522-bc4262653186
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require golang.org/x/sys v0.33.0 // indirect

replace github.com/rmkhl/halko/types => ../types
