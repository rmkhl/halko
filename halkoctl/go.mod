module github.com/rmkhl/halko/halkoctl

go 1.21.0

toolchain go1.24.4

require github.com/rmkhl/halko/types v0.0.0

require github.com/rmkhl/halko/types/log v0.0.0-20250607062522-bc4262653186 // indirect

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
