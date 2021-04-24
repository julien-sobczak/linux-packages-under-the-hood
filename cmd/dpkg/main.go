package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/julien-sobczak/linux-packages-from-scratch/internal/dpkg"
)

func main() {
	var flagBuild bool
	var flagInstall bool
	flag.BoolVar(&flagBuild, "build", false, "Creates a debian archive")
	flag.BoolVar(&flagInstall, "install", false, "Install a debian archive")
	flag.Parse()
	args := flag.Args()

	if flagBuild {
		if len(args) < 2 {
			fmt.Printf("Missing 'directory' and/or 'dest' arguments\n")
			os.Exit(1)
		}
		directory := args[0]
		dest := args[1]
		dpkg.Build(directory, dest)
	} else if flagInstall {
		if len(args) < 1 {
			fmt.Printf("Missing package archive(s)\n")
			os.Exit(1)
		}
		dpkg.Install(args)
	}

}
