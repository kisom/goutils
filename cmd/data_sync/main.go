package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.wntrmute.dev/kyle/goutils/config"
	"git.wntrmute.dev/kyle/goutils/fileutil"
	"git.wntrmute.dev/kyle/goutils/log"
)

func mustHostname() string {
	hostname, err := os.Hostname()
	log.FatalError(err, "couldn't retrieve hostname")

	if hostname == "" {
		log.Fatal("no hostname returned")
	}
	return strings.Split(hostname, ".")[0]
}

var (
	defaultDataDir   = mustHostname() + "_data"
	defaultProgName  = defaultDataDir + "_sync"
	defaultMountDir  = filepath.Join("/media", os.Getenv("USER"), defaultDataDir)
	defaultSyncDir   = os.Getenv("HOME")
	defaultTargetDir = filepath.Join(defaultMountDir, os.Getenv("USER"))
)

func usage(w io.Writer) {
	prog := filepath.Base(os.Args[0])
	fmt.Fprintf(w, `Usage: %s [-d path] [-l level] [-m path] [-nqsv]
				  [-t path]
	-d path		path to sync source directory
			(default "%s")
	-l level	log level to output (default "INFO"). Valid log
			levels are DEBUG, INFO, NOTICE, WARNING, ERR,
			CRIT, ALERT, EMERG. The default is INFO.
	-m path		path to sync mount directory
			(default "%s")
	-n		dry-run mode: only check paths and print files to
			exclude
	-q		suppress console output
	-s		suppress syslog output
	-t path		path to sync target directory
			(default "%s")
	-v		verbose rsync output

%s rsyncs the tree at the sync source directory (-d) to the sync target
directory (-t); it checks the mount directory (-m) exists; the sync target
target directory must exist on the mount directory.

`, prog, defaultSyncDir, defaultMountDir, defaultTargetDir, prog)
}

func checkPaths(mount, target string, dryRun bool) error {
	if !fileutil.DirectoryDoesExist(mount) {
		return fmt.Errorf("sync dir %s isn't mounted", mount)
	}

	if !strings.HasPrefix(target, mount) {
		return fmt.Errorf("target dir %s must exist in %s", target, mount)
	}

	if !fileutil.DirectoryDoesExist(target) {
		if dryRun {
			log.Infof("would mkdir %s", target)
		} else {
			log.Infof("mkdir %s", target)
			if err := os.Mkdir(target, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildExcludes(syncDir string) ([]string, error) {
	var excluded []string

	walker := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			excluded = append(excluded, strings.TrimPrefix(path, syncDir))
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.Mode().IsRegular() {
			if err = fileutil.Access(path, fileutil.AccessRead); err != nil {
				excluded = append(excluded, strings.TrimPrefix(path, syncDir))
			}
		}

		if info.IsDir() {
			if err = fileutil.Access(path, fileutil.AccessExec); err != nil {
				excluded = append(excluded, strings.TrimPrefix(path, syncDir))
			}
		}

		return nil
	}

	err := filepath.Walk(syncDir, walker)
	return excluded, err
}

func writeExcludes(excluded []string) (string, error) {
	if len(excluded) == 0 {
		return "", nil
	}

	excludeFile, err := os.CreateTemp("", defaultProgName)
	if err != nil {
		return "", err
	}

	for _, name := range excluded {
		fmt.Fprintln(excludeFile, name)
	}

	defer excludeFile.Close()
	return excludeFile.Name(), nil
}

func rsync(syncDir, target, excludeFile string, verboseRsync bool) error {
	var args []string

	if excludeFile != "" {
		args = append(args, "--exclude-from")
		args = append(args, excludeFile)
	}

	if verboseRsync {
		args = append(args, "--progress")
		args = append(args, "-v")
	}

	args = append(args, []string{"-au", syncDir + "/", target + "/"}...)

	path, err := exec.LookPath("rsync")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func main() {

	var logLevel, mountDir, syncDir, target string
	var dryRun, quietMode, noSyslog, verboseRsync bool

	flag.StringVar(&syncDir, "d", config.GetDefault("sync_dir", defaultSyncDir),
		"`path to sync source directory`")
	flag.StringVar(&logLevel, "l", config.GetDefault("log_level", "INFO"),
		"log level to output")
	flag.StringVar(&mountDir, "m", config.GetDefault("mount_dir", defaultMountDir),
		"`path` to sync mount directory")
	flag.BoolVar(&dryRun, "n", false, "dry-run mode: only check paths and print files to exclude")
	flag.BoolVar(&quietMode, "q", quietMode, "suppress console output")
	flag.BoolVar(&noSyslog, "s", noSyslog, "suppress syslog output")
	flag.StringVar(&target, "t", config.GetDefault("sync_target", defaultTargetDir),
		"`path` to sync target directory")
	flag.BoolVar(&verboseRsync, "v", false, "verbose rsync output")
	flag.Parse()

	if quietMode && noSyslog {
		fmt.Fprintln(os.Stderr, "both console and syslog output are suppressed")
		fmt.Fprintln(os.Stderr, "errors will NOT be reported")
	}

	logOpts := &log.Options{
		Level:        logLevel,
		Tag:          defaultProgName,
		Facility:     "user",
		WriteSyslog:  !noSyslog,
		WriteConsole: !quietMode,
	}
	err := log.Setup(logOpts)
	log.FatalError(err, "failed to set up logging")

	log.Infof("checking paths: mount=%s, target=%s", mountDir, target)
	err = checkPaths(mountDir, target, dryRun)
	log.FatalError(err, "target dir isn't ready")

	log.Infof("checking for files to exclude from %s", syncDir)
	excluded, err := buildExcludes(syncDir)
	log.FatalError(err, "couldn't build excludes")

	if dryRun {
		fmt.Println("excluded files:")
		for _, path := range excluded {
			fmt.Printf("\t%s\n", path)
		}
		return
	}

	excludeFile, err := writeExcludes(excluded)
	log.FatalError(err, "couldn't write exclude file")
	log.Infof("excluding %d files via %s", len(excluded), excludeFile)

	if excludeFile != "" {
		defer func() {
			log.Infof("removing exclude file %s", excludeFile)
			if err := os.Remove(excludeFile); err != nil {
				log.Warningf("failed to remove temp file %s", excludeFile)
			}
		}()
	}

	err = rsync(syncDir, target, excludeFile, verboseRsync)
	log.FatalError(err, "couldn't sync data")
}
