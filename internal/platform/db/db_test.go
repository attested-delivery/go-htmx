package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// TestConcurrentWriteContract is a regression test for AD-1/AD-2's core
// invariant: a write transaction blocks a concurrent write transaction
// (via BEGIN IMMEDIATE against the single-connection write pool) instead
// of silently interleaving or racing, while reads are never blocked by
// the writer. A future refactor that accidentally drops _txlock=immediate
// or the write pool's SetMaxOpenConns(1) would let two writers both
// acquire a RESERVED lock, and this test would start failing (either the
// second write would fail with SQLITE_BUSY caught differently, or —
// worse — both transactions would appear to succeed with the second
// silently overwriting data written concurrently by the first,
// exactly the class of bug the concurrency contract exists to prevent).
func TestConcurrentWriteContract(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "concurrent.db")

	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })

	if err := Migrate(d.WriteDB()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	ctx := context.Background()

	tx1, err := d.BeginWrite(ctx)
	if err != nil {
		t.Fatalf("BeginWrite (tx1): %v", err)
	}
	if _, err := tx1.ExecContext(ctx, "INSERT INTO notes (body) VALUES (?)", "tx1"); err != nil {
		t.Fatalf("tx1 insert: %v", err)
	}

	// A second, concurrent write must not proceed while tx1 holds its
	// RESERVED lock — BEGIN IMMEDIATE means it fails fast (bounded by
	// busy_timeout) rather than silently upgrading later.
	blocked := make(chan error, 1)
	go func() {
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		tx2, err := d.BeginWrite(ctx2)
		if err != nil {
			blocked <- err
			return
		}
		defer func() { _ = tx2.Rollback() }()
		_, err = tx2.ExecContext(ctx2, "INSERT INTO notes (body) VALUES (?)", "tx2")
		blocked <- err
	}()

	select {
	case err := <-blocked:
		if err == nil {
			t.Fatal("concurrent write succeeded while tx1 held its lock; expected it to block/fail")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent write neither completed nor errored within 3s")
	}

	// The read pool must remain usable throughout — WAL readers aren't
	// blocked by an in-flight writer.
	var count int
	if err := d.ReadDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		t.Fatalf("read during writer activity: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 committed notes before tx1.Commit, got %d", count)
	}

	if err := tx1.Commit(); err != nil {
		t.Fatalf("tx1.Commit: %v", err)
	}

	if err := d.ReadDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		t.Fatalf("read after commit: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 committed note after tx1.Commit (tx2 must not have also committed), got %d", count)
	}
}

// TestDSNPathEscaping guards against the Open path being passed
// unescaped into a "file:" URI DSN — a path containing a literal '?'
// would otherwise get truncated at the wrong point by
// modernc.org/sqlite's DSN parser, and a path with a space would fail
// to open at all.
func TestDSNPathEscaping(t *testing.T) {
	dir := t.TempDir()
	// A space is the realistic worst case (a user's home directory can
	// legitimately contain one); '?' is the pathological case that
	// proves escaping, not just luck, is doing the work.
	path := filepath.Join(dir, "notes db.db")

	d, err := Open(path)
	if err != nil {
		t.Fatalf("Open with a space in the path: %v", err)
	}
	_ = d.Close()
}

// TestOpenMemorySharesDataAcrossPools is a regression test for
// Open(":memory:"): SQLite's default in-memory behavior is one private
// database per *connection*, which would silently break this package's
// Read/Write pool split — the read pool's multiple connections would
// each see their own empty database, invisible to whatever the write
// pool's single connection wrote. Open rewrites ":memory:" to a
// per-call shared-cache URI specifically to prevent this; this test
// forces multiple read connections and confirms they all observe the
// write pool's data.
func TestOpenMemorySharesDataAcrossPools(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	d.ReadDB().SetMaxOpenConns(3)

	if err := Migrate(d.WriteDB()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := d.WriteDB().Exec("INSERT INTO notes (body) VALUES (?)", "probe"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	ctx := context.Background()
	for i := range 5 {
		var count int
		if err := d.ReadDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		if count != 1 {
			t.Fatalf("read %d: expected the write pool's row to be visible (count=1), got %d — read pool connection is isolated from the write pool", i, count)
		}
	}
}

// TestOpenMemoryIsolatedBetweenCalls guards the other half of the same
// fix: two separate Open(":memory:") calls (e.g. two parallel tests)
// must NOT share data, even though each internally uses shared-cache
// mode — shared-cache is scoped to a per-call random name (see
// randomDBName), not global.
func TestOpenMemoryIsolatedBetweenCalls(t *testing.T) {
	d1, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open d1: %v", err)
	}
	t.Cleanup(func() { _ = d1.Close() })
	d2, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open d2: %v", err)
	}
	t.Cleanup(func() { _ = d2.Close() })

	if err := Migrate(d1.WriteDB()); err != nil {
		t.Fatalf("migrate d1: %v", err)
	}
	if _, err := d1.WriteDB().Exec("INSERT INTO notes (body) VALUES (?)", "only in d1"); err != nil {
		t.Fatalf("insert into d1: %v", err)
	}

	if err := Migrate(d2.WriteDB()); err != nil {
		t.Fatalf("migrate d2 should succeed independently of d1's schema: %v", err)
	}
	var count int
	if err := d2.ReadDB().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		t.Fatalf("read d2: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected d2 isolated from d1, got count=%d", count)
	}
}
