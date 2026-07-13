// Package db provides the template's SQLite access layer, implementing
// the seed spec's fixed, non-configurable concurrency contract (AD-1,
// AD-2): modernc.org/sqlite (pure Go, CGO-free) in WAL mode, with writes
// serialized through a single connection that always issues BEGIN
// IMMEDIATE.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite"
)

// DB holds two connection pools over the same SQLite file:
//
//   - Write is capped to exactly one connection (SetMaxOpenConns(1)) and
//     its DSN sets _txlock=immediate, so every write transaction begins
//     with BEGIN IMMEDIATE — it acquires SQLite's RESERVED lock up front
//     and fails fast (bounded by busy_timeout) against a concurrent
//     writer, instead of upgrading its lock mid-transaction and risking
//     SQLITE_BUSY deep inside application logic.
//   - Read allows multiple connections; WAL mode lets readers proceed
//     concurrently with the single writer without blocking on it.
//
// This split — not a single shared *sql.DB — is the contract itself: it
// is what makes "only one writer" true at the connection-pool level
// rather than relying on every caller to remember to serialize writes
// themselves. Callers write through BeginWrite, never by opening a write
// transaction against Read.
type DB struct {
	Read  *sql.DB
	Write *sql.DB
}

// Open opens path as a WAL-mode SQLite database and returns the
// Read/Write pool pair described on DB. path may be a plain filesystem
// path or ":memory:"-style special name understood by modernc.org/sqlite
// (the in-memory case is what Story #5's test helpers build on).
func Open(path string) (*DB, error) {
	read, err := sql.Open("sqlite", dsn(path, "deferred"))
	if err != nil {
		return nil, fmt.Errorf("db: open read pool: %w", err)
	}

	write, err := sql.Open("sqlite", dsn(path, "immediate"))
	if err != nil {
		_ = read.Close()
		return nil, fmt.Errorf("db: open write pool: %w", err)
	}
	write.SetMaxOpenConns(1)

	d := &DB{Read: read, Write: write}

	// journal_mode is a database-level property (not per-connection);
	// setting it once against the write connection is enough for it to
	// take effect durably for every subsequent connection, but every
	// connection still needs busy_timeout/foreign_keys applied for
	// itself — both pools' DSNs already carry those via _pragma, this
	// just forces WAL on before anything else runs.
	if _, err := d.Write.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}

	return d, nil
}

// Close closes both pools, returning every close error joined together.
func (d *DB) Close() error {
	return errors.Join(d.Write.Close(), d.Read.Close())
}

// BeginWrite starts a write transaction. Because Write's DSN sets
// _txlock=immediate, the underlying driver issues BEGIN IMMEDIATE for
// this call — see the DB doc comment for why that matters.
func (d *DB) BeginWrite(ctx context.Context) (*sql.Tx, error) {
	return d.Write.BeginTx(ctx, nil)
}

func dsn(path, txlock string) string {
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(ON)")
	q.Set("_txlock", txlock)
	return fmt.Sprintf("file:%s?%s", path, q.Encode())
}
