package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"gopkg.in/yaml.v2"
)

// Config represents the top-level YAML configuration
type Config struct {
	Config struct {
		Hashes string `yaml:"hashes"`
		Expiry string `yaml:"expiry"`
	} `yaml:"config"`
	Chains map[string]ChainGroup `yaml:"chains"`
}

// ChainGroup represents a named group of certificate chains
type ChainGroup struct {
	Certs   []CertChain `yaml:"certs"`
	Outputs Outputs     `yaml:"outputs"`
}

// CertChain represents a root certificate and its intermediates
type CertChain struct {
	Root          string   `yaml:"root"`
	Intermediates []string `yaml:"intermediates"`
}

// Outputs defines output format options
type Outputs struct {
	IncludeSingle     bool     `yaml:"include_single"`
	IncludeIndividual bool     `yaml:"include_individual"`
	Manifest          bool     `yaml:"manifest"`
	Formats           []string `yaml:"formats"`
	Encoding          string   `yaml:"encoding"`
}

var (
	configFile string
	outputDir  string
)

var formatExtensions = map[string]string{
	"zip": ".zip",
	"tgz": ".tar.gz",
}

//go:embed README.txt
var readmeContent string

func usage() {
	fmt.Fprint(os.Stderr, readmeContent)
}

func main() {
	flag.Usage = usage
	flag.StringVar(&configFile, "c", "bundle.yaml", "path to YAML configuration file")
	flag.StringVar(&outputDir, "o", "pkg", "output directory for archives")
	flag.Parse()

	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file required (-c flag)\n")
		os.Exit(1)
	}

	// Load and parse configuration
	cfg, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Parse expiry duration (default 1 year)
	expiryDuration := 365 * 24 * time.Hour
	if cfg.Config.Expiry != "" {
		expiryDuration, err = parseDuration(cfg.Config.Expiry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing expiry: %v\n", err)
			os.Exit(1)
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Process each chain group
	// Pre-allocate createdFiles based on total number of formats across all groups
	totalFormats := 0
	for _, group := range cfg.Chains {
		totalFormats += len(group.Outputs.Formats)
	}
	createdFiles := make([]string, 0, totalFormats)
	for groupName, group := range cfg.Chains {
		files, err := processChainGroup(groupName, group, expiryDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing chain group %s: %v\n", groupName, err)
			os.Exit(1)
		}
		createdFiles = append(createdFiles, files...)
	}

	// Generate hash file for all created archives
	if cfg.Config.Hashes != "" {
		hashFile := filepath.Join(outputDir, cfg.Config.Hashes)
		if err := generateHashFile(hashFile, createdFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating hash file: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Certificate bundling completed successfully")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
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

func processChainGroup(groupName string, group ChainGroup, expiryDuration time.Duration) ([]string, error) {
	// Default encoding to "pem" if not specified
	encoding := group.Outputs.Encoding
	if encoding == "" {
		encoding = "pem"
	}

	// Collect certificates from all chains in the group
	singleFileCerts, individualCerts, err := loadAndCollectCerts(group.Certs, group.Outputs, expiryDuration)
	if err != nil {
		return nil, err
	}

	// Prepare files for inclusion in archives
	archiveFiles, err := prepareArchiveFiles(singleFileCerts, individualCerts, group.Outputs, encoding)
	if err != nil {
		return nil, err
	}

	// Create archives for the entire group
	createdFiles, err := createArchiveFiles(groupName, group.Outputs.Formats, archiveFiles)
	if err != nil {
		return nil, err
	}

	return createdFiles, nil
}

// loadAndCollectCerts loads all certificates from chains and collects them for processing
func loadAndCollectCerts(chains []CertChain, outputs Outputs, expiryDuration time.Duration) ([]*x509.Certificate, []certWithPath, error) {
	var singleFileCerts []*x509.Certificate
	var individualCerts []certWithPath

	for _, chain := range chains {
		// Load root certificate
		rootCert, err := certlib.LoadCertificate(chain.Root)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load root certificate %s: %v", chain.Root, err)
		}

		// Check expiry for root
		checkExpiry(chain.Root, rootCert, expiryDuration)

		// Add root to collections if needed
		if outputs.IncludeSingle {
			singleFileCerts = append(singleFileCerts, rootCert)
		}
		if outputs.IncludeIndividual {
			individualCerts = append(individualCerts, certWithPath{
				cert: rootCert,
				path: chain.Root,
			})
		}

		// Load and validate intermediates
		for _, intPath := range chain.Intermediates {
			intCert, err := certlib.LoadCertificate(intPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load intermediate certificate %s: %v", intPath, err)
			}

			// Validate that intermediate is signed by root
			if err := intCert.CheckSignatureFrom(rootCert); err != nil {
				return nil, nil, fmt.Errorf("intermediate %s is not properly signed by root %s: %v", intPath, chain.Root, err)
			}

			// Check expiry for intermediate
			checkExpiry(intPath, intCert, expiryDuration)

			// Add intermediate to collections if needed
			if outputs.IncludeSingle {
				singleFileCerts = append(singleFileCerts, intCert)
			}
			if outputs.IncludeIndividual {
				individualCerts = append(individualCerts, certWithPath{
					cert: intCert,
					path: intPath,
				})
			}
		}
	}

	return singleFileCerts, individualCerts, nil
}

// prepareArchiveFiles prepares all files to be included in archives
func prepareArchiveFiles(singleFileCerts []*x509.Certificate, individualCerts []certWithPath, outputs Outputs, encoding string) ([]fileEntry, error) {
	var archiveFiles []fileEntry

	// Handle a single bundle file
	if outputs.IncludeSingle && len(singleFileCerts) > 0 {
		files, err := encodeCertsToFiles(singleFileCerts, "bundle", encoding, true)
		if err != nil {
			return nil, fmt.Errorf("failed to encode single bundle: %v", err)
		}
		archiveFiles = append(archiveFiles, files...)
	}

	// Handle individual files
	if outputs.IncludeIndividual {
		for _, cp := range individualCerts {
			baseName := strings.TrimSuffix(filepath.Base(cp.path), filepath.Ext(cp.path))
			files, err := encodeCertsToFiles([]*x509.Certificate{cp.cert}, baseName, encoding, false)
			if err != nil {
				return nil, fmt.Errorf("failed to encode individual cert %s: %v", cp.path, err)
			}
			archiveFiles = append(archiveFiles, files...)
		}
	}

	// Generate manifest if requested
	if outputs.Manifest {
		manifestContent := generateManifest(archiveFiles)
		archiveFiles = append(archiveFiles, fileEntry{
			name:    "MANIFEST",
			content: manifestContent,
		})
	}

	return archiveFiles, nil
}

// createArchiveFiles creates archive files in the specified formats
func createArchiveFiles(groupName string, formats []string, archiveFiles []fileEntry) ([]string, error) {
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
				return nil, fmt.Errorf("failed to create zip archive: %v", err)
			}
		case "tgz":
			if err := createTarGzArchive(archivePath, archiveFiles); err != nil {
				return nil, fmt.Errorf("failed to create tar.gz archive: %v", err)
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
			fmt.Fprintf(os.Stderr, "WARNING: Certificate %s has EXPIRED (expired %d days ago)\n", path, -daysUntilExpiry)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Certificate %s will expire in %d days (on %s)\n", path, daysUntilExpiry, cert.NotAfter.Format("2006-01-02"))
		}
	}
}

type fileEntry struct {
	name    string
	content []byte
}

type certWithPath struct {
	cert *x509.Certificate
	path string
}

// encodeCertsToFiles converts certificates to file entries based on encoding type
// If isSingle is true, certs are concatenated into a single file; otherwise one cert per file
func encodeCertsToFiles(certs []*x509.Certificate, baseName string, encoding string, isSingle bool) ([]fileEntry, error) {
	var files []fileEntry

	switch encoding {
	case "pem":
		pemContent := encodeCertsToPEM(certs)
		files = append(files, fileEntry{
			name:    baseName + ".pem",
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
		} else {
			// Individual DER file (should only have one cert)
			if len(certs) > 0 {
				files = append(files, fileEntry{
					name:    baseName + ".crt",
					content: certs[0].Raw,
				})
			}
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
		} else {
			if len(certs) > 0 {
				files = append(files, fileEntry{
					name:    baseName + ".crt",
					content: certs[0].Raw,
				})
			}
		}
	default:
		return nil, fmt.Errorf("unsupported encoding: %s (must be 'pem', 'der', or 'both')", encoding)
	}

	return files, nil
}

// encodeCertsToPEM encodes certificates to PEM format
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
	var manifest strings.Builder
	for _, file := range files {
		if file.name == "MANIFEST" {
			continue
		}
		hash := sha256.Sum256(file.content)
		manifest.WriteString(fmt.Sprintf("%x  %s\n", hash, file.name))
	}
	return []byte(manifest.String())
}

func createZipArchive(path string, files []fileEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	w := zip.NewWriter(f)

	for _, file := range files {
		fw, err := w.Create(file.name)
		if err != nil {
			w.Close()
			f.Close()
			return err
		}
		if _, err := fw.Write(file.content); err != nil {
			w.Close()
			f.Close()
			return err
		}
	}

	// Check errors on close operations
	if err := w.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func createTarGzArchive(path string, files []fileEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: 0644,
			Size: int64(len(file.content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			tw.Close()
			gw.Close()
			f.Close()
			return err
		}
		if _, err := tw.Write(file.content); err != nil {
			tw.Close()
			gw.Close()
			f.Close()
			return err
		}
	}

	// Check errors on close operations in the correct order
	if err := tw.Close(); err != nil {
		gw.Close()
		f.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		f.Close()
		return err
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
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		hash := sha256.Sum256(data)
		fmt.Fprintf(f, "%x  %s\n", hash, filepath.Base(file))
	}

	return nil
}
