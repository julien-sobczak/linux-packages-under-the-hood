package apt

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/julien-sobczak/deb822"
	"github.com/julien-sobczak/linux-packages-from-scratch/internal/dpkg"
)

type CacheFile struct {
	cache    *pkgCache
	depCache *pkgDepCache
	sources  []*pkgSource
}

type pkgCache struct {
	packages map[string]*Package
}

type pkgDepCache struct {
	cache  *pkgCache
	states map[string]*StateCache
	order  []string
}

type pkgSource struct {
	doc     deb822.Paragraph // Release file content
	indices []*pkgIndexFile  // List of Packages files

	// parsed from the sources.list
	Type string
	URI  string
	Dist string

	// parsed from Index file
	Codename string
	Suite    string
	Origin   string
	Label    string
	Entries  map[string]string
}

func (s *pkgSource) EscapedURI() string {
	return strings.ReplaceAll(strings.TrimPrefix(s.URI, "http://"), "/", "_")
}

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
		// Basic parser (ignore options or unused attributes)
		source := &pkgSource{
			Type: parts[0],
			URI:  parts[1],
			Dist: parts[2],
		}
		results = append(results, source)
	}

	return results
}

func (c *CacheFile) BuildCaches() {
	c.cache = &pkgCache{
		packages: make(map[string]*Package),
	}
}

func (c *CacheFile) BuildDepCache() {
	states := make(map[string]*StateCache)

	// Read /var/lib/dpkg/status

	status, err := ParseStatus()
	if err != nil {
		fmt.Printf("E: The package lists or status file could not be parsed or opened.")
		os.Exit(1)
	}

	// Add state for packages found in source lists
	for _, pkg := range c.GetPackages() {
		state := StateCache{
			CandidateVersion: pkg.Version(),
		}
		states[pkg.Name()] = &state
	}

	// Add state for package already installed
	for _, pkg := range status.Paragraphs {
		state, ok := states[pkg.Value("Package")]
		if !ok {
			state = &StateCache{}
			states[pkg.Value("Package")] = state
		}
		state.CurrentVersion = pkg.Value("Version")
	}

	c.depCache = &pkgDepCache{
		cache:  c.cache,
		states: states,
	}
}

func ParseStatus() (*deb822.Document, error) {
	statusPath := filepath.Join(dpkg.VarDir, "status")
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
	return &statusContent, nil
}

func (c *CacheFile) BuildSourceList() {
	var sources []*pkgSource

	// Read /etc/apt/sources.list
	mainPath := filepath.Join(EtcDir, "sources.list")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(mainPath)
		if err != nil {
			fmt.Printf("E: Unable to read source file %s\n\t%s\n", mainPath, err)
			os.Exit(1)
		}
		sources = append(sources, ParseSourceFile(string(content))...)
	}

	// Read /etc/apt/sources.list.d/
	dirPath := filepath.Join(EtcDir, "sources.list.d")
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			fmt.Printf("E: Unable to read source dir %s\n\t%s\n", dirPath, err)
			os.Exit(1)
		}
		for _, file := range files {
			filePath := filepath.Join(dirPath, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("E: Unable to read source file %s\n\t%s\n", filePath, err)
				os.Exit(1)
			}
			sources = append(sources, ParseSourceFile(string(content))...)
		}
	}
	c.sources = sources
}

func (c *CacheFile) Open() {
	acq := NewPkgAcquire(c)

	if c.sources == nil {
		c.BuildCaches()
		c.BuildSourceList()
		c.BuildDepCache()
	}

	for _, source := range c.sources {
		if source.Type == "deb-src" {
			continue // We are interested only in binary packages
		}
		acq.Add(NewMetaIndexItem(source))
	}

	err := acq.Run()
	if err != nil {
		fmt.Printf("E: Unable to fetch resources\n\t%s\n", err)
		os.Exit(1)
	}
}

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

func (c *CacheFile) MarkForInstallation(pkgName string) {
	// We will write a very basic version. We ignore most issues like versioning.
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

	// Make sure to mark the package to prevent infinite cycles
	state.CandidateVersion = pkg.Version()
	state.flagInstall = true

	// Mark dependencies recursively
	for _, dep := range pkg.Depends() {
		c.MarkForInstallation(dep.Name)
	}

	// Add dependencies first in the installation sequence order
	c.depCache.order = append(c.depCache.order, pkgName)
}

func (c *CacheFile) GetState(pkg *Package) *StateCache {
	var state *StateCache
	state, ok := c.depCache.states[pkg.Name()]
	if !ok {
		state = &StateCache{
			CandidateVersion: pkg.Version(),
			flagInstall:      false,
		}
		c.depCache.states[pkg.Name()] = state
	}
	return state
}

func (c *CacheFile) InstCount() int {
	count := 0
	for _, state := range c.depCache.states {
		if state.Install() {
			count++
		}
	}
	return count
}

type StateCache struct {
	CandidateVersion string
	CurrentVersion   string
	flagInstall      bool
}

func (s *StateCache) Upgradable() bool {
	return s.CurrentVersion != "" && s.CandidateVersion != "" && s.CurrentVersion != s.CandidateVersion
}

func (s *StateCache) Install() bool {
	return s.flagInstall
}

func (s *StateCache) Installed() bool {
	return s.CurrentVersion != ""
}

func (s StateCache) String() string {
	res := ""
	if s.Installed() {
		res += fmt.Sprintf("installed (%s)", s.CurrentVersion)
	}
	if s.Upgradable() {
		res += fmt.Sprintf(" upgradable (%s)", s.CandidateVersion)
	}
	if s.Install() {
		res += fmt.Sprintf("mark for installation (%s)", s.CandidateVersion)
	}
	return res
}
