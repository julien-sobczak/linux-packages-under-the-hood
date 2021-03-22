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

	writer := ar.NewWriter(fdeb)
	if err := writer.WriteGlobalHeader(); err != nil {
		log.Fatal(err)
	}

	// Append debian-binary
	body := "2.0\n"
	hdr := &ar.Header{
		Name: "debian-binary",
		Mode: 0600,
		Uid:  0,
		Gid:  0,
		Size: int64(len(body)),
	}
	if err := writer.WriteHeader(hdr); err != nil {
		log.Fatal(err)
	}
	if _, err := writer.Write([]byte(body)); err != nil {
		log.Fatal(err)
	}

	// Append control.tar
	var bufcontrol bytes.Buffer
	twcontrol := tar.NewWriter(&bufcontrol)
	controlDir := filepath.Join(directory, "DEBIAN")
	err = filepath.Walk(controlDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fmt.Println("Control", path, info.Name())
		hdr := &tar.Header{
			Name: info.Name(),
			Mode: 0650,
			Size: info.Size(),
		}
		if err := twcontrol.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		var content []byte
		content, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := twcontrol.Write(content); err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := twcontrol.Close(); err != nil {
		log.Fatal(err)
	}

	hdr = &ar.Header{
		Name: "control.tar",
		Mode: 0600,
		Uid:  0,
		Gid:  0,
		Size: int64(bufcontrol.Len()),
	}
	if err := writer.WriteHeader(hdr); err != nil {
		log.Fatal(err)
	}
	if _, err := writer.Write(bufcontrol.Bytes()); err != nil {
		log.Fatal(err)
	}

	// Append data.tar
	var bufdata bytes.Buffer
	twdata := tar.NewWriter(&bufdata)
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, filepath.Join(directory, "DEBIAN")) { // Ignore control files
			return nil
		}
		fmt.Println("Data", path, info.Name())
		hdr := &tar.Header{
			Name: info.Name(),
			Uid:  0, // root
			Gid:  0, // root
			Size: info.Size(),
		}
		if err := twdata.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		var content []byte
		content, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := twdata.Write(content); err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := twdata.Close(); err != nil {
		log.Fatal(err)
	}

	hdr = &ar.Header{
		Name: "data.tar",
		Mode: 0600,
		Uid:  0,
		Gid:  0,
		Size: int64(bufdata.Len()),
	}
	if err := writer.WriteHeader(hdr); err != nil {
		log.Fatal(err)
	}
	if _, err := writer.Write(bufdata.Bytes()); err != nil {
		log.Fatal(err)
	}
}
