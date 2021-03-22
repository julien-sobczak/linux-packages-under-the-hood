package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
)

// LIKE tarball_pack
func tarballPack(directory string, filter func(string) bool) ([]byte, error) {
	var bufdata bytes.Buffer
	twdata := tar.NewWriter(&bufdata)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, errParent error) error {
		if info.IsDir() {
			return nil
		}
		if filter != nil && filter(path) {
			return nil
		}
		sep := fmt.Sprintf("%c", filepath.Separator)
		hdr := &tar.Header{
			Name: strings.TrimPrefix(strings.TrimPrefix(path, directory), sep), // Ex: hello/DEBIAN/control => control
			Uid:  0, // root
			Gid:  0, // root
			Mode: 0650,
			Size: info.Size(),
		}
		if err := twdata.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := twdata.Write(content); err != nil {
			log.Fatal(err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if err := twdata.Close(); err != nil {
		return nil, err
	}

	return bufdata.Bytes(), nil
}

// LIKE ar_member_put_file
func arPutFile(w *ar.Writer, name string, body []byte) error {
	hdr := &ar.Header{
		Name: name,
		Uid:  0,
		Gid:  0,
		Mode: 0644,
		Size: int64(len(body)),
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

func build(args []string) {
	if len(args) < 2 {
		log.Fatalf("Missing 'directory' and/or 'dest' arguments")
	}
	directory := args[0]
	dest := args[1]

	// Create the debian archive file
	fdeb, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fdeb.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	writer := ar.NewWriter(fdeb) // LIKE dpkg_ar_create
	if err := writer.WriteGlobalHeader(); err != nil {
		log.Fatal(err)
	}

	// Append debian-binary
	err = arPutFile(writer, "debian-binary", []byte("2.0\n"))
	if err != nil {
		log.Fatal(err)
	}

	// Append control.tar
	controlDir := filepath.Join(directory, "DEBIAN")
	controlTarball, err := tarballPack(controlDir, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = arPutFile(writer, "control.tar", controlTarball)
	if err != nil {
		log.Fatal(err)
	}

	// Append data.tar
	dataDir := directory
	dataTarball, err := tarballPack(dataDir, func(path string) bool {
		return strings.HasPrefix(path, controlDir)
	})
	if err != nil {
		log.Fatal(err)
	}
	err = arPutFile(writer, "data.tar", dataTarball)
	if err != nil {
		log.Fatal(err)
	}
}
