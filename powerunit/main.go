package main

import (
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/router"
	"github.com/rmkhl/halko/powerunit/shelly"
)

func main() {
	s := shelly.New("insert-addr-here")
	p := power.New(s)
	r := router.New(p)

	r.Run()
}
