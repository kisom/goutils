package main

import (
	"compress/flate"
	"compress/gzip"
	"encoding/asn1"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"

	goutilslib "git.wntrmute.dev/kyle/goutils/lib"
)

const gzipExt = ".gz"

// kgzExtraID is the two-byte subfield identifier used in the gzip Extra field
// for kgz-specific metadata.
var kgzExtraID = [2]byte{'K', 'G'}

// buildKGExtra constructs the gzip Extra subfield payload for kgz metadata.
//
// The payload is an ASN.1 DER-encoded struct with the following fields:
//
//	Version    INTEGER (currently 1)
//	UID        INTEGER
//	GID        INTEGER
//	Mode       INTEGER (permission bits)
//	CTimeSec   INTEGER (seconds)
//	CTimeNSec  INTEGER (nanoseconds)
//
// The ASN.1 blob is wrapped in a gzip Extra subfield with ID 'K','G'.
func buildKGExtra(uid, gid, mode uint32, ctimeS int64, ctimeNs int32) []byte {
	// Define the ASN.1 structure to encode
	type KGZExtra struct {
		Version   int
		UID       int
		GID       int
		Mode      int
		CTimeSec  int64
		CTimeNSec int32
	}

	payload, err := asn1.Marshal(KGZExtra{
		Version:   1,
		UID:       int(uid),
		GID:       int(gid),
		Mode:      int(mode),
		CTimeSec:  ctimeS,
		CTimeNSec: ctimeNs,
	})
	if err != nil {
		// On marshal failure, return empty to avoid breaking compression
		return nil
	}

	// Wrap in gzip subfield: [ID1 ID2 LEN(lo) LEN(hi) PAYLOAD]
	// Guard against payload length overflow to uint16 for the extra subfield length.
	if len(payload) > int(math.MaxUint16) {
		return nil
	}
	extra := make([]byte, 4+len(payload))
	extra[0] = kgzExtraID[0]
	extra[1] = kgzExtraID[1]
	binary.LittleEndian.PutUint16(extra[2:], uint16(len(payload)&0xFFFF)) //#nosec G115 - masked
	copy(extra[4:], payload)
	return extra
}

// clampToInt32 clamps an int value into the int32 range using a switch to
// satisfy linters that prefer switch over if-else chains for ordered checks.
func clampToInt32(v int) int32 {
	switch {
	case v > int(math.MaxInt32):
		return math.MaxInt32
	case v < int(math.MinInt32):
		return math.MinInt32
	default:
		return int32(v)
	}
}

// buildExtraForPath prepares the gzip Extra field for kgz by collecting
// uid/gid/mode and ctime information, applying any overrides, and encoding it.
func buildExtraForPath(st unix.Stat_t, path string, setUID, setGID int) []byte {
	uid := st.Uid
	gid := st.Gid
	if setUID >= 0 {
		if uint64(setUID) <= math.MaxUint32 {
			uid = uint32(setUID & 0xFFFFFFFF) //#nosec G115 - masked
		}
	}
	if setGID >= 0 {
		if uint64(setGID) <= math.MaxUint32 {
			gid = uint32(setGID & 0xFFFFFFFF) //#nosec G115 - masked
		}
	}
	mode := st.Mode & 0o7777

	// Use portable helper to gather ctime
	var cts int64
	var ctns int32
	if ft, err := goutilslib.LoadFileTime(path); err == nil {
		cts = ft.Changed.Unix()
		ctns = clampToInt32(ft.Changed.Nanosecond())
	}

	return buildKGExtra(uid, gid, mode, cts, ctns)
}

// parseKGExtra scans a gzip Extra blob and returns kgz metadata if present.
func parseKGExtra(extra []byte) (uint32, uint32, uint32, int64, int32, bool) {
	i := 0
	for i+4 <= len(extra) {
		id1 := extra[i]
		id2 := extra[i+1]
		l := int(binary.LittleEndian.Uint16(extra[i+2 : i+4]))
		i += 4
		if i+l > len(extra) {
			break
		}
		if id1 == kgzExtraID[0] && id2 == kgzExtraID[1] {
			// ASN.1 decode payload
			payload := extra[i : i+l]
			var s struct {
				Version   int
				UID       int
				GID       int
				Mode      int
				CTimeSec  int64
				CTimeNSec int32
			}
			if _, err := asn1.Unmarshal(payload, &s); err != nil {
				return 0, 0, 0, 0, 0, false
			}
			if s.Version != 1 {
				return 0, 0, 0, 0, 0, false
			}
			// Validate ranges before converting from int -> uint32 to avoid overflow.
			if s.UID < 0 || s.GID < 0 || s.Mode < 0 {
				return 0, 0, 0, 0, 0, false
			}
			if uint64(s.UID) > math.MaxUint32 || uint64(s.GID) > math.MaxUint32 || uint64(s.Mode) > math.MaxUint32 {
				return 0, 0, 0, 0, 0, false
			}

			return uint32(s.UID & 0xFFFFFFFF), uint32(s.GID & 0xFFFFFFFF),
				uint32(s.Mode & 0xFFFFFFFF), s.CTimeSec, s.CTimeNSec, true //#nosec G115 - masked
		}
		i += l
	}
	return 0, 0, 0, 0, 0, false
}

