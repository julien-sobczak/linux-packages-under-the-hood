package dpkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julien-sobczak/deb822"
)

var (
	RootDir string = "/"
	VarDir  string = "/var/lib/dpkg"
)

type Directory struct {
	Status   Status
	Packages []*PackageInfo
}

type Status struct {
	Content deb822.Document
}

func Load() (*Directory, error) {
	// Load the status file
	statusPath := filepath.Join(VarDir, "status")
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
	infoPath := filepath.Join(VarDir, "info")
	var packages []*PackageInfo
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
		packages = append(packages, &pkg)
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

	os.Rename(filepath.Join(VarDir, "status"), filepath.Join(VarDir, "status-old"))
	formatter := deb822.NewFormatter()
	formatter.SetFoldedFields("Description")
	formatter.SetMultilineFields("Conffiles")
	if err := os.WriteFile(filepath.Join(VarDir, "status"), []byte(formatter.Format(newStatus)), 0644); err != nil {
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
