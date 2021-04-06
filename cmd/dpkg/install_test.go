package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
)

func TestInstall(t *testing.T) {
	testdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(testdir)
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
`),
		"1.1-1/DEBIAN/preinst": []byte(`#!/bin/bash
echo "preinst says hello";
`),
		// Make sure all files end up under the temp directory!
		fmt.Sprintf("1.1-1/%s/out/test", testdir): []byte(`#!/bin/bash
echo "Test";
`),

		// A partial Dpkg database containing only vim
		"dpkg/status": []byte(`Package: vim
Status: install ok installed
Priority: optional
Section: editors
Installed-Size: 3286
Maintainer: Debian Vim Maintainers <team+vim@tracker.debian.org>
Architecture: amd64
Version: 2:8.2.2434-3
Provides: editor
Depends: vim-common (= 2:8.2.2434-3), vim-runtime (= 2:8.2.2434-3), libacl1 (>= 2.2.23), libc6 (>= 2.29), libgpm2 (>= 1.20.7), libselinux1 (>= 3.1~), libtinfo6 (>= 6)
Suggests: ctags, vim-doc, vim-scripts
Description: Vi IMproved - enhanced vi editorVim is an almost compatible version of the UNIX editor Vi.
 .
 Many new features have been added: multi level undo, syntax
 highlighting, command line history, on-line help, filename
 completion, block operations, folding, Unicode support, etc.
 .
 This package contains a version of vim compiled with a rather
 standard set of features.  This package does not provide a GUI
 version of Vim.  See the other vim-* packages if you need more
 (or less).
Homepage: https://www.vim.org/
`),
		"dpkg/info/vim.list": []byte(`
root@bullseye:/var/lib/dpkg# cat info/vim.*
/.
/usr
/usr/bin
/usr/bin/vim.basic
/usr/share
/usr/share/bug
/usr/share/bug/vim
/usr/share/bug/vim/presubj
/usr/share/bug/vim/script
/usr/share/doc
/usr/share/doc/vim
/usr/share/doc/vim/NEWS.Debian.gz
/usr/share/doc/vim/changelog.Debian.gz
/usr/share/doc/vim/changelog.gz
/usr/share/doc/vim/copyright
/usr/share/lintian
/usr/share/lintian/overrides
/usr/share/lintian/overrides/vim
`),
		"dpkg/info/vim.md5sums": []byte(`
4a3f9f0ca96f401e54f58a7bec8b659c  usr/bin/vim.basic
261cbb1ba18f3e1e0f45719f8cca0e57  usr/share/bug/vim/presubj
3bf849d905a93ddc7f7703774131a8f2  usr/share/bug/vim/script
251c75666e30e95a83be1eb1f2f614d0  usr/share/doc/vim/NEWS.Debian.gz
7f6ecc6270f2c1649113bb1a220114c0  usr/share/doc/vim/changelog.Debian.gz
1dc9356b84c20c365b2642ea9b79acee  usr/share/doc/vim/changelog.gz
40370ce9e3c7c0b9a4baabbc022d7206  usr/share/doc/vim/copyright
1f910419a870592b0ce7c6711bc8a54f  usr/share/lintian/overrides/vim
`),
		"dpkg/info/vim.postint": []byte(`
#!/bin/sh
set -e

pkg=vim
variant=basic
mandir=/usr/share/man

# add /usr/bin/vim.variant as alternative for /usr/bin/vim. Priority are
# chosen accordingly to the principle: more features, higher priority

add_variant_alternative () {
  if [ "$variant" != "tiny" ]; then
    update-alternatives --install /usr/bin/vim vim /usr/bin/vim.$variant $1
    update-alternatives --install /usr/bin/vimdiff vimdiff /usr/bin/vim.$variant $1
    update-alternatives --install /usr/bin/rvim rvim /usr/bin/vim.$variant $1
  fi
  update-alternatives --install /usr/bin/rview rview /usr/bin/vim.$variant $1
  # Since other packages provide these commands, we'll setup alternatives for
  # their manpages, too.
  for i in vi view ex editor ; do
    update-alternatives \
      --install /usr/bin/$i $i /usr/bin/vim.$variant $1 \
      --slave $mandir/da/man1/$i.1.gz $i.da.1.gz $mandir/da/man1/vim.1.gz \
      --slave $mandir/de/man1/$i.1.gz $i.de.1.gz $mandir/de/man1/vim.1.gz \
      --slave $mandir/fr/man1/$i.1.gz $i.fr.1.gz $mandir/fr/man1/vim.1.gz \
      --slave $mandir/it/man1/$i.1.gz $i.it.1.gz $mandir/it/man1/vim.1.gz \
      --slave $mandir/ja/man1/$i.1.gz $i.ja.1.gz $mandir/ja/man1/vim.1.gz \
      --slave $mandir/pl/man1/$i.1.gz $i.pl.1.gz $mandir/pl/man1/vim.1.gz \
      --slave $mandir/ru/man1/$i.1.gz $i.ru.1.gz $mandir/ru/man1/vim.1.gz \
      --slave $mandir/man1/$i.1.gz $i.1.gz \
              $mandir/man1/vim.1.gz
  done
  case "$variant" in
    gtk|gtk3|athena) # gui enabled variants
      add_gui_variant_alternative $1
      ;;
  esac
}

