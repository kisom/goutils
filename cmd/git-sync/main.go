package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// repo represents a git repository to sync.
type repo struct {
	path string
	name string
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fatalf("cannot resolve path %q: %v", root, err)
	}

	repos, err := discoverRepos(absRoot)
	if err != nil {
		fatalf("discovery failed: %v", err)
	}

	if len(repos) == 0 {
		fatalf("no git repositories found under %s", absRoot)
	}

	fmt.Printf("Found %d repo(s) to sync.\n\n", len(repos))

	var failed []string
	for _, r := range repos {
		if err := syncRepo(r); err != nil {
			fmt.Printf("   %s: %v\n\n", r.name, err)
			failed = append(failed, r.name)
		}
	}

	fmt.Println(strings.Repeat("", 60))
	if len(failed) > 0 {
		fmt.Printf("Done. %d/%d succeeded; failures: %s\n",
			len(repos)-len(failed), len(repos), strings.Join(failed, ", "))
		os.Exit(1)
	}
	fmt.Printf("Done. All %d repo(s) synced.\n", len(repos))
}

// discoverRepos finds the root repo and any immediate child repos.
func discoverRepos(root string) ([]repo, error) {
	var repos []repo

	// Check if root itself is a git repo.
	if isGitRepo(root) {
		repos = append(repos, repo{path: root, name: filepath.Base(root) + " (root)"})
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		child := filepath.Join(root, e.Name())
		if isGitRepo(child) {
			repos = append(repos, repo{path: child, name: e.Name()})
		}
	}

	return repos, nil
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	// .git can be a directory or a file (worktrees / submodules).
	return info.IsDir() || info.Mode().IsRegular()
}

// syncRepo performs the full sync sequence on a single repository:
//
//	fetch --all --prune → stash (if dirty) → pull --rebase each remote →
//	stash pop → push each remote → push tags each remote
func syncRepo(r repo) error {
	fmt.Printf(" %s (%s)\n", r.name, r.path)

	rems, err := remotes(r.path)
	if err != nil {
		return fmt.Errorf("listing remotes: %w", err)
	}
	if len(rems) == 0 {
		fmt.Println("  no remotes configured, skipping")
		return nil
	}

	branch, err := currentBranch(r.path)
	if err != nil {
		return fmt.Errorf("determining current branch: %w", err)
	}

	// 1. Fetch all remotes.
	if err := git(r.path, "fetch", "--all", "--prune"); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	// 2. Stash if the working tree is dirty.
	dirty, err := isDirty(r.path)
	if err != nil {
		return fmt.Errorf("checking dirty state: %w", err)
	}
	if dirty {
		fmt.Println("  stashing uncommitted changes")
		if err := git(r.path, "stash", "push", "-m", "git-sync auto-stash"); err != nil {
			return fmt.Errorf("stash push: %w", err)
		}
	}

	// 3. Pull --rebase from each remote.
	var pullErr error
	for _, rem := range rems {
		if err := git(r.path, "pull", "--rebase", rem, branch); err != nil {
			fmt.Printf("  pull from %s failed: %v\n", rem, err)
			pullErr = fmt.Errorf("pull from %s: %w", rem, err)
		}
	}

	// 4. Pop stash regardless of pull outcome (best effort to restore state).
	if dirty {
		fmt.Println("  restoring stashed changes")
		if popErr := git(r.path, "stash", "pop"); popErr != nil {
			fmt.Printf("  stash pop failed: %v (changes remain in stash)\n", popErr)
		}
	}

	if pullErr != nil {
		return pullErr
	}

	// 5. Push to each remote.
	for _, rem := range rems {
		if err := git(r.path, "push", rem); err != nil {
			return fmt.Errorf("push to %s: %w", rem, err)
		}
	}

	// 6. Push tags to each remote.
	for _, rem := range rems {
		if err := git(r.path, "push", "--tags", rem); err != nil {
			return fmt.Errorf("push tags to %s: %w", rem, err)
		}
	}

	fmt.Println("  synced")
	return nil
}

// remotes returns the list of git remote names for a repository.
func remotes(dir string) ([]string, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}
	return names, nil
}

// currentBranch returns the current branch name.
func currentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// isDirty returns true if the working tree or index has uncommitted changes.
func isDirty(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// git runs a git command in the given directory, forwarding stderr.
func git(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
