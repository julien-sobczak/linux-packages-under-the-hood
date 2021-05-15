package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/julien-sobczak/deb822"
	"github.com/ulikunitz/xz"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

// The command `apt install` requires more code
// than our previous implementation of the command `dpkg`.
// We will introduce the different component successively.

///////////////////////////////////////////////////////////

//
// The Acquire subsystem
//

// APT accepts package names and need to retrieve their archives
// from repositories, commonly using HTTP.
// The pkgAcquire struct download the various required files
// using a pool of worker to process each item to download.
// Like the real implementation, this system is not a generic downloader
// but contains some APT logic.

type pkgAcquire struct {
	// The downloaded items are used to populate the APT cache
	cacheFile *CacheFile

	// The items still not finished.
	pendingJobs int
	jobs        chan Item
	results     chan error
	// Workers are run in goroutines and push new items.
	jobsMutex sync.Mutex
}

// There are different types of files to retrieve from an APT repository:
// - `InRelease`: the metadata about the repository.
// - `Packages`: the list of packages present in the repository.
// - `.deb` files: the archives to install using `dpkg`.
// Each item is accessible from an URI, must be stored locally, and requires
// some postprocessing like checking the integrity of the files to prevent
// MITM attacks.

type Item interface {
	// DownloadURI returns the URI to retrieve the item.
	DownloadURI() string

	// DestFile returns the path where the file
	// represented by the URI must be written.
	DestFile(uri string) string

	// Done is called when the file has been downloaded.
	// This function updates the cache with the retrieved item
	// and can trigger new downloads.
	Done(c *CacheFile, a *pkgAcquire) error
}

// We will detail each type after the implementation of pkgAcquire.

// NewPkgAcquire initializes the Acquire system.
func NewPkgAcquire(c *CacheFile) *pkgAcquire {
	a := &pkgAcquire{
		cacheFile:   c,
		pendingJobs: 0,
		jobs:        make(chan Item, 1000),
		results:     make(chan error, 1000),
	}

	// Start the workers responsible to process the items in `jobs`.
	for w := 1; w <= 2; w++ {
		go a.worker(w, a.jobs, a.results)
	}

	return a
}

// Add is used to request the downloading of a new item.
// New items are simply send to the `jobs` channel.
func (a *pkgAcquire) Add(item Item) {
	// The function is called from different goroutines.
	// We use a lock to prevent data inconsistencies.
	a.jobsMutex.Lock()
	a.jobs <- item
	a.pendingJobs++
	a.jobsMutex.Unlock()
}

// A worker simply reads from the `jobs` channel and uses the different
// methods defined by `Item` to know what to do.

func (a *pkgAcquire) worker(id int, jobs <-chan Item, results chan<- error) {
	for item := range jobs {
		results <- a.downloadItem(item)
	}
}

func (a *pkgAcquire) downloadItem(item Item) error {
	uri := item.DownloadURI()
	dest := item.DestFile(uri)

	// Download the file
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Printf("Err: %v\n\t%s\n", item, err)
		return err
	}
	defer resp.Body.Close()

	// Create the local file
	os.MkdirAll(filepath.Dir(dest), 0755)

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the body to the local file
	io.Copy(out, resp.Body)

	fmt.Printf("Get: %v\n", item)

	return item.Done(a.cacheFile, a)
}

// There is one remaining method to cover.
// The Acquire system will try to download items in parallel
// but the code often need to block until all items have been downloaded
// to continue. The next function is used to wait.

/**
 * Run downloads all items that have been added to this
 * download process.
 *
 * This method will block until the download completes.
 */
func (a *pkgAcquire) Run() error {
	var errors []string
	var err error

	for {
		// Exit when there are no more remaining jobs
		a.jobsMutex.Lock()
		if a.pendingJobs == 0 {
			a.jobsMutex.Unlock()
			break
		}
		a.jobsMutex.Unlock()

		// Search for errors in the results
		err = <-a.results
		if err != nil {
			errors = append(errors, err.Error())
		}

		a.jobsMutex.Lock()
		a.pendingJobs--
		a.jobsMutex.Unlock()
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}

	return nil
}

// That's all for the Acquire system. What remains is the implementation
// of the various types of Item.

///////////////////////////////////////////////////////////

