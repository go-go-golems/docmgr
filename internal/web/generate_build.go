//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fatalf("find repo root: %v", err)
	}

	uiDir := filepath.Join(repoRoot, "ui")
	outDir := filepath.Join(repoRoot, "internal", "web", "embed", "public")

	if _, err := os.Stat(uiDir); err != nil {
		fatalf("ui directory not found at %s: %v", uiDir, err)
	}

	pnpmVersion := os.Getenv("WEB_PNPM_VERSION")
	if pnpmVersion == "" {
		pnpmVersion = "10.15.0"
	}

	builderImage := os.Getenv("WEB_BUILDER_IMAGE")
	if builderImage == "" {
		builderImage = "node:22"
	}

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fatalf("connect dagger: %v", err)
	}
	defer client.Close()

	if err := os.RemoveAll(outDir); err != nil {
		fatalf("remove %s: %v", outDir, err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", outDir, err)
	}

	webDir := client.Host().Directory(uiDir)
	ctr := client.Container().From(builderImage).
		WithWorkdir("/src").
		WithMountedDirectory("/src", webDir).
		WithEnvVariable("PNPM_HOME", "/pnpm")

	if pnpmCacheDir := os.Getenv("PNPM_CACHE_DIR"); pnpmCacheDir != "" {
		if err := os.MkdirAll(pnpmCacheDir, 0o755); err != nil {
			fatalf("mkdir %s: %v", pnpmCacheDir, err)
		}
		cacheDir := client.Host().Directory(pnpmCacheDir)
		ctr = ctr.WithMountedDirectory("/pnpm/store", cacheDir).
			WithEnvVariable("PNPM_STORE_DIR", "/pnpm/store")
	}

	if os.Getenv("WEB_BUILDER_IMAGE") == "" || !strings.Contains(builderImage, ":") {
		ctr = ctr.WithExec([]string{
			"sh", "-lc",
			fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate", pnpmVersion),
		})
	}

	ctr = ctr.
		WithExec([]string{"sh", "-lc", "pnpm --version"}).
		WithExec([]string{"sh", "-lc", "pnpm install --reporter=append-only"}).
		WithExec([]string{"sh", "-lc", "pnpm build"})

	dist := ctr.Directory("/src/dist/public")
	if _, err := dist.Export(ctx, outDir); err != nil {
		fatalf("export dist: %v", err)
	}
	log.Printf("exported web dist to %s", outDir)
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

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
