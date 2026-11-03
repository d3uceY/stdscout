package backend

import (
	"os"
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
		if loc.Count <= 0 {
			t.Errorf("%q counted %d binaries", loc.Path, loc.Count)
		}
		if loc.Label == "" {
			t.Errorf("%q has no label", loc.Path)
		}
		total += loc.Count
	}
	if total != len(binaries) {
		t.Errorf("locations hold %d binaries, List returned %d", total, len(binaries))
	}
}
