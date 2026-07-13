package db

import (
	"io/fs"
	"sort"
	"testing"
)

// TestMigrationFilenamesArePadded guards the zero-padding invariant Task
// #16 requires: migration filenames must sort identically whether
// compared as strings (what the OS and goose's directory walk actually
// do) or as the numeric prefix they encode (what determines migration
// order). A non-padded filename (e.g. "2_foo.sql" next to
// "10_bar.sql") would sort those two out of numeric order — "10_"
// lexicographically precedes "2_" — silently changing the applied
// order without any error.
func TestMigrationFilenamesArePadded(t *testing.T) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	if len(names) == 0 {
		t.Fatal("no migration files found")
	}

	lexical := append([]string(nil), names...)
	sort.Strings(lexical)

	numeric := append([]string(nil), names...)
	sort.SliceStable(numeric, func(i, j int) bool {
		return migrationSeq(t, numeric[i]) < migrationSeq(t, numeric[j])
	})

	for i := range lexical {
		if lexical[i] != numeric[i] {
			t.Fatalf("migration filenames are not zero-padded consistently: lexical order %v != numeric order %v", lexical, numeric)
		}
	}
}

// migrationSeq extracts the leading digit run goose treats as the
// migration's sequence number (e.g. "00001" from
// "00001_create_notes.sql").
func migrationSeq(t *testing.T, name string) int {
	t.Helper()
	var n int
	i := 0
	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		n = n*10 + int(name[i]-'0')
		i++
	}
	if i == 0 {
		t.Fatalf("migration filename %q has no leading sequence number", name)
	}
	return n
}
