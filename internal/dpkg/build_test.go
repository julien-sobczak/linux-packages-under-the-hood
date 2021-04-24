package dpkg_test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/blakesmith/ar"
	"github.com/julien-sobczak/linux-packages-from-scratch/internal/dpkg"
	"github.com/julien-sobczak/linux-packages-from-scratch/testutil"
)

func TestBuild(t *testing.T) {
	testdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	t.Logf("Working in temp dir %s", testdir)

	testfiles := map[string][]byte{

		"1.1-1/DEBIAN/control": []byte(`Package: test
Version: 1.1-1
Section: base
Priority: optional
Architecture: all
Maintainer: Julien Sobczak
Description: Test
`),

		"1.1-1/DEBIAN/preinst": []byte(`#!/bin/bash
echo "preinst says hello";
`),

		"1.1-1/usr/bin/test": []byte(`#!/bin/bash
echo "Test";
`),
	}
	testutil.PopulateTestDir(t, testdir, testfiles)

	// Execute the command dpkg --build
	directory := filepath.Join(testdir, "1.1-1")
	dest := filepath.Join(testdir, "test.deb")
	dpkg.Build(directory, dest)

	// Check the resulting package archive
	checkDebianArchive(t, dest, testfiles)
}

/* Test helpers */

/** checkDebianArchive checks the structure of a Debian archive and compare the content of files with the fixtures. */
func checkDebianArchive(t *testing.T, dest string, testfiles map[string][]byte) {
	f, err := os.Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := ar.NewReader(f)

	// Skip debian-binary
	header, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "debian-binary" {
		t.Fatalf("First file must be debian-binary")
	}

	// control.tar
	header, err = reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "control.tar" {
		t.Fatalf("Second file must be control.tar")
	}
	var bufControl bytes.Buffer
	io.Copy(&bufControl, reader)
	checkTarArchive(t, bufControl, testfiles, "1.1-1/DEBIAN/")

	// data.tar
	header, err = reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "data.tar" {
		t.Fatalf("Third file must be data.tar")
	}
	var bufData bytes.Buffer
	io.Copy(&bufData, reader)
	checkTarArchive(t, bufData, testfiles, "1.1-1/")
}

/** checkTarArchive compare the content of files with the fixtures. */
func checkTarArchive(t *testing.T, buf bytes.Buffer, testfiles map[string][]byte, prefix string) {
	tr := tar.NewReader(&buf)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil {
			t.Fatal(err)
		}

		name := hdr.Name
		content, ok := testfiles[prefix+name]
		if !ok {
			continue
		}
		if buf.String() != string(content) {
			t.Errorf("Differences found in %s:\n%v", hdr.Name, diff.CharacterDiff(buf.String(), string(content)))
		}
	}
}
