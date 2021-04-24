package dpkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/julien-sobczak/deb822"
	"github.com/ulikunitz/xz"
)

func Install(archiveFilepaths []string) {
	// Read the database
	db, err := Load()
	if err != nil {
		fmt.Printf("Unable to read the database: %v", err)
		os.Exit(1)
	}
	fmt.Printf("(Reading database ... %d files and directories currently installed.)\n", db.InstalledFiles())

	// Unpack and configure the archive(s)
	for _, archivePath := range archiveFilepaths {
		err := processArchive(db, archivePath)
		if err != nil {
			fmt.Printf("dpkg-deb: error: %s\n", err)
			fmt.Printf("Errors were encountered while processing:\n\t%s\n", archivePath)
		}
	}
}

func processArchive(db *Directory, archivePath string) error {
	// Read the debian archive file
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := ar.NewReader(f)

	// Skip debian-binary
	_, err = reader.Next()
	if err != nil {
		return err
	}

	// control.tar
	header, err := reader.Next()
	if err != nil {
		return err
	}
	var bufControl bytes.Buffer
	err = extractTar(header.Name, &bufControl, reader)
	if err != nil {
		return err
	}

	pkg, err := ParseControl(db, bufControl)
	if err != nil {
		return err
	}

	// Add new package in database
	db.Packages = append(db.Packages, pkg)
	db.Sync()

	// data.tar
	header, err = reader.Next()
	if err != nil {
		return err
	}
	var bufData bytes.Buffer
	err = extractTar(header.Name, &bufData, reader)
	if err != nil {
		return err
	}

	fmt.Printf("Preparing to unpack %s ...\n", filepath.Base(archivePath))

	if err := pkg.Unpack(bufData); err != nil {
		return err
	}
	if err := pkg.Configure(); err != nil {
		return err
	}

	db.Sync()

	return nil
}

func extractTar(filename string, writer io.Writer, reader io.Reader) error {
	if strings.HasSuffix(filename, ".gz") {
		gzf, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		reader = gzf
	} else if strings.HasSuffix(filename, ".xz") {
		xzf, err := xz.NewReader(reader)
		if err != nil {
			return err
		}
		reader = xzf
	}
	io.Copy(writer, reader)
	return nil
}

func ParseControl(db *Directory, buf bytes.Buffer) (*PackageInfo, error) {

	pkg := PackageInfo{
		MD5sums:           make(map[string]string),
		MaintainerScripts: make(map[string]string),
		Status:            "not-installed",
		StatusDirty:       true,
	}

	tr := tar.NewReader(&buf)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil {
			return nil, err
		}

		switch filepath.Base(hdr.Name) {
		case "control":
			parser, err := deb822.NewParser(strings.NewReader(buf.String()))
			if err != nil {
				return nil, err
			}
			document, err := parser.Parse()
			if err != nil {
				return nil, err
			}
			controlParagraph := document.Paragraphs[0]

			// Copy control fields and add the Status field in second position
			pkg.Paragraph = deb822.Paragraph{
				Values: make(map[string]string),
			}
			// Make sure the field Package comes first
			pkg.Paragraph.Order = append(pkg.Paragraph.Order, "Package", "Status")
			pkg.Paragraph.Values["Package"] = controlParagraph.Value("Package")
			pkg.Paragraph.Values["Status"] = "install ok non-installed"
			// Add remaining ordered fields
			for _, field := range controlParagraph.Order {
				if field == "Package" {
					continue
				}
				pkg.Paragraph.Order = append(pkg.Paragraph.Order, field)
				pkg.Paragraph.Values[field] = controlParagraph.Value(field)
			}
		case "conffiles":
			pkg.Conffiles, err = ParseConffiles(buf.String())
			if err != nil {
				return nil, err
			}
		case "md5sums":
			pkg.MD5sums, err = ParseMD5Sums(buf.String())
			if err != nil {
				return nil, err
			}
		case "prerm":
			fallthrough
		case "preinst":
			fallthrough
		case "postinst":
			fallthrough
		case "postrm":
			pkg.MaintainerScripts[filepath.Base(hdr.Name)] = buf.String()
		}
	}

	return &pkg, nil
}