func compress(path, target string, level int, includeExtra bool, setUID, setGID int) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file for read: %w", err)
	}
	defer sourceFile.Close()

	// Gather file metadata
	var st unix.Stat_t
	if err = unix.Stat(path, &st); err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	fi, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	destFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("opening file for write: %w", err)
	}
	defer destFile.Close()

	gzipCompressor, err := gzip.NewWriterLevel(destFile, level)
	if err != nil {
		return fmt.Errorf("invalid compression level: %w", err)
	}
	// Set header metadata
	gzipCompressor.ModTime = fi.ModTime()
	if includeExtra {
		gzipCompressor.Extra = buildExtraForPath(st, path, setUID, setGID)
	}
	defer gzipCompressor.Close()

	_, err = io.Copy(gzipCompressor, sourceFile)
	if err != nil {
		return fmt.Errorf("compressing file: %w", err)
	}

	return nil
}

func uncompress(path, target string, unrestrict bool, preserveMtime bool) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file for read: %w", err)
	}
	defer sourceFile.Close()

	fi, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("reading file stats: %w", err)
	}

	maxDecompressionSize := fi.Size() * 32

	gzipUncompressor, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("reading gzip headers: %w", err)
	}
	defer gzipUncompressor.Close()

	var reader io.Reader = &io.LimitedReader{
		R: gzipUncompressor,
		N: maxDecompressionSize,
	}

	if unrestrict {
		reader = gzipUncompressor
	}

	destFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("opening file for write: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, reader)
	if err != nil {
		return fmt.Errorf("uncompressing file: %w", err)
	}
	// Apply metadata from Extra (uid/gid/mode) if present
	if gzipUncompressor.Header.Extra != nil {
		if uid, gid, mode, _, _, ok := parseKGExtra(gzipUncompressor.Header.Extra); ok {
			// Chmod
			_ = os.Chmod(target, os.FileMode(mode))
			// Chown (may fail without privileges)
			_ = os.Chown(target, int(uid), int(gid))
		}
	}
	// Preserve mtime if requested
	if preserveMtime {
		mt := gzipUncompressor.Header.ModTime
		if !mt.IsZero() {
			// Set both atime and mtime to mt for simplicity
			_ = os.Chtimes(target, mt, mt)
		}
	}
	return nil
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: %s [-l] [-k] [-m] [-x] [--uid N] [--gid N] source [target]

kgz is like gzip, but supports compressing and decompressing to a different
directory than the source file is in.

Flags:
    -l level    Compression level (0-9). Only meaningful when compressing.
    -u          Do not restrict the size during decompression (gzip bomb guard is 32x).
    -k          Keep the source file (do not remove it after successful (de)compression).
    -m          On decompression, set the file mtime from the gzip header.
    -x          On compression, include uid/gid/mode/ctime in the gzip Extra field.
    --uid N     When used with -x, set UID in Extra to N (overrides source owner).
    --gid N     When used with -x, set GID in Extra to N (overrides source group).
`, os.Args[0])
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func isDir(path string) bool {
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		stat, err2 := file.Stat()
		if err2 != nil {
			return false
		}

		if stat.IsDir() {
			return true
		}
	}

	return false
}

func pathForUncompressing(source, dest string) (string, error) {
	if !isDir(dest) {
		return dest, nil
	}

	source = filepath.Base(source)
	if !strings.HasSuffix(source, gzipExt) {
		return "", fmt.Errorf("%s is a not gzip-compressed file", source)
	}
	outFile := source[:len(source)-len(gzipExt)]
	outFile = filepath.Join(dest, outFile)
	return outFile, nil
}

func pathForCompressing(source, dest string) (string, error) {
	if !isDir(dest) {
		return dest, nil
	}

	source = filepath.Base(source)
	if strings.HasSuffix(source, gzipExt) {
		return "", fmt.Errorf("%s is a gzip-compressed file", source)
	}

	dest = filepath.Join(dest, source+gzipExt)
	return dest, nil
}

func main() {
	var level int
	var path string
	var target = "."
	var err error
	var unrestrict bool
	var keep bool
	var preserveMtime bool
	var includeExtra bool
	var setUID int
	var setGID int

	flag.IntVar(&level, "l", flate.DefaultCompression, "compression level")
	flag.BoolVar(&unrestrict, "u", false, "do not restrict decompression")
	flag.BoolVar(&keep, "k", false, "keep the source file (do not remove it)")
	flag.BoolVar(&preserveMtime, "m", false, "on decompression, set mtime from gzip header")
	flag.BoolVar(&includeExtra, "x", false, "on compression, include uid/gid/mode/ctime in gzip Extra")
	flag.IntVar(&setUID, "uid", -1, "when used with -x, set UID in Extra to this value")
	flag.IntVar(&setGID, "gid", -1, "when used with -x, set GID in Extra to this value")
	flag.Parse()

	if flag.NArg() < 1 || flag.NArg() > 2 {
		usage(os.Stderr)
		os.Exit(1)
	}

	path = flag.Arg(0)
	if flag.NArg() == 2 {
		target = flag.Arg(1)
	}

	if strings.HasSuffix(path, gzipExt) {
		target, err = pathForUncompressing(path, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		err = uncompress(path, target, unrestrict, preserveMtime)
		if err != nil {
			os.Remove(target)
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if !keep {
			_ = os.Remove(path)
		}
		return
	}

	target, err = pathForCompressing(path, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	err = compress(path, target, level, includeExtra, setUID, setGID)
	if err != nil {
		os.Remove(target)
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if !keep {
		_ = os.Remove(path)
	}
}
