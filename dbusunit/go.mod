module github.com/rmkhl/halko/dbusunit

go 1.23

require (
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/rmkhl/halko/types v0.0.0
	github.com/rmkhl/halko/types/log v0.0.0-20250607062522-bc4262653186
)

replace (
	github.com/rmkhl/halko/types => ../types
	github.com/rmkhl/halko/types/log => ../types/log
)
