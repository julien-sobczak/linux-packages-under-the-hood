package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

/** populateTestDir creates the test files in the existing test directory. */
func populateTestDir(t *testing.T, testdir string, testfiles map[string][]byte) {
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
