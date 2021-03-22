package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var flagBuild bool

func init() {
	flag.BoolVar(&flagBuild, "build", false, "Creates a debian archive")
	flag.Parse()
}

func main() {
	fmt.Println("dpkg")

	args := flag.Args()

	if flagBuild {
		build(args)
	}

}

func demoTar() {
	// Create the archive file
	fo, err := os.Create("test.tar")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Add some files to the archive.
	tw := tar.NewWriter(fo)
	var files = []struct {
		Name, Body string
	}{
		{"test/readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	///////////////////////

	fi, err := os.Open("test.tar")
	if err != nil {
		log.Fatal(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Open and iterate through the files in the archive.
	tr := tar.NewReader(fi)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, tr); err != nil {
			log.Fatal(err)
		}
		fmt.Println()
	}
}
