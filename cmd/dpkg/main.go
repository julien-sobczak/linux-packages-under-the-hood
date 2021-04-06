package main

import (
	"flag"
)

func main() {
	var flagBuild bool
	var flagInstall bool
	flag.BoolVar(&flagBuild, "build", false, "Creates a debian archive")
	flag.BoolVar(&flagInstall, "install", false, "Install a debian archive")
	flag.Parse()
	args := flag.Args()

	if flagBuild {
		build(args)
	} else if flagInstall {
		install("/var/lib/dpkg", args)
	}

}
