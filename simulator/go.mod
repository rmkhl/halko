module github.com/rmkhl/halko/simulator

go 1.26.2

require (
	github.com/rmkhl/halko/types v0.0.0
	github.com/rmkhl/halko/types/log v0.0.0
	golang.org/x/sys v0.39.0
)

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
