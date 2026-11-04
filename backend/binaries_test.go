package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// List is the whole feature: if it returns dirs, non-executables, or duplicates
// the UI shows junk, so check those three things on the real PATH.
func TestList(t *testing.T) {
	got, err := (&Binaries{}).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("no binaries found on PATH")
	}
	t.Logf("found %d binaries", len(got))

	seen := map[string]bool{}
	for _, b := range got {
		if seen[strings.ToLower(b.Path)] {
			t.Errorf("duplicate path %q", b.Path)
		}
		seen[strings.ToLower(b.Path)] = true

		info, err := os.Stat(b.Path)
		if err != nil {
			t.Errorf("%q: %v", b.Path, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("%q is a directory", b.Path)
		}
		if b.Name != info.Name() {
			t.Errorf("name %q does not match %q", b.Name, info.Name())
		}
	}
}

// Locations is a view over List, so the counts have to add back up to it.
func TestLocations(t *testing.T) {
	service := &Binaries{}

	binaries, err := service.List()
	if err != nil {
		t.Fatal(err)
	}
	locations, err := service.Locations()
	if err != nil {
		t.Fatal(err)
	}
	if len(locations) == 0 {
		t.Fatal("no locations")
	}

	total := 0
	for _, loc := range locations {
		if loc.Label == "" {
			t.Errorf("%q has no label", loc.Path)
		}
		if loc.Dead {
			if loc.Count != 0 {
				t.Errorf("dead %q counted %d binaries", loc.Path, loc.Count)
			}
			continue
		}
		if loc.Count <= 0 {
			t.Errorf("%q counted %d binaries", loc.Path, loc.Count)
		}
		total += loc.Count
	}
	if total != len(binaries) {
		t.Errorf("locations hold %d binaries, List returned %d", total, len(binaries))
	}
}

// A label with a logo that is not on disk is a broken image in the grid and a
// broken stack in the table, and nothing else would catch it: the Go side never
// touches the file.
func TestKnownDirLogosExist(t *testing.T) {
	seen := map[string]bool{}
	for _, known := range knownDir {
		if known.logo == "" || seen[known.logo] {
			continue
		}
		seen[known.logo] = true

		path := filepath.Join("..", "frontend", "public", "logos", known.logo+".svg")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%q points at %s, which is missing", known.label, path)
		}
	}
	if len(seen) == 0 {
		t.Fatal("no known dir has a logo")
	}
}

// The table is order sensitive, and the failure mode is silent: a generic
// fragment listed too early labels half the machine "Windows". These are the
// pairs where the specific entry has to win, plus the installer-root aliases.
func TestDescribeDir(t *testing.T) {
	cases := []struct{ dir, want string }{
		{`C:\Windows\System32\WindowsPowerShell\v1.0`, "PowerShell"},
		{`C:\Windows\System32\Wbem`, "WMI"},
		{`C:\Program Files (x86)\Windows Kits\10\bin`, "Windows SDK"},
		{`C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.4\bin`, "CUDA"},
		{`C:\Program Files\NVIDIA Corporation\NVSMI`, "NVIDIA"},
		{`C:\Program Files\Microsoft Visual Studio\2022\Community\MSBuild\Current\Bin`, "Visual Studio"},
		{`C:\Program Files\Go\bin`, "Go"},
		{`C:\Users\me\go\bin`, "Go"},
		{`C:\Program Files\Git\cmd`, "Git"},
		{`C:\Users\me\AppData\Local\GitHub CLI`, "GitHub CLI"},
		{`C:\Users\me\AppData\Roaming\npm`, "npm"},
		{`C:\Users\me\AppData\Local\pnpm`, "pnpm"},
		{`C:\Users\me\AppData\Local\Programs\Microsoft VS Code\bin`, "VS Code"},
		{`C:\Users\me\AppData\Local\Programs\Ollama`, "Ollama"},
		{`C:\Users\me\AppData\Local\Android\Sdk\platform-tools`, "Android SDK"},
		{`C:\flutter\bin\cache\dart-sdk\bin`, "Flutter"},
		{`C:\Users\me\.cargo\bin`, "Cargo"},
		// A winget cache wraps the tool, so the tool has to win.
		{`C:\Users\me\AppData\Local\Microsoft\WinGet\Packages\Brecht.WinLibs_Microsoft.Winget.Source_x\mingw64\bin`, "MinGW"},
		{`C:\Users\me\AppData\Local\Microsoft\WinGet\Packages\Gyan.FFmpeg_Microsoft.Winget.Source_x\ffmpeg-9-build\bin`, "FFmpeg"},
		{`C:\Users\me\AppData\Local\Microsoft\WinGet\Links`, "winget"},
		// Unknown folders keep their parent: a bare "Binn" identifies nothing.
		{`C:\Program Files\Acme Suite\Binn`, `Acme Suite\Binn`},
		{`C:\Tools\whatever`, `Tools\whatever`},
		{`D:\loose`, "loose"},
	}

	for _, c := range cases {
		if got, _ := describeDir(c.dir); got != c.want {
			t.Errorf("describeDir(%q) = %q, want %q", c.dir, got, c.want)
		}
	}
}
