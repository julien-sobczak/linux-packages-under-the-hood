package main

import (
	"flag"
)

var flagBuild bool
var flagInstall bool

func init() {
	flag.BoolVar(&flagBuild, "build", false, "Creates a debian archive")
	flag.BoolVar(&flagInstall, "install", false, "Install a debian archive")
	flag.Parse()
}

func main() {
	args := flag.Args()

	if flagBuild {
		build(args)
	} else if flagInstall {
		install(args)
	}

}
