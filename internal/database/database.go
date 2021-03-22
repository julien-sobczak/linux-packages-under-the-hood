package database

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/julien-sobczak/linux-packages-from-scratch/internal/deb822"
)

const (
	Path string = "/var/lib/dpkg"
)

type Directory struct {
	Status   Status
	Packages []PackageInfo
}

type Status struct {
	Content deb822.Document
}

type PackageInfo struct {
	Paragraph deb822.Paragraph // Info present in /var/lib/dpkg/status

	// info
	Files             []string          // File <name>.list
	MD5sums           map[string]string // File <name>.md5sums
	Conffiles         []string          // File <name>.conffiles
	MaintainerScripts map[string]string // File <name>.{preinst,prerm,postinst,postrm}

	Status      string
	StatusDirty bool
}

func (p *PackageInfo) Name() string {
	return p.Paragraph.Value("Package")
}

func (p *PackageInfo) Version() string {
	return p.Paragraph.Value("Version")
}

func (p *PackageInfo) isConffile(path string) bool {
	for _, conffile := range p.Conffiles {
		if path == conffile {
			return true
		}
	}
	return false
}

func (p *PackageInfo) Unpack(buf bytes.Buffer) error {
	if err := p.runMaintainerScript("preinst"); err != nil {
		return err
	}

	fmt.Printf("Unpacking %s (%s) ...\n", p.Name(), p.Version())

	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeReg:
			// ./usr/bin/hello => /usr/bin/hello
			dest := hdr.Name[1:len(hdr.Name)]
			if p.isConffile(dest) {
				// Extract using the extension .dpkg-new
				dest += ".dpkg-new"
			}

			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				log.Fatalf("Failed to unpack directory %s: %v", dest, err.Error())
			}

			outFile, err := os.Create(dest)
			if err != nil {
				log.Fatalf("Failed to unpack file %s: %v", dest, err.Error())
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				log.Fatalf("Failed to unpack file content %s: %v", dest, err.Error())
			}
			outFile.Close()

			// TODO determine md5 and pushes the value into p.MD5sums
		}
	}

	p.Status = "unpacked"
	p.StatusDirty = true
	p.Sync()

	return nil
}

func (p *PackageInfo) Configure() error {
	fmt.Printf("Setting up %s (%s) ...\n", p.Name(), p.Version())

	// Rename conffiles
	for _, conffile := range p.Conffiles {
		// Rename to remove extension .dpkg-new
		tmpConffile := conffile + ".dpkg-new"
		if _, err := os.Stat(tmpConffile); os.IsNotExist(err) {
			return fmt.Errorf("conffile %s is missing", tmpConffile)
		}
		os.Rename(tmpConffile, conffile)
	}
	p.Status = "half-configured"
	p.StatusDirty = true
	p.Sync()

	// Run maintainer script
	if err := p.runMaintainerScript("postinst"); err != nil {
		return err
	}
	p.Status = "installed"
	p.StatusDirty = true
	p.Sync()

	return nil
}

func (p *PackageInfo) runMaintainerScript(name string) error {
	_, ok := p.MaintainerScripts[name]
	if !ok {
		// Nothing to run
		return nil
	}

	out, err := exec.Command("/bin/sh", p.InfoPath(name)).Output()
	if err != nil {
		return err
	}
	fmt.Print(string(out))

	return nil
}

func (p *PackageInfo) PrefixName() string {
	if p.Status == "not-installed" {
		return p.Name()
	}

	// Determine the files prefix currently used
	infoPath := filepath.Join(Path, "info")
	prefix := p.Name()
	if _, err := os.Stat(filepath.Join(infoPath, prefix+".list")); !os.IsNotExist(err) {
		return prefix
	}

	// Try with the arch
	prefix = p.Name() + ":" + p.Paragraph.Value("Architecture")
	if _, err := os.Stat(filepath.Join(infoPath, prefix+".list")); !os.IsNotExist(err) {
		return prefix
	}

	return p.Name()
}

func (p *PackageInfo) InfoPath(filename string) string {
	infoPath := filepath.Join(Path, "info")
	return filepath.Join(infoPath, p.PrefixName()+"."+filename)
}

func (p *PackageInfo) Sync() error {
	// Write <package>.list
	if err := os.WriteFile(p.InfoPath("list"), []byte(FormatList(p.Files)), 0644); err != nil {
		return err
	}

	// Write <package>.md5sums
	if err := os.WriteFile(p.InfoPath("md5sums"), []byte(FormatMD5Sums(p.MD5sums)), 0644); err != nil {
		return err
	}

	// Write <package>.conffiles
	if err := os.WriteFile(p.InfoPath("conffiles"), []byte(FormatConffiles(p.Conffiles)), 0644); err != nil {
		return err
	}

	// Write <package>.{preinst,prerm,postinst,postrm}
	for name, content := range p.MaintainerScripts {
		err := os.WriteFile(p.InfoPath(name), []byte(content), 0755)
		if err != nil {
			return err
		}
	}
	p.StatusDirty = false
	return nil
}

