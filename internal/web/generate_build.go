//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fatalf("find repo root: %v", err)
	}

	uiDir := filepath.Join(repoRoot, "ui")
	outDir := filepath.Join(repoRoot, "internal", "web", "embed", "public")
	builtDir := filepath.Join(uiDir, "dist", "public")

	if err := run(repoRoot, "pnpm", "-C", uiDir, "build"); err != nil {
		fatalf("ui build: %v", err)
	}

	if err := os.RemoveAll(outDir); err != nil {
		fatalf("remove %s: %v", outDir, err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", outDir, err)
	}

	if err := copyDir(builtDir, outDir); err != nil {
		fatalf("copy %s -> %s: %v", builtDir, outDir, err)
	}
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			return "", fmt.Errorf("go.mod not found walking up from %s", os.Getwd)
		}
		wd = next
	}
}

func run(repoRoot, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func copyDir(src, dst string) error {
	return fs.WalkDir(os.DirFS(src), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}

		srcPath := filepath.Join(src, p)
		dstPath := filepath.Join(dst, p)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		in, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = out.Close() }()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}

		// Ensure readability in repo checkouts.
		if err := os.Chmod(dstPath, 0o644); err != nil {
			return err
		}
		return nil
	})
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
