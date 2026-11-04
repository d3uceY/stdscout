package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Binary is one runnable file found on PATH.
type Binary struct {
	Name string `json:"name"`
	Path string `json:"path"`
	// Logo is the brand mark of the folder the file lives in (/logos/<logo>.svg),
	// empty when the folder is not a known one.
	Logo string `json:"logo"`
	// Shadowed means an earlier directory on PATH holds a file of the same
	// name, so this copy is the one that does NOT run.
	Shadowed bool `json:"shadowed"`
	// System means the file lives under a folder Windows owns, not one the
	// user installed something into.
	System bool `json:"system"`
}

// scanResult is everything one pass over PATH knows.
type scanResult struct {
	binaries []Binary
	dead     []string // directories that are on PATH but cannot be read
}

// scan is the whole feature in one pass.
//
// Directories are visited in PATH order, so "the copy that runs" is simply
// "the first time this name is seen" — no sorting required, and the answer
// stays right after the list is sorted for display. Directories that no longer
// exist are collected instead of dropped: a stale PATH entry is the usual
// reason a command that used to work has stopped.
func scan(dirs []string) scanResult {
	return scanWith(dirs, systemRoots())
}

// scanWith takes the system roots as an argument so the rule can be tested
// without depending on the machine's environment.
func scanWith(dirs []string, roots []string) scanResult {
	seenPath := make(map[string]bool)
	seenName := make(map[string]bool)
	seenDead := make(map[string]bool)
	var res scanResult

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if key := strings.ToLower(dir); !seenDead[key] {
				seenDead[key] = true
				res.dead = append(res.dead, dir)
			}
			continue
		}

		_, dirLogo := knownLabel(dir)

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil || info.IsDir() || !isRunnable(entry.Name(), info.Mode()) {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			pathKey := strings.ToLower(path)
			if seenPath[pathKey] {
				continue // the same directory is often on PATH twice
			}
			seenPath[pathKey] = true

			nameKey := strings.ToLower(entry.Name())
			res.binaries = append(res.binaries, Binary{
				Name:     entry.Name(),
				Path:     path,
				Logo:     dirLogo,
				Shadowed: seenName[nameKey],
				System:   isSystem(path, roots),
			})
			seenName[nameKey] = true
		}
	}
	return res
}

// pathDirs is PATH split into directories, in priority order. Empty segments
// are dropped because Windows ignores them.
func pathDirs() []string {
	var dirs []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir = strings.TrimSpace(dir); dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// systemRoots are the folders Windows itself owns. Anything under them shipped
// with the OS or the Store, so it is noise in a list of what the user installed.
func systemRoots() []string {
	var roots []string
	add := func(p string) {
		if p != "" {
			roots = append(roots, strings.ToLower(filepath.Clean(p)))
		}
	}

	add(os.Getenv("SystemRoot"))
	if len(roots) == 0 {
		add(`C:\Windows`) // a PATH this app cannot see is worse than a guess
	}
	add(joinEnv("ProgramFiles", "WindowsApps"))
	add(joinEnv("LOCALAPPDATA", "Microsoft", "WindowsApps"))
	return roots
}

func joinEnv(env string, rest ...string) string {
	base := os.Getenv(env)
	if base == "" {
		return ""
	}
	return filepath.Join(append([]string{base}, rest...)...)
}

// isSystem reports whether path is one of the roots or lives under one.
func isSystem(path string, roots []string) bool {
	lower := strings.ToLower(path)
	for _, root := range roots {
		if lower == root || strings.HasPrefix(lower, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// isRunnable reports whether a PATH entry is something you can execute.
// Windows has no exec bit, so PATHEXT is the filter it actually uses.
func isRunnable(name string, mode os.FileMode) bool {
	if runtime.GOOS != "windows" {
		return mode&0o111 != 0
	}
	ext := strings.ToLower(filepath.Ext(name))
	for _, candidate := range strings.Split(strings.ToLower(os.Getenv("PATHEXT")), ";") {
		if candidate != "" && ext == candidate {
			return true
		}
	}
	return false
}
