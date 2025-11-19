package bundler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"git.wntrmute.dev/kyle/goutils/certlib"
)

const defaultFileMode = 0644

// Config represents the top-level YAML configuration.
type Config struct {
	Config struct {
		Hashes string `yaml:"hashes"`
		Expiry string `yaml:"expiry"`
	} `yaml:"config"`
	Chains map[string]ChainGroup `yaml:"chains"`
}

// ChainGroup represents a named group of certificate chains.
type ChainGroup struct {
	Certs   []CertChain `yaml:"certs"`
	Outputs Outputs     `yaml:"outputs"`
}

// CertChain represents a root certificate and its intermediates.
type CertChain struct {
	Root          string   `yaml:"root"`
	Intermediates []string `yaml:"intermediates"`
}

// Outputs defines output format options.
type Outputs struct {
	IncludeSingle     bool     `yaml:"include_single"`
	IncludeIndividual bool     `yaml:"include_individual"`
	Manifest          bool     `yaml:"manifest"`
	Formats           []string `yaml:"formats"`
	Encoding          string   `yaml:"encoding"`
}

var formatExtensions = map[string]string{
	"zip": ".zip",
	"tgz": ".tar.gz",
}

// Run performs the bundling operation given a config file path and an output directory.
func Run(configFile string, outputDir string) error {
	if configFile == "" {
		return errors.New("configuration file required")
	}

	cfg, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	expiryDuration := 365 * 24 * time.Hour
	if cfg.Config.Expiry != "" {
		expiryDuration, err = parseDuration(cfg.Config.Expiry)
		if err != nil {
			return fmt.Errorf("parsing expiry: %w", err)
		}
	}

	if err = os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	totalFormats := 0
	for _, group := range cfg.Chains {
		totalFormats += len(group.Outputs.Formats)
	}
	createdFiles := make([]string, 0, totalFormats)
	for groupName, group := range cfg.Chains {
		files, perr := processChainGroup(groupName, group, expiryDuration, outputDir)
		if perr != nil {
			return fmt.Errorf("processing chain group %s: %w", groupName, perr)
		}
		createdFiles = append(createdFiles, files...)
	}

	if cfg.Config.Hashes != "" {
		hashFile := filepath.Join(outputDir, cfg.Config.Hashes)
		if gerr := generateHashFile(hashFile, createdFiles); gerr != nil {
			return fmt.Errorf("generating hash file: %w", gerr)
		}
	}

	return nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if uerr := yaml.Unmarshal(data, &cfg); uerr != nil {
		return nil, uerr
	}

	return &cfg, nil
}

func parseDuration(s string) (time.Duration, error) {
	// Support simple formats like "1y", "6m", "30d"
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	unit := s[len(s)-1]
	value := s[:len(s)-1]

	var multiplier time.Duration
	switch unit {
	case 'y', 'Y':
		multiplier = 365 * 24 * time.Hour
	case 'm', 'M':
		multiplier = 30 * 24 * time.Hour
	case 'd', 'D':
		multiplier = 24 * time.Hour
	default:
		return time.ParseDuration(s)
	}

	var num int
	_, err := fmt.Sscanf(value, "%d", &num)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", s)
	}

	return time.Duration(num) * multiplier, nil
}

func processChainGroup(
	groupName string,
	group ChainGroup,
	expiryDuration time.Duration,
	outputDir string,
) ([]string, error) {
	// Default encoding to "pem" if not specified
	encoding := group.Outputs.Encoding
	if encoding == "" {
		encoding = "pem"
	}

	// Collect certificates from all chains in the group
	singleFileCerts, individualCerts, sourcePaths, err := loadAndCollectCerts(
		group.Certs,
		group.Outputs,
		expiryDuration,
	)
	if err != nil {
		return nil, err
	}

	// Prepare files for inclusion in archives
	archiveFiles, err := prepareArchiveFiles(singleFileCerts, individualCerts, sourcePaths, group.Outputs, encoding)
	if err != nil {
		return nil, err
	}

	// Create archives for the entire group
	createdFiles, err := createArchiveFiles(groupName, group.Outputs.Formats, archiveFiles, outputDir)
	if err != nil {
		return nil, err
	}

	return createdFiles, nil
}