/*
 * The first kind of `Item` we have to download are `InRelease` files.
 * These files contain metadata about other index files (ex: `Packages`)
 * present in the same repository and are used to check the integrity
 * of these files.
 */

type MetaIndexItem struct { // InRelease/Release files
	// The Debian source pointing to this repository.
	// The source contains fields required to determine the target URI.
	source *pkgSource
}

func NewMetaIndexItem(source *pkgSource) *MetaIndexItem {
	return &MetaIndexItem{
		source: source,
	}
}

func (i *MetaIndexItem) DownloadURI() string {
	// Ex: http://deb.debian.org/debian/dists/buster/InRelease
	return i.source.URI + "/dists/" + i.source.Dist + "/InRelease"
}

func (i *MetaIndexItem) DestFile(uri string) string {
	// Ex: /var/lib/apt/lists/deb.debian.org_debian_dists_buster_InRelease
	s := i.source
	return "/var/lib/apt/lists/" +
		fmt.Sprintf("%s.%s_InRelease", s.EscapedURI(), s.Dist)
}

func (i *MetaIndexItem) Done(c *CacheFile, acq *pkgAcquire) error {
	s := i.source

	filePath := i.DestFile(s.URI)

	// 1. Check the file integrity

	// APT loads all GPG keys under /etc/apt/trusted.gpg.d/.
	// Here, for simplicity, we load only the single key we really need:
	// /etc/apt/trusted.gpg.d/debian-archive-buster-stable.gpg
	publicKey := fmt.Sprintf(
		"/etc/apt/trusted.gpg.d/debian-archive-%s-stable.gpg", s.Dist)
	decodedContent, err := gpgDecode(filePath, publicKey)
	if err != nil {
		return fmt.Errorf("the following signature couldn't be verified %s\n%v",
			filePath, err)
	}

	// 2. Parse the content to extract metadata like the checksums
	// for other files to download
	parser, err := deb822.NewParser(strings.NewReader(string(decodedContent)))
	if err != nil {
		return fmt.Errorf("malformed Release file: %v", err)
	}
	doc, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("malformed Release file: %v", err)
	}

	// Extract values
	s.doc = doc.Paragraphs[0]
	s.Codename = s.doc.Value("Codename") // Ex: buster
	s.Suite = s.doc.Value("Suite")       // Ex: stable
	s.Origin = s.doc.Value("Origin")     // Ex: Debian
	s.Label = s.doc.Value("Label")       // Ex: Debian
	s.Entries = make(map[string]string)
	for _, entry := range strings.Split(s.doc.Value("MD5Sum"), "\n") {
		// Ex: 0233ae8f041ca0f1aa5a7f395d326e80    57365 contrib/Contents-all.gz
		fields := regexp.MustCompile(`\s+`).Split(entry, -1)
		relativePath := strings.TrimSpace(fields[2])
		md5sum := fields[0]
		s.Entries[relativePath] = md5sum
	}

	// 3. Download the `Packages` files
	acq.Add(NewIndexItem(s, "main", "amd64"))
	// The real code download other Packages files in addition
	// like the ones for the `contrib` and `non-free` components.

	return nil
}
func (i MetaIndexItem) String() string {
	// Ex: https://packages.grafana.com/oss/deb stable InRelease
	return fmt.Sprintf("%s stable InRelease", i.source.URI)
}

///////////////////////////////////////////////////////////

/*
 * The second kind of Item we have to download are index files
 * (Packages and Sources files).
 * In this implementation, we are ignore Sources index files.
 * Packages index files list the Debian control files (DEBIAN/control)
 * with a few additional fields for every .deb package available.
 */

type IndexItem struct { // `Packages`/`Sources` files
	source       *pkgSource
	component    string // Ex: main, free or non-free
	architecture string // Ex: amd64

}

func NewIndexItem(source *pkgSource,
	component string, architecture string) *IndexItem {
	return &IndexItem{
		source:       source,
		component:    component,
		architecture: architecture,
	}
}

func (i *IndexItem) DownloadURI() string {
	// Ex: http://deb.debian.org/debian/dists/buster/main/binary-all/Packages.xz
	return i.source.URI + "/dists/" + i.source.Dist + "/" + i.component +
		"/binary-" + i.architecture + "/Packages.xz"
}

