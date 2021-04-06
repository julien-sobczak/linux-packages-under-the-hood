package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/julien-sobczak/deb822"
	"github.com/julien-sobczak/linux-packages-from-scratch/internal/database"
)

func install(dbdir string, args []string) {
	if len(args) < 1 {
		log.Fatalf("Missing package archive(s)")
	}

	// Read the database
	db, err := database.Load(dbdir)
	if err != nil {
		log.Fatalf("Unable to read the database: %v", err)
	}
	fmt.Printf("(Reading database ... %d files and directories currently installed.)", db.InstalledFiles())

	// Unpack and configure the archive(s)
	for _, archivePath := range args {
		err := processArchive(db, archivePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func processArchive(db *database.Directory, archivePath string) error {

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
	_, err = reader.Next()
	if err != nil {
		return err
	}
	var bufControl bytes.Buffer
	io.Copy(&bufControl, reader)

	pkg, err := parseControl(db, bufControl)
	if err != nil {
		return err
	}

	// Add new package in database
	db.Packages = append(db.Packages, pkg)
	db.Sync()

	// data.tar
	_, err = reader.Next()
	if err != nil {
		return err
	}
	var bufData bytes.Buffer
	io.Copy(&bufData, reader)

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

func parseControl(db *database.Directory, buf bytes.Buffer) (*database.PackageInfo, error) {

	pkg := database.PackageInfo{
		DatabasePath:      db.Path,
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
			pkg.Conffiles, err = database.ParseConffiles(buf.String())
			if err != nil {
				return nil, err
			}
		case "md5sums":
			pkg.MD5sums, err = database.ParseMD5Sums(buf.String())
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