func Load() (*Directory, error) {
	// Load the status file
	statusPath := filepath.Join(Path, "status")
	f, err := os.Open(statusPath)
	if err != nil {
		return nil, err
	}
	parser, err := deb822.NewParser(f)
	if err != nil {
		return nil, err
	}
	statusContent, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	status := Status{
		Content: statusContent,
	}

	// Read the info directory
	infoPath := filepath.Join(Path, "info")
	var packages []PackageInfo
	for _, statusParagraph := range statusContent.Paragraphs {
		name := statusParagraph.Value("Package")
		arch := statusParagraph.Value("Architecture")

		statusField := statusParagraph.Value("Status")
		statusValues := strings.Split(statusField, " ")

		pkg := PackageInfo{
			Paragraph:         statusParagraph,
			MD5sums:           make(map[string]string),
			MaintainerScripts: make(map[string]string),
			Status:            statusValues[2],
			StatusDirty:       false,
		}

		// Determine the files prefix
		prefix := name
		if _, err := os.Stat(filepath.Join(infoPath, prefix+".list")); os.IsNotExist(err) {
			// Try with the arch
			prefix := name + ":" + arch
			if _, err := os.Stat(filepath.Join(infoPath, prefix+".list")); os.IsNotExist(err) {
				continue
			}
		}
		prefixPath := filepath.Join(infoPath, prefix+".")

		// Read list file
		listPath := prefixPath + "list"
		md5sumsPath := prefixPath + "md5sums"
		conffilesPath := prefixPath + "conffiles"
		if _, err := os.Stat(listPath); !os.IsNotExist(err) {
			content, err := os.ReadFile(listPath)
			if err != nil {
				return nil, err
			}
			pkg.Files, err = ParseList(string(content))
			if err != nil {
				return nil, err
			}
		}
		if _, err := os.Stat(md5sumsPath); !os.IsNotExist(err) {
			content, err := os.ReadFile(md5sumsPath)
			if err != nil {
				return nil, err
			}
			pkg.MD5sums, err = ParseMD5Sums(string(content))
			if err != nil {
				return nil, err
			}
		}
		if _, err := os.Stat(conffilesPath); !os.IsNotExist(err) {
			content, err := os.ReadFile(conffilesPath)
			if err != nil {
				return nil, err
			}
			pkg.Conffiles, err = ParseConffiles(string(content))
			if err != nil {
				return nil, err
			}
		}

		// Read maintainer scripts
		for _, script := range []string{"preinst", "postinst", "prerm", "postrm"} {
			scriptPath := prefixPath + script
			if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
				content, err := os.ReadFile(scriptPath)
				if err != nil {
					return nil, err
				}
				pkg.MaintainerScripts[script] = string(content)
			}
		}
		packages = append(packages, pkg)
	}

	return &Directory{
		Status:   status,
		Packages: packages,
	}, nil
}

func (d *Directory) InstalledFiles() int {
	count := 0
	for _, pkg := range d.Packages {
		if pkg.Status == "installed" {
			count += len(pkg.Files)
		}
	}
	return count
}

func (d *Directory) Sync() error {
	newStatus := deb822.Document{
		Paragraphs: []deb822.Paragraph{},
	}

	for _, pkg := range d.Packages {
		newStatus.Paragraphs = append(newStatus.Paragraphs, pkg.Paragraph)

		if pkg.StatusDirty {
			if err := pkg.Sync(); err != nil {
				return err
			}
		}
	}

	os.Rename(filepath.Join(Path, "status"), filepath.Join(Path, "status-old"))
	formatter := deb822.NewFormatter()
	formatter.SetFoldedFields("Description")
	formatter.SetMultilineFields("Conffiles")
	if err := os.WriteFile(filepath.Join(Path, "status"), []byte(formatter.Format(newStatus)), 0644); err != nil {
		return err
	}

	return nil
}

func ParseList(content string) ([]string, error) {
	var files []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		files = append(files, line)
	}
	return files, nil
}

func ParseConffiles(content string) ([]string, error) {
	var files []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		files = append(files, line)
	}
	return files, nil
}

func ParseMD5Sums(content string) (map[string]string, error) {
	ret := make(map[string]string)

	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fileAndChecksum := strings.Split(line, "  ")
		ret[fileAndChecksum[1]] = fileAndChecksum[0]
	}
	return ret, nil
}

func FormatList(files []string) string {
	return strings.Join(files, "\n") + "\n"
}

func FormatMD5Sums(checksums map[string]string) string {
	var sb strings.Builder
	for file, checksum := range checksums {
		sb.WriteString(fmt.Sprintf("%s  %s\n", checksum, file))
	}
	return sb.String()
}

func FormatConffiles(files []string) string {
	return strings.Join(files, "\n") + "\n"
}
