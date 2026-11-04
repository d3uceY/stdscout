package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fixture builds a throwaway PATH tree:
//
//	first/  tool.exe, notes.txt, sub/     (a directory must never be listed)
//	second/ TOOL.EXE, other.cmd           (TOOL.EXE shadows nothing: tool.exe came first)
//	gone/                                  does not exist
func fixture(t *testing.T) (first, second, gone string) {
	t.Helper()
	t.Setenv("PATHEXT", ".EXE;.CMD")

	root := t.TempDir()
	first = filepath.Join(root, "first")
	second = filepath.Join(root, "second")
	gone = filepath.Join(root, "gone")

	for _, dir := range []string{first, second} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{
		filepath.Join(first, "tool.exe"),
		filepath.Join(first, "notes.txt"),
		filepath.Join(second, "TOOL.EXE"),
		filepath.Join(second, "other.cmd"),
	} {
		if err := os.WriteFile(file, []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(first, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	return first, second, gone
}

// scan is the whole feature, so the fixture checks all four jobs at once:
// only runnables, one entry per path, the winner marked, dead dirs collected.
func TestScan(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PATHEXT only exists on Windows")
	}
	first, second, gone := fixture(t)

	res := scanWith([]string{first, gone, second, first}, nil)

	if len(res.binaries) != 3 {
		t.Fatalf("got %d binaries, want 3: %+v", len(res.binaries), res.binaries)
	}

	byPath := map[string]Binary{}
	for _, b := range res.binaries {
		byPath[b.Path] = b
	}

	winner := byPath[filepath.Join(first, "tool.exe")]
	if winner.Name != "tool.exe" || winner.Shadowed {
		t.Errorf("first copy = %+v, want unshadowed tool.exe", winner)
	}

	loser := byPath[filepath.Join(second, "TOOL.EXE")]
	if !loser.Shadowed {
		t.Errorf("TOOL.EXE should be shadowed by tool.exe (names match case-insensitively)")
	}

	if _, ok := byPath[filepath.Join(second, "other.cmd")]; !ok {
		t.Error("other.cmd missing from the scan")
	}
	if _, ok := byPath[filepath.Join(first, "notes.txt")]; ok {
		t.Error("notes.txt is not runnable and must not be listed")
	}
	if _, ok := byPath[filepath.Join(first, "sub")]; ok {
		t.Error("a directory must not be listed")
	}

	if len(res.dead) != 1 || res.dead[0] != gone {
		t.Errorf("dead = %v, want [%s]", res.dead, gone)
	}
	if res.binaries[0].System || res.binaries[1].System || res.binaries[2].System {
		t.Error("temp dirs are not system folders")
	}
}

// The system flag is what lets the table hide 80% of a Windows PATH, so the
// prefix rule has to be exact: a sibling folder called WindowsApps-extra is not
// inside C:\WindowsApps.
func TestIsSystem(t *testing.T) {
	roots := []string{`c:\windows`, `c:\program files\windowsapps`}
	cases := []struct {
		path string
		want bool
	}{
		{`C:\Windows\System32\cmd.exe`, true},
		{`c:\windows\system32\WindowsPowerShell\v1.0\powershell.exe`, true},
		{`C:\Program Files\WindowsApps\Microsoft.WindowsTerminal\wt.exe`, true},
		{`C:\WindowsApps\tool.exe`, false},
		{`C:\Program Files\nodejs\npm.cmd`, false},
		{`D:\Windows\System32\cmd.exe`, false},
		// The root itself has to count, or the WindowsApps folder shows up as a
		// user folder that nobody installed.
		{`C:\Program Files\WindowsApps`, true},
	}
	for _, c := range cases {
		if got := isSystem(c.path, roots); got != c.want {
			t.Errorf("isSystem(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// A dead directory has no binaries, so Locations has to carry it itself or the
// user never finds out their PATH points at an uninstalled tool.
func TestLocationsReportsDeadDirs(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PATHEXT only exists on Windows")
	}
	first, _, gone := fixture(t)

	t.Setenv("PATH", first+string(os.PathListSeparator)+gone)
	locations, err := (&Binaries{}).Locations()
	if err != nil {
		t.Fatal(err)
	}

	var dead *Location
	for i := range locations {
		if locations[i].Dead {
			dead = &locations[i]
		}
	}
	if dead == nil {
		t.Fatal("the missing directory was not reported")
	}
	if dead.Path != gone || dead.Count != 0 {
		t.Errorf("dead location = %+v, want path %q with count 0", *dead, gone)
	}
	if len(locations) == 0 || locations[0].Dead == false {
		t.Error("dead folders sort first: they are the actionable ones")
	}
}