func (i *IndexItem) DestFile(uri string) string {
	// Ex: /var/lib/apt/lists/
	//       deb.debian.org_debian_dists_buster_main_binary-amd64_Packages.xz
	s := i.source
	return "/var/lib/apt/lists/" + fmt.Sprintf("%s.%s_%s_binary-%s_Packages.xz",
		s.EscapedURI(), s.Dist, i.component, i.architecture)
}

func (i *IndexItem) Done(c *CacheFile, a *pkgAcquire) error {
	s := i.source
	path := i.DestFile(s.URI)

	// 1. Read the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("missing file: %v", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %v", path, err)
	}

	// 2. Check integrity
	hash := md5.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		return fmt.Errorf("unable to determine MD5 sum: %s", err)
	}
	md5sum := fmt.Sprintf("%x", hash.Sum(nil))
	md5sumRef := s.Entries[i.EntryName()]
	if md5sum != md5sumRef {
		return fmt.Errorf("found MD5 mismatch: %v != %v", md5sum, md5sumRef)
	}

	// 3. Extract content
	r, err := xz.NewReader(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("unable to open xz file: %v", err)
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read index file content: %v", err)
	}

	// 4. Parse content
	parser, err := deb822.NewParser(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("malformed index file: %v", err)
	}
	doc, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("malformed index file: %v", err)
	}

	// 5. Add the package into the APT cache.
	for _, paragraph := range doc.Paragraphs {
		c.AddPackage(&Package{
			doc:    paragraph,
			source: s,
		})
	}

	return nil
}

// EntryName returns the key in MD5Sum for this file in the Release file.
func (i IndexItem) EntryName() string {
	// Ex: main/binary-am64/Packages
	return fmt.Sprintf("%s/binary-%s/Packages.xz", i.component, i.architecture)
}

func (i IndexItem) String() string {
	// Ex: https://packages.grafana.com/oss/deb stable/main amd64 Packages
	return fmt.Sprintf("%s stable/%s %s Packages",
		i.source.URI, i.component, i.architecture)
}

///////////////////////////////////////////////////////////

/*
 * The last kind of Item we have to download are .deb archives that will
 * be passed to the dpkg command to proceed to the installation.
 * These files are downloaded under /var/cache/apt/archives/.
 */

type PackageItem struct { // `.deb` files
	// The package metadata associated with the archive to download.
	pkg *Package
}

func NewPackageItem(pkg *Package) *PackageItem {
	return &PackageItem{
		pkg: pkg,
	}
}

func (i *PackageItem) DownloadURI() string {
	// Ex: http://deb.debian.org/debian/pool/main/r/rsync/rsync_3.2.3_amd64.deb
	return i.pkg.source.URI + "/" + i.pkg.doc.Value("Filename")
}

func (i *PackageItem) DestFile(uri string) string {
	// Ex: /var/cache/apt/archives/rsync_3.2.3-4_amd64.deb
	pkg := i.pkg
	pkg.cacheFilepath = "/var/cache/apt/archives/" + filepath.Base(uri)
	return pkg.cacheFilepath
}

