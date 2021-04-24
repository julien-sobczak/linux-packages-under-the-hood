package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
)

/** PopulateTestDir creates the test files in the existing test directory. */
func PopulateTestDir(t *testing.T, testdir string, testfiles map[string][]byte) {
	for file, content := range testfiles {
		dir := filepath.Dir(file)
		if err := os.MkdirAll(filepath.Join(testdir, dir), 0755); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(testdir, file), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// CheckFileContains checks the content of a single file.
func CheckFileContains(t *testing.T, path string, content string) {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("Found differences in file %s:\n%s", path, diff.CharacterDiff(string(data), content))
	}
}

// CheckFileExists checks the presence of a single file.
// The path argument can contains glob patterns.
func CheckFileExists(t *testing.T, path string) {
	matches, err := filepath.Glob(path)
	if err != nil {
		t.Errorf("Globbing error in path expression %s", path)
		return
	}
	if len(matches) == 0 {
		t.Errorf("Missing file matching %s", path)
	}
}
