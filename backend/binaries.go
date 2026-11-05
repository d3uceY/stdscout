// Package backend is everything the frontend is allowed to call.
package backend

import (
	"sort"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Binaries is the service registered with Wails in main.go. Every exported
// method becomes a JS binding.
type Binaries struct {
	// AppVersion is the release stamped in by the linker; main.go passes it on.
	AppVersion string
}

// Version is what the top bar prints beside the name.
func (b *Binaries) Version() string {
	return b.AppVersion
}

// List walks PATH and returns every runnable file, sorted by name.
//
// This is the Go equivalent of `Get-Command -CommandType Application`: same
// source (PATH), same filter (PATHEXT on Windows, the exec bit elsewhere), but
// no shell to spawn, so no quoting, no encoding, no console truncation — the
// paths are always complete. On top of that each entry says whether it is the
// copy that actually runs and whether Windows owns it.
func (b *Binaries) List() ([]Binary, error) {
	out := scan(pathDirs()).binaries

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