func (i *PackageItem) Done(c *CacheFile, a *pkgAcquire) error {
	// 1. Check file integrity
	f, err := os.Open(i.pkg.cacheFilepath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	indexChecksum := i.pkg.doc.Value("SHA256")
	effectiveChecksum := fmt.Sprintf("%x", h.Sum(nil))

	if indexChecksum != effectiveChecksum {
		return fmt.Errorf("invalid checksum for %s", i.pkg.cacheFilepath)
	}

	// 2. Nothing more to do.
	// The archive will be processed later when delegating to the `dpkg` command.

	return nil
}

func (i PackageItem) String() string {
	// Ex: https://grafana.com/oss/deb stable/main amd64 grafana amd64 7.5.5
	pkg := i.pkg
	return fmt.Sprintf("%s stable/main %s %s %s", pkg.source.URI,
		pkg.Name(), pkg.Architecture(), pkg.Version())
}

///////////////////////////////////////////////////////////

//
// The APT Cache
//

// We try to using the same naming as for the real implementation
// using similar structs but containing only the main fields.

// CacheFile is the high-level component for the APT cache.
type CacheFile struct {
	cache    *pkgCache
	depCache *pkgDepCache
	sources  []*pkgSource
}

// pkgCache contains all known packages
// (found in Dpkg database and in repositories)
type pkgCache struct {
	packages map[string]*Package // The key is the package name
}

// pkgDepCache contains the state information for every package
// (installed, to install, upgradable, ...).
type pkgDepCache struct {
	states map[string]*StateCache
	// The ordered list of packages waiting to be installed.
	order []string
}

// pkgSource represents a single line in a source.list file.
type pkgSource struct {
	doc deb822.Paragraph // `Release` file content

	// parsed from the sources.list file
	Type string
	URI  string
	Dist string

	// parsed from the Packages file
	Codename string
	Suite    string
	Origin   string
	Label    string
	Entries  map[string]string // Checksums of all repository files
}

// EscapedURI returns a name based on the URI that can be used in filename.
// Indeed, most retrieved files are stored under /var/lib/apt/
// and are named after their source.
func (s *pkgSource) EscapedURI() string {
	return strings.ReplaceAll(strings.TrimPrefix(s.URI, "http://"), "/", "_")
}

// The core of the APT cache is the list of packages.

// Package is a Debian package.
type Package struct {
	// The metadata as present in `Packages` or `status` file
	doc deb822.Paragraph
	// The source where this package is coming from.
	// Can be undefined for already installed packages.
	source *pkgSource

	// The path under /var/cache/apt/packages.
	// Initialized after the download of the package.
	cacheFilepath string
}

// We expose a few additional methods to extract attributes
// from the underlying DEB822 document.

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

func ParseDependencies(values string) []Dependency {
	// Ex: "adduser, gpgv | gpgv2 | gpgv1, libapt-pkg5.0 (>= 1.7.0~alpha3~)"
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
	// "adduser", "gpgv | gpgv2", "libc6 (>= 2.15)",
	// "python3:any (>= 3.5~)", "foo [i386]", "perl:any", "perlapi-5.28.0"

	var dep Dependency

	r := regexp.MustCompile(`^(?P<name>[\w\.-]+)(?:[:]\w+)?` +
		`(?: [(](?P<relation>(?:>>|>=|=|<=|<<)) ` +
		`(?P<version>\S+)[)])?(?: [|].*)?$`)
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

// That's all for the different structures relating to the APT cache.

///////////////////////////////////////////////////////////

// Now, we need to initialize the three main components.
// The first step is thus to create the array containing all known packages.
// This array will be populated in the successive steps.

func (c *CacheFile) BuildCaches() {
	c.cache = &pkgCache{
		packages: make(map[string]*Package),
	}
}

// The second step is to read the lists of source to find the `Packages` files
// containing the list of available packages.
// So, we need a function to parse these local source files.

// ParseSourceFile parses a single source file.
// It only supports the common multi-line format,
// and not the most recent DEB822 format.
func ParseSourceFile(content string) []*pkgSource {
	var results []*pkgSource

	scanner := bufio.NewScanner(strings.NewReader(content))
	// Read line by line
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			// Ignore blank lines
			continue
		}
		if strings.HasPrefix("#", line) {
			// Ignore comments
			continue
		}
		parts := strings.Split(line, " ")
		// Basic parser (ignore some options or unused attributes)
		source := &pkgSource{
			Type: parts[0],
			URI:  parts[1],
			Dist: parts[2],
		}
		results = append(results, source)
	}

	return results
}

// BuildSourceList parses every source file.
func (c *CacheFile) BuildSourceList() {
	var sources []*pkgSource

	// Read /etc/apt/sources.list
	mainPath := "/etc/apt/sources.list"
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(mainPath)
		if err != nil {
			fmt.Printf("E: Unable to read source file\n\t%s\n", err)
			os.Exit(1)
		}
		sources = append(sources, ParseSourceFile(string(content))...)
	}

	// Read /etc/apt/sources.list.d/
	dirPath := "/etc/apt/sources.list.d/"
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			fmt.Printf("E: Unable to read source dir\n\t%s\n", err)
			os.Exit(1)
		}
		for _, file := range files {
			filePath := filepath.Join(dirPath, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("E: Unable to read source file\n\t%s\n", err)
				os.Exit(1)
			}
			sources = append(sources, ParseSourceFile(string(content))...)
		}
	}
	c.sources = sources
}

