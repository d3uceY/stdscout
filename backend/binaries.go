// Package backend is everything the frontend is allowed to call.
package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Binary is one executable found on PATH.
type Binary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Binaries is the service registered with Wails in main.go. Every exported
// method becomes a JS binding.
type Binaries struct{}

// List walks PATH and returns every runnable file, sorted by name.
//
// This is the Go equivalent of `Get-Command -CommandType Application`: same
// source (PATH), same filter (PATHEXT on Windows, the exec bit elsewhere), but
// no shell to spawn, so no quoting, no encoding, no console truncation — the
// paths are always complete.
func (b *Binaries) List() ([]Binary, error) {
	found := make(map[string]bool)
	var out []Binary

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // a stale PATH entry is normal, not an error worth failing on
		}
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil || info.IsDir() || !isRunnable(entry.Name(), info.Mode()) {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			// The same directory is often on PATH twice; key on the resolved path.
			if key := strings.ToLower(path); !found[key] {
				found[key] = true
				out = append(out, Binary{Name: entry.Name(), Path: path})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		if a == b {
			return out[i].Path < out[j].Path
		}
		return a < b
	})
	return out, nil
}

// Reveal opens the system file manager with the binary selected.
func (b *Binaries) Reveal(path string) error {
	return application.Get().Env.OpenFileManager(path, true)
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