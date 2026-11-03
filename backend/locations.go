package backend

import (
	"path/filepath"
	"sort"
	"strings"
)

// Location is one directory on PATH, with how many binaries it contributes.
type Location struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Logo  string `json:"logo"` // brand mark in /logos, empty when none is known
	Count int    `json:"count"`
}

// knownDir labels the directories people actually look for, and gives each a
// brand mark. Matching is a substring test against the lowercased path, so the
// deeper fragments have to come before their parents.
var knownDir = []struct{ fragment, label, logo string }{
	{`\windows\system32\windowspowershell`, "PowerShell", ""},
	{`\windows\system32\openssh`, "OpenSSH", ""},
	{`\windows\system32\wbem`, "WMI", "windows"},
	{`\windows\system32`, "System32", "windows"},
	{`\windows\syswow64`, "SysWOW64", "windows"},
	{`\windowsapps`, "WindowsApps", "windows"},
	{`\windows`, "Windows", "windows"},
	{`\microsoft vs code\bin`, "VS Code", ""},
	{`\appdata\roaming\npm`, "npm", "npm"},
	{`\.bun\bin`, "Bun", "bun"},
	{`\.cargo\bin`, "Cargo", "rust"},
	{`\program files\nodejs`, "Node.js", "nodedotjs"},
	{`\program files\go\bin`, "Go", "go"},
	{`\go\bin`, "Go", "go"},
	{`\git\`, "Git", "git"},
	{`\python`, "Python", "python"},
	{`\cloudflare`, "Cloudflare WARP", "cloudflare"},
	{`\ffmpeg`, "FFmpeg", "ffmpeg"},
	{`\dotnet`, ".NET", "dotnet"},
	{`\nsis`, "NSIS", "nsis"},
	{`\mingw64`, "MinGW", ""},
	{`copilot`, "Copilot", "githubcopilot"},
}

// Locations groups the PATH scan by directory: the folders that keep a family
// of binaries together, which is what the grid view shows.
func (b *Binaries) Locations() ([]Location, error) {
	binaries, err := b.List()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(binaries))
	for _, binary := range binaries {
		counts[filepath.Dir(binary.Path)]++
	}

	out := make([]Location, 0, len(counts))
	for dir, count := range counts {
		label, logo := describeDir(dir)
		out = append(out, Location{Path: dir, Label: label, Logo: logo, Count: count})
	}

	// Biggest clusters first: System32 before a single-tool folder.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func describeDir(dir string) (string, string) {
	lower := strings.ToLower(strings.ReplaceAll(dir, "/", `\`))
	for _, known := range knownDir {
		if strings.Contains(lower, known.fragment) {
			return known.label, known.logo
		}
	}
	return filepath.Base(dir), ""
}
