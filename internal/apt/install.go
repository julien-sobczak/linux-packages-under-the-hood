package apt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/julien-sobczak/deb822"
	"github.com/julien-sobczak/linux-packages-from-scratch/internal/dpkg"
)

func Install(args []string) {
	var pkgNames []string
	if len(args) >= 1 {
		pkgNames = append(pkgNames, args...)
	}

	// Load the Cache
	cache := &CacheFile{}
	cache.Open()

	// Search for requested packages to install
	pkgs := make(map[string]*Package)
	for _, pkgName := range pkgNames {
		if strings.HasSuffix(pkgName, ".deb") { // Archive not in cache
			pkg, err := registerPackage(cache, pkgName)
			if err != nil {
				fmt.Printf("E: Unable to locate package %s\n", pkgName)
				os.Exit(1)
			}
			pkgName = pkg.Name()
		}
		cache.MarkForInstallation(pkgName)
		pkgs[pkgName] = cache.GetPackage(pkgName)
	}

	// Print out the list of additional packages to install
	if cache.InstCount() != len(pkgNames) {
		var extras []string
		for _, pkg := range cache.GetPackages() {
			state := cache.GetState(pkg)
			if !state.Install() {
				continue
			}
			if _, ok := pkgs[pkg.Name()]; !ok {
				extras = append(extras, pkg.Name())
			}
		}
		fmt.Printf("The following additional packages will be installed:\n\t%s\n", strings.Join(extras, " "))
	}

	// Print out the list of suggested packages
	var suggests []string
	for _, pkg := range cache.GetPackages() {
		state := cache.GetState(pkg)

		// Just look at the ones we want to install */
		if !state.Install() {
			continue
		}

		// get the suggestions for the candidate ver
		for _, dependency := range pkg.Suggests() {
			suggests = append(suggests, dependency.Name)
		}
	}
	if len(suggests) > 0 {
		fmt.Printf("Suggested packages:\n\t%s\n", strings.Join(suggests, " "))
	}

	err := InstallPackages(cache)
	if err != nil {
		fmt.Printf("E: %s\n", err)
		os.Exit(1)
	}
}

type pkgIndexFile struct {
	doc deb822.Document // Content of the Packages file
}

type Package struct {
	doc    deb822.Paragraph
	source *pkgSource

	local         bool
	localFilepath string
	cacheFilepath string
}

func (p *Package) Name() string {
	return p.doc.Value("Package")
}

func (p *Package) Version() string {
	return p.doc.Value("Version")
}

func (p *Package) Architecture() string {
	return p.doc.Value("Architecture")
}

func (p *Package) Depends() []Dependency {
	return ParseDependencies(p.doc.Value("Depends"))
}

func (p *Package) Suggests() []Dependency {
	return ParseDependencies(p.doc.Value("Suggests"))
}

type Dependency struct {
	Name     string
	Version  string
	Relation string
}

func (d Dependency) String() string {
	res := d.Name
	if d.Version != "" {
		res += fmt.Sprintf(" (%s %s)", d.Relation, d.Version)
	}
	return res
}

func ParseDependencies(values string) []Dependency {
	depsValues := strings.TrimSpace(values)
	if depsValues == "" {
		return nil
	}

	var deps []Dependency
	for _, value := range strings.Split(depsValues, ", ") {
		deps = append(deps, ParseDependency(value))
	}
	return deps
}

func ParseDependency(value string) Dependency {
	// Example of syntax:
	// "adduser", "gpgv | gpgv2", "libc6 (>= 2.15)", "python3:any (>= 3.5~)", "foo [i386]", "perl:any", "perlapi-5.28.0"

	var dep Dependency

	r := regexp.MustCompile(`^(?P<name>[\w\.-]+)(?:[:]\w+)?(?: [(](?P<relation>(?:>>|>=|=|<=|<<)) (?P<version>\S+)[)])?(?: [|].*)?$`)
	res := r.FindStringSubmatch(value)
	names := r.SubexpNames()
	for i, _ := range res {
		switch names[i] {
		case "name":
			dep.Name = res[i]
		case "relation":
			dep.Relation = res[i]
		case "version":
			dep.Version = res[i]
		}
	}
	return dep
}

func InstallPackages(cache *CacheFile) error {
	acq := NewPkgAcquire(cache)

	// Download package archive
	for _, pkgName := range cache.depCache.order {
		pkg := cache.GetPackage(pkgName)
		if pkg.local {
			// Copy archive to /var/cache/apt/archives/
			src, err := os.Open(pkg.localFilepath)
			if err != nil {
				return err
			}
			defer src.Close()

			// Ex: rsync_3.1.3-6_amd64.deb
			destFilename := fmt.Sprintf("%s_%s_%s.deb", pkg.Name(), pkg.Version(), pkg.Architecture())
			destFilepath := filepath.Join(CacheDir, "archives", destFilename)
			dest, err := os.Create(destFilepath)
			if err != nil {
				return err
			}
			defer dest.Close()

			_, err = io.Copy(dest, src)
			if err != nil {
				return err
			}

			pkg.cacheFilepath = destFilepath
		} else {
			acq.Add(NewPackageItem(pkg))
		}
	}
	err := acq.Run()
	if err != nil {
		return err
	}

	// Run dpkg
	var archives []string
	for _, pkgName := range cache.depCache.order {
		pkg := cache.GetPackage(pkgName)
		archives = append(archives, pkg.cacheFilepath)
	}
	dpkg.Install(archives)

	return nil
}

func registerPackage(cache *CacheFile, archivePath string) (*Package, error) {
	// Code is inspired by dpkg Golang implementation

	db, err := dpkg.Load()
	if err != nil {
		return nil, err
	}

	// Read the debian archive file
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := ar.NewReader(f)

	// Skip debian-binary
	_, err = reader.Next()
	if err != nil {
		return nil, err
	}

	// control.tar
	_, err = reader.Next()
	if err != nil {
		return nil, err
	}
	var bufControl bytes.Buffer
	io.Copy(&bufControl, reader)

	pkgInfo, err := dpkg.ParseControl(db, bufControl)
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		doc:           pkgInfo.Paragraph,
		local:         true,
		localFilepath: archivePath,
	}
	cache.AddPackage(pkg)

	return pkg, nil
}
