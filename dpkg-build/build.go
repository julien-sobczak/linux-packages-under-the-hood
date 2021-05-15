package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
)

func main() {
	// This program expects two arguments:
	// - The directory following the resources to package in the archive.
	// - The name of the output .deb file
	if len(os.Args) < 3 {
		log.Fatalf("Missing 'directory' and/or 'dest' arguments.")
	}

	directory := os.Args[1]
	dest := os.Args[2]

	// Create the Debian archive file
	fdeb, _ := os.Create(dest)
	defer fdeb.Close()

	// A Debian package is an archive using the AR format.
	// We use an external Go module to create the archive.
	// as the standard library does not support it but supports
	// the tar format that will be used for the control and data files.

	writer := ar.NewWriter(fdeb)
	writer.WriteGlobalHeader()

	// A Debian package contains 3 files that must be
	// added in a precise order.
	// We use two utility functions that will be defined later:
	// - arPutFile is a wrapper around the library to add an entry.
	// - tarballPack creates a tarball using the Go

	// Append debian-binary
	arPutFile(writer, "debian-binary", []byte("2.0\n"))

	// Append control.tar
	controlDir := filepath.Join(directory, "DEBIAN")
	controlTarball := tarballPack(controlDir, nil)
	arPutFile(writer, "control.tar", controlTarball)

	// Append data.tar
	dataDir := directory
	dataTarball := tarballPack(dataDir, func(path string) bool {
		// Filter DEBIAN/ files
		return strings.HasPrefix(path, controlDir)
	})
	arPutFile(writer, "data.tar", dataTarball)
}


// arPutFile adds a new entry in a AR archive.
func arPutFile(w *ar.Writer, name string, body []byte) {
	hdr := &ar.Header{
		Name: name,
		Mode: 0600,
		Uid:  0,
		Gid:  0,
		Size: int64(len(body)),
	}
	w.WriteHeader(hdr)
	w.Write(body)
}

// tarballPack traverses a local directory to add all files under it
// into a tarball.
func tarballPack(directory string, filter func(string) bool) []byte {
	var bufdata bytes.Buffer
	twdata := tar.NewWriter(&bufdata)
	filepath.Walk(
		directory,
		func(path string, info os.FileInfo, errParent error) error {
			if info.IsDir() {
				return nil
			}
			if filter != nil && filter(path) {
				return nil
			}
			sep := fmt.Sprintf("%c", filepath.Separator)
			name := strings.TrimPrefix(strings.TrimPrefix(path, directory), sep)
			hdr := &tar.Header{
				Name: name,
				Uid:  0, // root
				Gid:  0, // root
				Mode: 0650,
				Size: info.Size(),
			}
			twdata.WriteHeader(hdr)
			content, _ := ioutil.ReadFile(path)
			twdata.Write(content)

			return nil
		})
	twdata.Close()

	return bufdata.Bytes()
}
