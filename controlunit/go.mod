module github.com/rmkhl/halko/controlunit

go 1.26.2

require github.com/rmkhl/halko/types v0.0.0

require (
	github.com/gorilla/websocket v1.5.3
	github.com/rmkhl/halko/types/log v0.0.0
)

replace github.com/rmkhl/halko/types => ../types

replace github.com/rmkhl/halko/types/log => ../types/log