// loadAndCollectCerts loads all certificates from chains and collects them for processing.
func loadAndCollectCerts(
	chains []CertChain,
	outputs Outputs,
	expiryDuration time.Duration,
) ([]*x509.Certificate, []certWithPath, []string, error) {
	var singleFileCerts []*x509.Certificate
	var individualCerts []certWithPath
	var sourcePaths []string

	for _, chain := range chains {
		s, i, cerr := collectFromChain(chain, outputs, expiryDuration)
		if cerr != nil {
			return nil, nil, nil, cerr
		}
		if len(s) > 0 {
			singleFileCerts = append(singleFileCerts, s...)
		}
		if len(i) > 0 {
			individualCerts = append(individualCerts, i...)
		}
		// Record source paths for timestamp preservation
		// Only append when loading succeeded
		sourcePaths = append(sourcePaths, chain.Root)
		sourcePaths = append(sourcePaths, chain.Intermediates...)
	}

	return singleFileCerts, individualCerts, sourcePaths, nil
}

// collectFromChain loads a single chain, performs checks, and returns the certs to include.
func collectFromChain(
	chain CertChain,
	outputs Outputs,
	expiryDuration time.Duration,
) (
	[]*x509.Certificate,
	[]certWithPath,
	error,
) {
	var single []*x509.Certificate
	var indiv []certWithPath

	// Load root certificate
	rootCert, rerr := certlib.LoadCertificate(chain.Root)
	if rerr != nil {
		return nil, nil, fmt.Errorf("failed to load root certificate %s: %w", chain.Root, rerr)
	}

	// Check expiry for root
	checkExpiry(chain.Root, rootCert, expiryDuration)

	// Add root to collections if needed
	if outputs.IncludeSingle {
		single = append(single, rootCert)
	}
	if outputs.IncludeIndividual {
		indiv = append(indiv, certWithPath{cert: rootCert, path: chain.Root})
	}

	// Load and validate intermediates
	for _, intPath := range chain.Intermediates {
		intCert, lerr := certlib.LoadCertificate(intPath)
		if lerr != nil {
			return nil, nil, fmt.Errorf("failed to load intermediate certificate %s: %w", intPath, lerr)
		}

		// Validate that intermediate is signed by root
		if sigErr := intCert.CheckSignatureFrom(rootCert); sigErr != nil {
			return nil, nil, fmt.Errorf(
				"intermediate %s is not properly signed by root %s: %w",
				intPath,
				chain.Root,
				sigErr,
			)
		}

		// Check expiry for intermediate
		checkExpiry(intPath, intCert, expiryDuration)

		// Add intermediate to collections if needed
		if outputs.IncludeSingle {
			single = append(single, intCert)
		}
		if outputs.IncludeIndividual {
			indiv = append(indiv, certWithPath{cert: intCert, path: intPath})
		}
	}

	return single, indiv, nil
}

// prepareArchiveFiles prepares all files to be included in archives.
func prepareArchiveFiles(
	singleFileCerts []*x509.Certificate,
	individualCerts []certWithPath,
	sourcePaths []string,
	outputs Outputs,
	encoding string,
) ([]fileEntry, error) {
	var archiveFiles []fileEntry

	// Track used filenames to avoid collisions inside archives
	usedNames := make(map[string]int)

	// Handle a single bundle file
	if outputs.IncludeSingle && len(singleFileCerts) > 0 {
		bundleTime := maxModTime(sourcePaths)
		files, err := encodeCertsToFiles(singleFileCerts, "bundle", encoding, true)
		if err != nil {
			return nil, fmt.Errorf("failed to encode single bundle: %w", err)
		}
		for i := range files {
			files[i].name = makeUniqueName(files[i].name, usedNames)
			files[i].modTime = bundleTime
			// Best-effort: we do not have a portable birth/creation time.
			// Use the same timestamp for created time to track deterministically.
			files[i].createTime = bundleTime
		}
		archiveFiles = append(archiveFiles, files...)
	}

	// Handle individual files
	if outputs.IncludeIndividual {
		for _, cp := range individualCerts {
			baseName := strings.TrimSuffix(filepath.Base(cp.path), filepath.Ext(cp.path))
			files, err := encodeCertsToFiles([]*x509.Certificate{cp.cert}, baseName, encoding, false)
			if err != nil {
				return nil, fmt.Errorf("failed to encode individual cert %s: %w", cp.path, err)
			}
			mt := fileModTime(cp.path)
			for i := range files {
				files[i].name = makeUniqueName(files[i].name, usedNames)
				files[i].modTime = mt
				files[i].createTime = mt
			}
			archiveFiles = append(archiveFiles, files...)
		}
	}

	// Generate manifest if requested
	if outputs.Manifest {
		manifestContent := generateManifest(archiveFiles)
		manifestName := makeUniqueName("MANIFEST", usedNames)
		mt := maxModTime(sourcePaths)
		archiveFiles = append(archiveFiles, fileEntry{
			name:       manifestName,
			content:    manifestContent,
			modTime:    mt,
			createTime: mt,
		})
	}

	return archiveFiles, nil
}

