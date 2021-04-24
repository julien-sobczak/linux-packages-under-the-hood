package main

import (
	"flag"

	"github.com/julien-sobczak/linux-packages-from-scratch/internal/apt"
)

func main() {
	var flagInstall bool
	flag.BoolVar(&flagInstall, "install", false, "Install a debian package")
	flag.Parse()
	args := flag.Args()

	if flagInstall {
		apt.Install(args)
	}

}