// The last step is to read the Dpkg database
// to determine the packages already installed.
// Therefore, we need a function to parse the status file.

func ParseStatus() (*deb822.Document, error) {
	f, err := os.Open("/var/lib/dpkg/status")
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
	return &statusContent, nil
}

func (c *CacheFile) BuildDepCache() {
	states := make(map[string]*StateCache)

	// Read /var/lib/dpkg/status
	status, err := ParseStatus()
	if err != nil {
		fmt.Printf("E: The package lists or status file could not be parsed.")
		os.Exit(1)
	}

	// Add state for packages already installed
	for _, pkg := range status.Paragraphs {
		// The status file also contains packages
		// that was partially installed or removed.
		if !strings.Contains(pkg.Value("Status"), "installed") {
			continue
		}
		state, ok := states[pkg.Value("Package")]
		if !ok {
			state = &StateCache{}
			states[pkg.Value("Package")] = state
		}
		state.CurrentVersion = pkg.Value("Version")
	}

	c.depCache = &pkgDepCache{
		states: states,
	}
}

///////////////////////////////////////////////////////////

// We now have the three functions required to initialize the APT cache.
// We will hide them behind a simple method.

func (c *CacheFile) Open() {
	// Initialize the Acquire system to download file from repositories
	acq := NewPkgAcquire(c)

	// Initialize the cache structure
	if c.sources == nil {
		c.BuildCaches()
		c.BuildSourceList()
		c.BuildDepCache()
	}

	// Download items from repositories
	for _, source := range c.sources {
		if source.Type == "deb-src" {
			continue // We are interested only in binary packages
		}
		acq.Add(NewMetaIndexItem(source))
	}

	// Wait for all items to be downloaded to return
	err := acq.Run()
	if err != nil {
		fmt.Printf("E: Unable to fetch resources\n\t%s\n", err)
		os.Exit(1)
	}
}

// As we have glimpsed before, the cache content is populated
// from the `Done()` methods of the different types of `Item`.
// We need to expose additional methods to easily add or retrieve
// these packages and their state.

func (c *CacheFile) AddPackage(p *Package) {
	c.cache.packages[p.Name()] = p
}

func (c *CacheFile) GetPackage(name string) *Package {
	if p, ok := c.cache.packages[name]; ok {
		return p
	}
	return nil
}

func (c *CacheFile) GetPackages() []*Package {
	values := make([]*Package, 0, len(c.cache.packages))
	for _, v := range c.cache.packages {
		values = append(values, v)
	}
	return values
}

func (c *CacheFile) GetState(pkg *Package) *StateCache {
	var state *StateCache
	state, ok := c.depCache.states[pkg.Name()]
	if !ok {
		// Only the state of installed packages is present.
		// We defer the initializion for other packages until
		// the first access.
		state = &StateCache{
			CandidateVersion: pkg.Version(),
			flagInstall:      false,
		}
		c.depCache.states[pkg.Name()] = state
	}
	return state
}

///////////////////////////////////////////////////////////

// We are almost done with the APT cache.
// We have discussed several times about the state we keep about each package
// without explaining what it means.

type StateCache struct {
	// The version that can be installed determined using sources.
	CandidateVersion string
	// The version currently installed determined using the Dpkg database.
	CurrentVersion string
	// A flag to determine if the package is marked for installation.
	flagInstall bool
}

func (s *StateCache) Upgradable() bool {
	return s.CurrentVersion != "" &&
		s.CandidateVersion != "" && s.CurrentVersion != s.CandidateVersion
}

func (s *StateCache) Install() bool {
	return s.flagInstall
}

func (s *StateCache) Installed() bool {
	return s.CurrentVersion != ""
}

// When installing a package, we must make sure its dependencies
// are already installed or we need to install them first.
// The logic is rather complicated as many things can go wrong
// with dependency management like conflicts between two packages.
// For this article, we will use a very basic approach.
// We ignore versions completely and install each missing dependencies
// without checking if it brokes other packages. This is another
// reason why you must not run this code on your host directly :).