// createArchiveFiles creates archive files in the specified formats.
func createArchiveFiles(
	groupName string,
	formats []string,
	archiveFiles []fileEntry,
	outputDir string,
) ([]string, error) {
	createdFiles := make([]string, 0, len(formats))

	for _, format := range formats {
		ext, ok := formatExtensions[format]
		if !ok {
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
		archivePath := filepath.Join(outputDir, groupName+ext)
		switch format {
		case "zip":
			if err := createZipArchive(archivePath, archiveFiles); err != nil {
				return nil, fmt.Errorf("failed to create zip archive: %w", err)
			}
		case "tgz":
			if err := createTarGzArchive(archivePath, archiveFiles); err != nil {
				return nil, fmt.Errorf("failed to create tar.gz archive: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
		createdFiles = append(createdFiles, archivePath)
	}

	return createdFiles, nil
}

func checkExpiry(path string, cert *x509.Certificate, expiryDuration time.Duration) {
	now := time.Now()
	expiryThreshold := now.Add(expiryDuration)

	if cert.NotAfter.Before(expiryThreshold) {
		daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)
		if daysUntilExpiry < 0 {
			fmt.Fprintf(
				os.Stderr,
				"WARNING: Certificate %s has EXPIRED (expired %d days ago)\n",
				path,
				-daysUntilExpiry,
			)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Certificate %s will expire in %d days (on %s)\n", path, daysUntilExpiry, cert.NotAfter.Format("2006-01-02"))
		}
	}
}

type fileEntry struct {
	name       string
	content    []byte
	modTime    time.Time
	createTime time.Time
}

type certWithPath struct {
	cert *x509.Certificate
	path string
}

// encodeCertsToFiles converts certificates to file entries based on encoding type
// If isSingle is true, certs are concatenated into a single file; otherwise one cert per file.
func encodeCertsToFiles(
	certs []*x509.Certificate,
	baseName string,
	encoding string,
	isSingle bool,
) ([]fileEntry, error) {
	var files []fileEntry

	switch encoding {
	case "pem":
		pemContent := encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".pem",
			content: pemContent,
		})
	case "crt":
		pemContent := encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".crt",
			content: pemContent,
		})
	case "pemcrt":
		pemContent := encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".pem",
			content: pemContent,
		})

		pemContent = encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".crt",
			content: pemContent,
		})
	case "der":
		if isSingle {
			// For single file in DER, concatenate all cert DER bytes
			var derContent []byte
			for _, cert := range certs {
				derContent = append(derContent, cert.Raw...)
			}
			files = append(files, fileEntry{
				name:    baseName + ".crt",
				content: derContent,
			})
		} else if len(certs) > 0 {
			// Individual DER file (should only have one cert)
			files = append(files, fileEntry{
				name:    baseName + ".crt",
				content: certs[0].Raw,
			})
		}
	case "both":
		// Add PEM version
		pemContent := encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".pem",
			content: pemContent,
		})
		// Add DER version
		if isSingle {
			var derContent []byte
			for _, cert := range certs {
				derContent = append(derContent, cert.Raw...)
			}
			files = append(files, fileEntry{
				name:    baseName + ".crt",
				content: derContent,
			})
		} else if len(certs) > 0 {
			files = append(files, fileEntry{
				name:    baseName + ".crt",
				content: certs[0].Raw,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported encoding: %s (must be 'pem', 'der', or 'both')", encoding)
	}

	return files, nil
}

// encodeCertsToPEM encodes certificates to PEM format.
func encodeCertsToPEM(certs []*x509.Certificate) []byte {
	var pemContent []byte
	for _, cert := range certs {
		pemBlock := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		pemContent = append(pemContent, pem.EncodeToMemory(pemBlock)...)
	}
	return pemContent
}

func generateManifest(files []fileEntry) []byte {
	// Build a sorted list of files by filename to ensure deterministic manifest ordering
	sorted := make([]fileEntry, 0, len(files))
	for _, f := range files {
		// Defensive: skip any existing manifest entry
		if f.name == "MANIFEST" {
			continue
		}
		sorted = append(sorted, f)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].name < sorted[j].name })

	var manifest strings.Builder
	for _, file := range sorted {
		hash := sha256.Sum256(file.content)
		manifest.WriteString(fmt.Sprintf("%x  %s\n", hash, file.name))
	}
	return []byte(manifest.String())
}

