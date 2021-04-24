package apt_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/julien-sobczak/linux-packages-from-scratch/internal/apt"
	"github.com/julien-sobczak/linux-packages-from-scratch/internal/dpkg"
	"github.com/julien-sobczak/linux-packages-from-scratch/testutil"
)

// NOTE: These tests are not expected to continue working in the future.
// They are downloading files from Debian repositories which are going to change with future releases.
// The tests were written when the latest stable release of Debian was the version 10, named Buster.

func TestInstall(t *testing.T) {
	testdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testdir)
	t.Logf("Working in temp dir %s", testdir)

	testfiles := map[string][]byte{

		// Some files to create a test Debian archive
		"1.1-1/DEBIAN/control": []byte(`Package: test
Version: 1.1-1
Section: base
Priority: optional
Architecture: all
Maintainer: Julien Sobczak
Description: Test
Depends: cowsay
`),
		"1.1-1/usr/bin/test": []byte(`#!/bin/bash
echo "Test" | /usr/games/cowsay;
`),

		// A partial Dpkg database containing cowsay dependencies
		"/var/lib/dpkg/status": []byte(`Package: libtext-charwidth-perl
Status: install ok installed
Priority: optional
Section: perl
Installed-Size: 43
Maintainer: Anibal Monsalve Salazar <anibal@debian.org>
Architecture: amd64
Source: libtext-charwidth-perl (0.04-7.1)
Version: 0.04-7.1+b1
Depends: libc6 (>= 2.2.5), perl-base (>= 5.28.0-3), perlapi-5.28.0
Description: get display widths of characters on the terminal
 This module permits perl software to get the display widths of characters
 and strings on the terminal, using wcwidth() and wcswidth() from libc.
 .
 It provides mbwidth(), mbswidth(), and mblen().
Homepage: http://search.cpan.org/search?module=Text::CharWidth

Package: perl
Status: install ok installed
Priority: standard
Section: perl
Architecture: amd64
Multi-Arch: allowed
Version: 5.28.1-6+deb10u1
`),

		// A list of sources containing only the buster repository
		"/etc/apt/sources.list": []byte(`
deb http://deb.debian.org/debian buster main
deb-src http://deb.debian.org/debian buster main
`),

		// The GPG public key to validate the repository
		"/etc/apt/trusted.gpg.d/debian-archive-buster-stable.gpg": testdata(t, "../../testdata/debian-archive-buster-stable.gpg"),

		// Lock files to initialize common APT directories
		"/var/lib/apt/lists/lock": []byte(``),
		"/var/cache/apt/archives/lock": []byte(``),
	}
	testutil.PopulateTestDir(t, testdir, testfiles)

	// Create the archive test.deb
	pkgdir := filepath.Join(testdir, "1.1-1")
	testArchive := filepath.Join(testdir, "test.deb")
	dpkg.VarDir = filepath.Join(testdir, "/var/lib/dpkg/")
	dpkg.RootDir = testdir
	dpkg.Build(pkgdir, testArchive)

	// Install the package
	apt.EtcDir = filepath.Join(testdir, "/etc/apt")
	apt.VarDir = filepath.Join(testdir, "/var/lib/apt")
	apt.CacheDir = filepath.Join(testdir, "/var/cache/apt")
	apt.Install([]string{testArchive})

	// Check that the APT cache has been uploaded
	testutil.CheckFileExists(t, filepath.Join(testdir, "/var/cache/apt/archives/cowsay_*_all.deb"))
	testutil.CheckFileExists(t, filepath.Join(testdir, "/var/cache/apt/archives/test_1.1-1_all.deb"))

	// Check that the packages have been installed
	testutil.CheckFileExists(t, filepath.Join(testdir, "/usr/games/cowsay")) // Unpacked from cowsay
	testutil.CheckFileExists(t, filepath.Join(testdir, "/usr/bin/test")) // Unpacked from test
}

/* Test Helpers */

func testdata(t *testing.T, filename string) []byte {

	r, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Error opening test file: %s", err)
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Error reading test file: %s", err)
	}

	return data
}