add_gui_variant_alternative () {
  for i in gvim gview rgview rgvim evim eview gvimdiff ; do
    update-alternatives --install /usr/bin/$i $i /usr/bin/vim.$variant $1
  done
}

case "$pkg" in
  vim-tiny)
    add_variant_alternative 15
    ;;
  vim)
    add_variant_alternative 30
    ;;
  vim-nox)
    add_variant_alternative 40
    ;;
  vim-gtk|vim-gtk3|vim-athena)
    add_variant_alternative 50
    ;;
esac

# Automatically added by dh_installdeb/13.3.3
dpkg-maintscript-helper symlink_to_dir /usr/share/doc/vim vim-common 2:8.0.1451-1\~ vim -- "$@"
# End automatically added section


exit 0
`),
	}
	populateTestDir(t, testdir, testfiles)

	// Create the test.deb to install
	pkgdir := filepath.Join(testdir, "1.1-1")
	dest := filepath.Join(testdir, "test.deb")
	build([]string{pkgdir, dest})

	// Install the package
	dbdir := filepath.Join(testdir, "dpkg")
	install(dbdir, []string{dest})

	// Check the database
	checkFileContains(t, filepath.Join(dbdir, "status"), `Package: vim
Status: install ok installed
Priority: optional
Section: editors
Installed-Size: 3286
Maintainer: Debian Vim Maintainers <team+vim@tracker.debian.org>
Architecture: amd64
Version: 2:8.2.2434-3
Provides: editor
Depends: vim-common (= 2:8.2.2434-3), vim-runtime (= 2:8.2.2434-3), libacl1 (>= 2.2.23), libc6 (>= 2.29), libgpm2 (>= 1.20.7), libselinux1 (>= 3.1~), libtinfo6 (>= 6)
Suggests: ctags, vim-doc, vim-scripts
Description: Vi IMproved - enhanced vi editorVim is an almost compatible version of the UNIX editor Vi.
 .
 Many new features have been added: multi level undo, syntax
 highlighting, command line history, on-line help, filename
 completion, block operations, folding, Unicode support, etc.
 .
 This package contains a version of vim compiled with a rather
 standard set of features.  This package does not provide a GUI
 version of Vim.  See the other vim-* packages if you need more
 (or less).
Homepage: https://www.vim.org/

Package: test
Status: install ok installed
Version: 1.1-1
Section: base
Priority: optional
Architecture: all
Maintainer: Julien Sobczak
Description: Test
`)
	checkFileContains(t, filepath.Join(dbdir, "info/test.md5sums"), fmt.Sprintf(`a6101e71800b523cb9533b074387d693  %s/out/test
`, testdir))
	checkFileContains(t, filepath.Join(dbdir, "info/test.list"), fmt.Sprintf(`%s/out/test
`, testdir))
	checkFileContains(t, filepath.Join(dbdir, "info/test.preinst"), `#!/bin/bash
echo "preinst says hello";
`)
	// Check unpacked files
	checkFileContains(t, filepath.Join(testdir, "out/test"), `#!/bin/bash
echo "Test";
`)
}

/* Test helpers */

func checkFileContains(t *testing.T, path string, content string) {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("Found differences in file %s:\n%s", path, diff.CharacterDiff(string(data), content))
	}
}