// closeWithErr attempts to close all provided closers, joining any close errors with baseErr.
func closeWithErr(baseErr error, closers ...io.Closer) error {
	for _, c := range closers {
		if c == nil {
			continue
		}
		if cerr := c.Close(); cerr != nil {
			baseErr = errors.Join(baseErr, cerr)
		}
	}
	return baseErr
}

func createZipArchive(path string, files []fileEntry) error {
	f, zerr := os.Create(path)
	if zerr != nil {
		return zerr
	}

	w := zip.NewWriter(f)

	for _, file := range files {
		hdr := &zip.FileHeader{
			Name:   file.name,
			Method: zip.Deflate,
		}
		if !file.modTime.IsZero() {
			hdr.SetModTime(file.modTime)
		}
		fw, werr := w.CreateHeader(hdr)
		if werr != nil {
			return closeWithErr(werr, w, f)
		}
		if _, werr = fw.Write(file.content); werr != nil {
			return closeWithErr(werr, w, f)
		}
	}

	// Check errors on close operations
	if cerr := w.Close(); cerr != nil {
		_ = f.Close()
		return cerr
	}
	return f.Close()
}

func createTarGzArchive(path string, files []fileEntry) error {
	f, terr := os.Create(path)
	if terr != nil {
		return terr
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Uid:  0,
			Gid:  0,
			Mode: defaultFileMode,
			Size: int64(len(file.content)),
			ModTime: func() time.Time {
				if file.modTime.IsZero() {
					return time.Now()
				}
				return file.modTime
			}(),
		}
		// Set additional times if supported
		hdr.AccessTime = hdr.ModTime
		if !file.createTime.IsZero() {
			hdr.ChangeTime = file.createTime
		} else {
			hdr.ChangeTime = hdr.ModTime
		}
		if herr := tw.WriteHeader(hdr); herr != nil {
			return closeWithErr(herr, tw, gw, f)
		}
		if _, werr := tw.Write(file.content); werr != nil {
			return closeWithErr(werr, tw, gw, f)
		}
	}

	// Check errors on close operations in the correct order
	if cerr := tw.Close(); cerr != nil {
		_ = gw.Close()
		_ = f.Close()
		return cerr
	}
	if cerr := gw.Close(); cerr != nil {
		_ = f.Close()
		return cerr
	}
	return f.Close()
}

func generateHashFile(path string, files []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, file := range files {
		data, rerr := os.ReadFile(file)
		if rerr != nil {
			return rerr
		}

		hash := sha256.Sum256(data)
		fmt.Fprintf(f, "%x  %s\n", hash, filepath.Base(file))
	}

	return nil
}

// makeUniqueName ensures that each file name within the archive is unique by appending
// an incremental numeric suffix before the extension when collisions occur.
// Example: "root.pem" -> "root-2.pem", "root-3.pem", etc.
func makeUniqueName(name string, used map[string]int) string {
	// If unused, mark and return as-is
	if _, ok := used[name]; !ok {
		used[name] = 1
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	// Track a counter per base+ext key
	key := base + ext
	counter := max(used[key], 1)
	for {
		counter++
		candidate := fmt.Sprintf("%s-%d%s", base, counter, ext)
		if _, exists := used[candidate]; !exists {
			used[key] = counter
			used[candidate] = 1
			return candidate
		}
	}
}

// fileModTime returns the file's modification time, or time.Now() if stat fails.
func fileModTime(path string) time.Time {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Now()
	}
	return fi.ModTime()
}

// maxModTime returns the latest modification time across provided paths.
// If the list is empty or stats fail, returns time.Now().
func maxModTime(paths []string) time.Time {
	var zero time.Time
	maxTime := zero
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		mt := fi.ModTime()
		if maxTime.IsZero() || mt.After(maxTime) {
			maxTime = mt
		}
	}
	if maxTime.IsZero() {
		return time.Now()
	}
	return maxTime
}
