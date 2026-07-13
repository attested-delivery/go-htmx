// Package db provides the template's SQLite access layer, implementing
// the seed spec's fixed, non-configurable concurrency contract (AD-1,
// AD-2): modernc.org/sqlite (pure Go, CGO-free) in WAL mode, with writes
// serialized through a single connection that always issues BEGIN
// IMMEDIATE.
package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite"
)

// DB holds two connection pools over the same SQLite file:
//
//   - the write pool is capped to exactly one connection
//     (SetMaxOpenConns(1)) and its DSN sets _txlock=immediate, so every
//     write transaction begins with BEGIN IMMEDIATE — it acquires
//     SQLite's RESERVED lock up front and fails fast (bounded by
//     busy_timeout) against a concurrent writer, instead of upgrading
//     its lock mid-transaction and risking SQLITE_BUSY deep inside
//     application logic.
//   - the read pool allows multiple connections; WAL mode lets readers
//     proceed concurrently with the single writer without blocking on
//     it.
//
// This split — not a single shared *sql.DB — is the contract itself: it
// is what makes "only one writer" true at the connection-pool level
// rather than relying on every caller to remember to serialize writes
// themselves. Both fields are unexported specifically so a caller can't
// reach around the contract (e.g. running a write against the read
// pool, where a lock upgrade mid-transaction can hit SQLITE_BUSY) —
// ReadDB/WriteDB/BeginWrite below are the only sanctioned access points.
type DB struct {
	read  *sql.DB
	write *sql.DB
}

// ReadDB returns the read pool, for building read-only query wrappers
// (e.g. sqlc.New(db.ReadDB())).
func (d *DB) ReadDB() *sql.DB { return d.read }

// WriteDB returns the write pool, for building query wrappers that issue
// single-statement writes in autocommit mode. Multi-statement writes
// needing atomicity should use BeginWrite instead.
func (d *DB) WriteDB() *sql.DB { return d.write }

// Open opens path as a WAL-mode SQLite database and returns the
// Read/Write pool pair described on DB. path is normally a plain
// filesystem path; passing ":memory:" opens an in-memory database
// instead (what Story #5's test helpers build on) — internally rewritten
// to a per-call shared-cache URI (see dsn) so the read pool's multiple
// connections and the write pool's single connection all see the same
// in-memory database, not each their own private, mutually invisible
// one (SQLite's default ":memory:" behavior is per-connection, which
// would silently break this package's Read/Write pool split).
func Open(path string) (*DB, error) {
	memory := path == ":memory:"
	if memory {
		name, err := randomDBName()
		if err != nil {
			return nil, fmt.Errorf("db: generate in-memory db name: %w", err)
		}
		path = name
	}

	read, err := sql.Open("sqlite", dsn(path, "deferred", memory))
	if err != nil {
		return nil, fmt.Errorf("db: open read pool: %w", err)
	}

	write, err := sql.Open("sqlite", dsn(path, "immediate", memory))
	if err != nil {
		_ = read.Close()
		return nil, fmt.Errorf("db: open write pool: %w", err)
	}
	write.SetMaxOpenConns(1)

	d := &DB{read: read, write: write}

	// journal_mode is a database-level property (not per-connection);
	// setting it once against the write connection is enough for it to
	// take effect durably for every subsequent connection, but every
	// connection still needs busy_timeout/foreign_keys applied for
	// itself — both pools' DSNs already carry those via _pragma, this
	// just forces WAL on before anything else runs. SQLite doesn't
	// support WAL for in-memory databases; PRAGMA journal_mode=WAL
	// against one is a documented no-op (it stays "memory"), not an
	// error, so this is safe to run unconditionally.
	if _, err := d.write.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}

	return d, nil
}

// randomDBName returns a random hex string suitable as a shared-cache
// in-memory database name (see dsn) — unique per Open() call so
// concurrent Open(":memory:") calls (e.g. parallel tests) never
// collide on the same in-memory database.
func randomDBName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "memdb_" + hex.EncodeToString(b), nil
}

// Close closes both pools, returning every close error joined together.
func (d *DB) Close() error {
	return errors.Join(d.write.Close(), d.read.Close())
}

// BeginWrite starts a write transaction. Because the write pool's DSN
// sets _txlock=immediate, the underlying driver issues BEGIN IMMEDIATE
// for this call — see the DB doc comment for why that matters.
func (d *DB) BeginWrite(ctx context.Context) (*sql.Tx, error) {
	return d.write.BeginTx(ctx, nil)
}

// dsn builds a SQLite "file:" URI DSN for path with the given _txlock
// mode. modernc.org/sqlite splits a raw DSN on the first literal '?' to
// separate path from query — but leaves a "file:"-prefixed DSN whole and
// hands it to SQLite's own URI parser, which expects the path segment
// RFC-3986-percent-encoded (see sqlite.org/uri.html). url.URL.EscapedPath
// does exactly that while preserving "/" as a path separator (unlike
// url.PathEscape, which would also escape "/" and break absolute
// paths) — without it, a path containing a literal '?' truncates at the
// wrong point and a path with spaces fails to open.
//
// When memory is true, path is a name (from randomDBName), not a
// filesystem path, and the DSN adds mode=memory&cache=shared: without
// cache=shared, each connection opening this same name would get its
// own private in-memory database, invisible to every other connection
// — silently breaking the read pool (multiple connections) seeing
// anything the write pool (a different connection) wrote.
func dsn(path, txlock string, memory bool) string {
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(ON)")
	q.Set("_txlock", txlock)
	if memory {
		q.Set("mode", "memory")
		q.Set("cache", "shared")
	}

	escapedPath := (&url.URL{Path: path}).EscapedPath()
	return fmt.Sprintf("file:%s?%s", escapedPath, q.Encode())
}
