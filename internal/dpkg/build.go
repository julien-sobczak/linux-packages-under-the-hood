package dpkg

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
)

func Build(directory string, dest string) {
	// Create the debian archive file
	fdeb, err := os.Create(dest)
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}
	defer fdeb.Close()

	writer := ar.NewWriter(fdeb)
	if err := writer.WriteGlobalHeader(); err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}

	// Append debian-binary
	err = arPutFile(writer, "debian-binary", []byte("2.0\n"))
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}

	// Append control.tar
	controlDir := filepath.Join(directory, "DEBIAN")
	controlTarball, err := tarballPack(controlDir, nil)
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}
	err = arPutFile(writer, "control.tar", controlTarball)
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}

	// Append data.tar
	dataDir := directory
	dataTarball, err := tarballPack(dataDir, func(path string) bool {
		return strings.HasPrefix(path, controlDir)
	})
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}
	err = arPutFile(writer, "data.tar", dataTarball)
	if err != nil {
		fmt.Printf("dpkg-deb: %s\n", err)
		os.Exit(1)
	}
}

/** arPutFile appends a new file in an ar archive. */
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

/** tarballPack appends every file under directory that passes the filter in the tar archive. */
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
			Uid:  0,                                                            // root
			Gid:  0,                                                            // root
			Mode: 0650,
			Size: info.Size(),
		}
		if err := twdata.WriteHeader(hdr); err != nil {
			fmt.Printf("Error while adding a new header in tarball: %s", err)
			os.Exit(1)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error while reading file %s: %s", path, err)
			os.Exit(1)
		}
		if _, err := twdata.Write(content); err != nil {
			fmt.Printf("Error while adding a new file in tarball: %s", err)
			os.Exit(1)
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
