package apt

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/julien-sobczak/deb822"
	"github.com/ulikunitz/xz"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

type pkgAcquire struct {
	cacheFile *CacheFile

	pendingJobs int
	jobs        chan Item
	results     chan error
	jobsMutex   sync.Mutex

	// An increase number to identify each item uniquely in output messages
	hit      int
	hitMutex sync.Mutex
}

type Item interface {
	// URI to retrieve the item.
	DownloadURI() string

	// DestFile returns the path where the file
	// represented by the URI must be written.
	DestFile(uri string) string

	// Done is called when the file has been downloaded.
	// This function updates the cache with the retrieved item
	// and can trigger new downloads.
	Done(c *CacheFile, a *pkgAcquire) error
}

func NewPkgAcquire(c *CacheFile) *pkgAcquire {
	a := &pkgAcquire{
		cacheFile:   c,
		hit:         0,
		pendingJobs: 0,
		jobs:        make(chan Item, 1000),
		results:     make(chan error, 1000),
	}

	for w := 1; w <= 2; w++ {
		go a.worker(w, a.jobs, a.results)
	}

	return a
}

func (a *pkgAcquire) Add(item Item) {
	a.jobsMutex.Lock()
	a.jobs <- item
	a.pendingJobs++
	a.jobsMutex.Unlock()
}

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
		a.jobsMutex.Lock()
		if a.pendingJobs == 0 {
			a.jobsMutex.Unlock()
			break
		}
		a.jobsMutex.Unlock()

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

func (a *pkgAcquire) worker(id int, jobs <-chan Item, results chan<- error) {
	for item := range jobs {
		results <- a.downloadItem(item)
	}
}

func (a *pkgAcquire) downloadItem(item Item) error {
	uri := item.DownloadURI()

	dest := item.DestFile(uri)
	resp, err := http.Get(uri)

	a.hitMutex.Lock()
	a.hit++
	hit := a.hit
	a.hitMutex.Unlock()

	if err != nil {
		fmt.Printf("Err%d: %v\n\t%s\n", hit, item, err)
		return err
	}
	defer resp.Body.Close()

	// Create the file
	err = os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Get:%d %v [%s]\n", hit, item, humanReadable(fileSize(dest)))

	return item.Done(a.cacheFile, a)
}

/*
 * The first kind of Item we have to download are Release files.
 * These files contain meta-information about other index files (ex: Packages) present in a repository
 * and are used to check the integrity of these files.
 */

type MetaIndexItem struct { // Release
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

	return filepath.Join(VarDir, "lists", fmt.Sprintf("%s.%s_InRelease", s.EscapedURI(), s.Dist))
}

func (i *MetaIndexItem) Done(c *CacheFile, acq *pkgAcquire) error {
	s := i.source

	filePath := i.DestFile(s.URI)

	// APT loads all GPG keys under /etc/apt/trusted.gpg.d/.
	// Here, for simplicity, we load only the single key we really need:
	// /etc/apt/trusted.gpg.d/debian-archive-buster-stable.gpg
	publicKey := fmt.Sprintf("%s/trusted.gpg.d/debian-archive-%s-stable.gpg", EtcDir, s.Dist)
	decodedContent, err := gpgDecode(filePath, publicKey)
	if err != nil {
		return fmt.Errorf("the following signatures couldn't be verified because the public key is not available: %s\n%v", filePath, err)
	}

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

	// Download the packages files
	acq.Add(NewIndexItem(s, "main", "amd64"))

	return nil
}
func (i MetaIndexItem) String() string {
	// Ex: https://packages.grafana.com/oss/deb stable InRelease
	return fmt.Sprintf("%s stable InRelease", i.source.URI)
}

/*
 * The second kind of Item we have to download are index files (Packages and Sources files).
 * In this implementation, we are ignore Sources index files.
 * Packages index files list the Debian control files (DEBIAN/control) with a few additional fields
 * for every .deb package available.
 */

type IndexItem struct { // Packages/Sources
	source       *pkgSource
	component    string // Ex: main, free or non-free
	architecture string // Ex: amd64

}