func (c *CacheFile) MarkForInstallation(pkgName string) {
	pkg := c.GetPackage(pkgName)
	if pkg == nil {
		fmt.Printf("E: Unable to locate package %s\n", pkgName)
		os.Exit(1)
	}

	state := c.GetState(pkg)
	if state.Installed() || state.Install() {
		// Already installed or marked for installation
		return
	}

	// Make sure to mark the package before checking its dependencies
	// to prevent infinite cycles
	state.CandidateVersion = pkg.Version()
	state.flagInstall = true

	// Mark dependencies recursively
	for _, dep := range pkg.Depends() {
		c.MarkForInstallation(dep.Name)
	}

	// Add dependencies first in the installation sequence order
	c.depCache.order = append(c.depCache.order, pkgName)
}

// We end this section with an utility method to report the total
// number of packages that will be installed.
// This number differs commonly as packages have dependencies
// that must be installed and we will use this method to notify
// the user that more packages will be installed as the ones
// passed in argument.

func (c *CacheFile) InstCount() int {
	count := 0
	for _, state := range c.depCache.states {
		if state.Install() {
			count++
		}
	}
	return count
}

///////////////////////////////////////////////////////////

//
// Main
//

// We have everything we need to implement the command `apt install`.
// We will integrate everything we have covered so far.

func main() {
	var pkgNames []string
	// The command `apt install` can be called without any package to install.
	if len(os.Args) > 1 {
		pkgNames = append(pkgNames, os.Args[1:]...)
	}

	// Load the Cache
	cache := &CacheFile{}
	cache.Open()

	// Search for the packages to install
	pkgs := make(map[string]*Package)
	for _, pkgName := range pkgNames {
		// The command `apt install` also supports `.deb` file.
		// We ignore this for simplicity to avoid
		// duplicating code from the previous parts of this blog post.
		// Check https://github.com/julien-sobczak/linux-packages-under-the-hood
		// for a more complete implementation.

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
		fmt.Printf(
			"The following additional packages will be installed:\n\t%s\n",
			strings.Join(extras, " "))
	}

	// Print out the list of suggested packages
	var suggests []string
	for _, pkg := range cache.GetPackages() {
		state := cache.GetState(pkg)

		// Just look at the ones we want to install
		if !state.Install() {
			continue
		}

		// Get the suggestions for the candidate version
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

func InstallPackages(cache *CacheFile) error {
	acq := NewPkgAcquire(cache)

	// 1. Download package archives
	for _, pkgName := range cache.depCache.order {
		pkg := cache.GetPackage(pkgName)
		acq.Add(NewPackageItem(pkg))
	}
	err := acq.Run()
	if err != nil {
		return err
	}

	// 2. Run the command `dpkg -i` to install them
	var archives []string
	for _, pkgName := range cache.depCache.order {
		pkg := cache.GetPackage(pkgName)
		archives = append(archives, pkg.cacheFilepath)
	}

	// We delegate to the dpkg command to avoid repeating the previous code
	// but the complete code source of this repository reuse the same code.
	// Check https://github.com/julien-sobczak/linux-packages-under-the-hood
	out, err := exec.Command("dpkg", "-i", strings.Join(archives, " ")).Output()
	if err != nil {
		return err
	}
	fmt.Print(string(out))

	return nil
}

///////////////////////////////////////////////////////////

// Helpers

// gpgDecode checks the GPP signature of a clearsigned document and
// returns the content.
func gpgDecode(filename string, publicKey string) ([]byte, error) {
	// Open gpg clearsigned document
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening public key: %s", err)
	}
	defer r.Close()

	// Read the content
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Decode the content
	b, _ := clearsign.Decode(data)
	if b == nil {
		return nil, fmt.Errorf("not PGP signed")
	}

	// Open the public key to validate the signature
	rk, err := os.Open(publicKey)
	if err != nil {
		return nil, fmt.Errorf("error opening public key: %s", err)
	}
	defer r.Close()
	keyring, err := openpgp.ReadKeyRing(rk) // binary
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Check the signature using the public key
	_, err = openpgp.CheckDetachedSignature(keyring,
		bytes.NewBuffer(b.Bytes), b.ArmoredSignature.Body)
	if err != nil {
		return nil, err
	}

	return b.Plaintext, nil
}
