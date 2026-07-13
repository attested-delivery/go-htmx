// Command init rewrites this template's own identity into a new
// project's, in one deterministic, idempotent pass (Story #6, Task
// #25). Run via `just init name=<name> module=<module>` — never
// invoked by the built application itself, and not part of `just
// build`'s single-binary output.
//
// Identity audit (Task #26's source of truth — every category of
// identity-bearing string in the tree, and how this tool derives the
// "old" value it replaces rather than hardcoding "go-htmx"/"attested-
// delivery" literals, so re-running after a prior init is still
// correct, not just a no-op when inputs are unchanged):
//
//   - Module path (go.mod's module line, every Go/.templ import
//     statement, .golangci.yml's depguard allow-list): read live from
//     go.mod via currentModule, not hardcoded.
//   - GitHub owner/repo slug (SECURITY.md, .github/ISSUE_TEMPLATE/
//     config.yml, .config/gdlc/config.yml — anywhere the bare
//     "owner/repo" form appears without a "github.com/" prefix):
//     derived from the module path via repoSlug, only when both the
//     old and new module are github.com paths.
//   - Binary/cmd name (the cmd/ subdirectory, justfile's build output
//     path, README title, the templ page title, the default SQLite
//     filename, narrative comments mentioning the app by name, and
//     internal/notes/testdata/*.golden's snapshot of the rendered page
//     title — golden fixtures encode this app's actual rendered output,
//     so they need the same rewrite or `just test` fails right after
//     init on a diff that's really just the identity rename working
//     correctly): read live from the current cmd/ directory's name via
//     currentCmdName.
//   - Env var prefix (GO_HTMX_ADDR/ENV/DB_PATH in config.go and
//     litestream.yml): derived from the binary name via envPrefix
//     (upper-case, hyphens to underscores) — not a separate input,
//     since the template's own convention ties it to the binary name.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/init <name> <module>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err)
		os.Exit(1)
	}
}

func run(newName, newModule string) error {
	if newName == "" || newModule == "" {
		return fmt.Errorf("name and module are both required")
	}
	if strings.ContainsAny(newName, " \t/\\") {
		return fmt.Errorf("name must not contain whitespace or path separators, got %q", newName)
	}

	oldModule, err := currentModule()
	if err != nil {
		return err
	}
	oldName, err := currentCmdName()
	if err != nil {
		return err
	}

	if oldModule == newModule && oldName == newName {
		fmt.Println("init: name and module already match the current tree — nothing to do")
		return nil
	}

	type replacement struct{ old, new string }
	replacements := []replacement{{oldModule, newModule}}

	if oldSlug, oldOK := repoSlug(oldModule); oldOK {
		if newSlug, newOK := repoSlug(newModule); newOK {
			replacements = append(replacements, replacement{oldSlug, newSlug})
		}
	}
	replacements = append(replacements,
		replacement{oldName, newName},
		replacement{envPrefix(oldName), envPrefix(newName)},
	)

	files, err := targetFiles()
	if err != nil {
		return err
	}

	changed := 0
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		out := data
		for _, r := range replacements {
			if r.old == "" || r.old == r.new {
				continue
			}
			out = bytes.ReplaceAll(out, []byte(r.old), []byte(r.new))
		}
		if !bytes.Equal(out, data) {
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("stat %s: %w", path, err)
			}
			if err := os.WriteFile(path, out, info.Mode()); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			changed++
			fmt.Println("rewrote", path)
		}
	}

	oldCmdDir := filepath.Join("cmd", oldName)
	newCmdDir := filepath.Join("cmd", newName)
	if oldCmdDir != newCmdDir {
		if _, err := os.Stat(oldCmdDir); err == nil {
			if err := os.Rename(oldCmdDir, newCmdDir); err != nil {
				return fmt.Errorf("rename %s -> %s: %w", oldCmdDir, newCmdDir, err)
			}
			fmt.Printf("renamed %s -> %s\n", oldCmdDir, newCmdDir)
		}
	}

	// Belt-and-suspenders: the string replace above already rewrote
	// go.mod's module line, but `go mod edit -module` is the
	// toolchain-native way to do it and keeps go.mod's formatting
	// canonical — no extra tool needed beyond what the template
	// already requires.
	cmd := exec.Command("go", "mod", "edit", "-module", newModule)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod edit -module: %w", err)
	}

	fmt.Printf("init complete: %d file(s) rewritten, module %s -> %s, name %s -> %s\n",
		changed, oldModule, newModule, oldName, newName)
	return nil
}

func currentModule() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(after), nil
		}
	}
	return "", fmt.Errorf("go.mod has no module line")
}

func currentCmdName() (string, error) {
	entries, err := os.ReadDir("cmd")
	if err != nil {
		return "", fmt.Errorf("read cmd/: %w", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) != 1 {
		return "", fmt.Errorf("expected exactly one directory under cmd/, found %d: %v", len(dirs), dirs)
	}
	return dirs[0], nil
}

func envPrefix(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

// repoSlug extracts "owner/repo" from a "github.com/owner/repo" module
// path. ok is false for a non-GitHub module path, in which case the
// caller skips the bare-slug replacement entirely rather than guessing.
func repoSlug(module string) (slug string, ok bool) {
	const prefix = "github.com/"
	if !strings.HasPrefix(module, prefix) {
		return "", false
	}
	return strings.TrimPrefix(module, prefix), true
}

// targetFiles returns every file init scans for identity strings — an
// explicit extension/name allow-list, not a blind repo walk, so
// vendored assets (internal/web/assets/static/js/*.min.js), go.sum,
// binary files, and .git internals are never touched.
func targetFiles() ([]string, error) {
	allowExt := map[string]bool{".go": true, ".templ": true, ".md": true, ".yml": true, ".yaml": true, ".golden": true}
	allowName := map[string]bool{"justfile": true, "go.mod": true}
	skipDirs := map[string]bool{".git": true, "bin": true, "sqlc": true}

	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if allowExt[filepath.Ext(path)] || allowName[filepath.Base(path)] {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