func NewIndexItem(source *pkgSource, component string, architecture string) *IndexItem {
	return &IndexItem{
		source:       source,
		component:    component,
		architecture: architecture,
	}
}

func (i *IndexItem) DownloadURI() string {
	// Ex: http://deb.debian.org/debian/dists/buster/main/binary-amd64/Packages.xz
	return i.source.URI + "/dists/" + i.source.Dist + "/" + i.component + "/binary-" + i.architecture + "/Packages.xz"
}

func (i *IndexItem) DestFile(uri string) string {
	// Ex: /var/lib/apt/lists/deb.debian.org_debian_dists_buster_main_binary-amd64_Packages.xz
	s := i.source
	return filepath.Join(VarDir, "lists", fmt.Sprintf("%s.%s_%s_binary-%s_Packages.xz", s.EscapedURI(), s.Dist, i.component, i.architecture))
}

func (i *IndexItem) Done(c *CacheFile, a *pkgAcquire) error {
	s := i.source
	path := i.DestFile(s.URI)

	// Read the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("missing file: %v", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %v", path, err)
	}

	// Check integrity
	hash := md5.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		return fmt.Errorf("unable to determine MD5 sum: %s", err)
	}
	md5sum := fmt.Sprintf("%x", hash.Sum(nil))
	md5sumReference := s.Entries[i.EntryName()]
	if md5sum != md5sumReference {
		return fmt.Errorf("found MD5 mismatch: %v != %v", md5sum, md5sumReference)
	}

	// Extract content
	r, err := xz.NewReader(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("unable to open xz file: %v", err)
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read index file content: %v", err)
	}

	// Parse content
	parser, err := deb822.NewParser(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("malformed index file: %v", err)
	}
	doc, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("malformed index file: %v", err)
	}

	// Process content
	s.indices = append(s.indices, &pkgIndexFile{
		doc: doc,
	})

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
	return fmt.Sprintf("%s stable/%s %s Packages", i.source.URI, i.component, i.architecture)
}

/*
 * The last kind of Item we have to download are .deb archives that will be passed
 * to the dpkg command to proceed to the installation.
 * These files are downloaded under /var/cache/apt/archives/.
 */

type PackageItem struct { // .deb
	pkg *Package
}

func NewPackageItem(pkg *Package) *PackageItem {
	return &PackageItem{
		pkg: pkg,
	}
}

func (i *PackageItem) DownloadURI() string {
	// Ex: http://deb.debian.org/debian/pool/main/r/rsync/rsync_3.2.3-4_amd64.deb
	return i.pkg.source.URI + "/" + i.pkg.doc.Value("Filename")
}

func (i *PackageItem) DestFile(uri string) string {
	// Ex: /var/cache/apt/archives/rsync_3.2.3-4_amd64.deb
	pkg := i.pkg
	pkg.cacheFilepath = filepath.Join(CacheDir, "archives", filepath.Base(uri))
	return pkg.cacheFilepath
}

func (i *PackageItem) Done(c *CacheFile, a *pkgAcquire) error {
	// Check file integrity

	// Calculate the checksum
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

	return nil
}

func (i PackageItem) String() string {
	// Ex: https://packages.grafana.com/oss/deb stable/main amd64 grafana amd64 7.5.5
	pkg := i.pkg
	return fmt.Sprintf("%s stable/main %s %s %s", pkg.source.URI, pkg.Name(), pkg.Architecture(), pkg.Version())
}

// Helpers

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
	_, err = openpgp.CheckDetachedSignature(keyring, bytes.NewBuffer(b.Bytes), b.ArmoredSignature.Body)
	if err != nil {
		return nil, err
	}

	for k, v := range b.ArmoredSignature.Header {
		log.Printf("%s => %s", k, v)
	}
	// TODO FIXME
	//  W: GPG error: http://ppa.launchpad.net precise
	//  Release: The following signatures couldn't be verified because the public key is not available:
	//  NO_PUBKEY 2EA8F35793D8809A

	return b.Plaintext, nil
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func humanReadable(b int64) string {
	// From https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
