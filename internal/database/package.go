package database

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/julien-sobczak/deb822"
)

type PackageInfo struct {
	DatabasePath string
	Paragraph    deb822.Paragraph // Extracted package section in /var/lib/dpkg/status

	// info
	Files             []string          // File <name>.list
	MD5sums           map[string]string // File <name>.md5sums
	Conffiles         []string          // File <name>.conffiles
	MaintainerScripts map[string]string // File <name>.{preinst,prerm,postinst,postrm}

	Status      string // Current status (as also present in Paragraph under the field Status)
	StatusDirty bool   // True to ask for sync
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

func (p *PackageInfo) SetStatus(new string) {
	p.Status = new
	p.StatusDirty = true
	// Override in DEB 822 document used to write the status file
	old := p.Paragraph.Values["Status"]
	parts := strings.Split(old, " ")
	p.Paragraph.Values["Status"] = fmt.Sprintf("%s %s %s", parts[0], parts[1], new)
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
			dest := hdr.Name
			if strings.HasPrefix(dest, "./") {
				// ./usr/bin/hello => /usr/bin/hello
				dest = dest[1:]
			}
			if !strings.HasPrefix(dest, "/") {
				// usr/bin/hello => /usr/bin/hello
				dest = "/" + dest
			}

			tmpdest := dest
			if p.isConffile(tmpdest) {
				// Extract using the extension .dpkg-new
				tmpdest += ".dpkg-new"
			}

			if err := os.MkdirAll(filepath.Dir(tmpdest), 0755); err != nil {
				log.Fatalf("Failed to unpack directory %s: %v", tmpdest, err)
			}

			content := buf.Bytes()
			if err := os.WriteFile(tmpdest, content, 0755); err != nil {
				log.Fatalf("Failed to unpack file %s: %v", tmpdest, err)
			}

			p.Files = append(p.Files, dest)
			p.MD5sums[dest] = fmt.Sprintf("%x", md5.Sum(content))
		}
	}

	p.SetStatus("unpacked")
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
	p.SetStatus("half-configured")
	p.Sync()

	// Run maintainer script
	if err := p.runMaintainerScript("postinst"); err != nil {
		return err
	}
	p.SetStatus("installed")
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
	infoPath := filepath.Join(p.DatabasePath, "info")
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
	infoPath := filepath.Join(p.DatabasePath, "info")
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
