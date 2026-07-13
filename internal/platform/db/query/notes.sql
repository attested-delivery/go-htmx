-- name: ListNotes :many
SELECT id, body, created_at FROM notes ORDER BY id DESC;

-- name: GetNote :one
SELECT id, body, created_at FROM notes WHERE id = ?;

-- name: CreateNote :one
INSERT INTO notes (body) VALUES (?)
RETURNING id, body, created_at;
