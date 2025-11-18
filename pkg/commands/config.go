// Package commands provides CLI command implementations for docmgr.
//
// Commands are built using the Glazed framework and implement the glazed.Command
// interface. Each command handles a specific docmgr operation such as creating tickets,
// adding documents, searching, or managing vocabulary.
//
// Configuration management functions in this package handle the resolution of
// documentation workspace roots using a multi-level fallback chain:
//  1. --root flag (explicit command-line argument)
//  2. .ttmp.yaml in current directory
//  3. .ttmp.yaml in parent directories (walk up tree)
//  4. DOCMGR_ROOT environment variable
//  5. Git repository root: <git-root>/ttmp
//  6. Default: ttmp in current directory
//
// See ResolveRoot() for the complete fallback logic.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// isVerbose returns true if verbose/debug output is enabled via DOCMGR_DEBUG env var.
func isVerbose() bool {
	return os.Getenv("DOCMGR_DEBUG") != ""
}

// verboseLog prints a message if verbose mode is enabled.
func verboseLog(format string, args ...interface{}) {
	if isVerbose() {
		fmt.Fprintf(os.Stderr, "[docmgr:debug] "+format+"\n", args...)
	}
}

// WorkspaceConfig defines repository-level configuration for docmgr.
//
// WorkspaceConfig holds settings for the documentation workspace, including the root
// directory path, default metadata values, vocabulary file location, and filename
// prefix policies.
//
// Example config file (.ttmp.yaml):
//
//	root: ~/projects/myapp/docs
//	defaults:
//	  owners: [alice, bob]
//	  intent: long-term
//	vocabulary: ~/projects/myapp/docs/vocabulary.yaml
//	filenamePrefixPolicy: numeric
//
// The root directory contains ticket workspaces organized by date:
//
//	root/
//	  2025/
//	    11/
//	      18/
//	        MEN-3475-add-feature/
//	          analysis/
//	          design-doc/
//	          playbook/
type WorkspaceConfig struct {
	Root     string `yaml:"root"`
	Defaults struct {
		Owners []string `yaml:"owners"`
		Intent string   `yaml:"intent"`
	} `yaml:"defaults"`
	FilenamePrefixPolicy string `yaml:"filenamePrefixPolicy"`
	Vocabulary           string `yaml:"vocabulary"`
}

// TTMPConfig is a deprecated alias for WorkspaceConfig.
// Use WorkspaceConfig instead.
//
// Deprecated: Use WorkspaceConfig instead. This alias is maintained for backward
// compatibility and will be removed in a future version.
type TTMPConfig = WorkspaceConfig

// FindTTMPConfigPath walks up from cwd to find .ttmp.yaml
func FindTTMPConfigPath() (string, error) {
	// 0) Explicit override via environment
	if env := os.Getenv("DOCMGR_CONFIG"); env != "" {
		if !filepath.IsAbs(env) {
			if cwd, err := os.Getwd(); err == nil {
				env = filepath.Join(cwd, env)
			}
		}
		if _, err := os.Stat(env); err == nil {
			return env, nil
		}
	}

	// 1) Walk up from CWD for .ttmp.yaml
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		cfg := filepath.Join(dir, ".ttmp.yaml")
		if _, err := os.Stat(cfg); err == nil {
			return cfg, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".ttmp.yaml not found")
}

// LoadWorkspaceConfig loads the nearest .ttmp.yaml configuration file.
// Returns nil if no config file is found (not an error condition).
//
// The function searches for .ttmp.yaml starting from the current directory
// and walking up the directory tree. Relative paths in the config are resolved
// relative to the config file's directory.
//
// If DOCMGR_DEBUG is set, the function logs the config resolution path.
// If a config file exists but is malformed, a warning is printed to stderr.
func LoadWorkspaceConfig() (*WorkspaceConfig, error) {
	path, err := FindTTMPConfigPath()
	if err != nil {
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			verboseLog("No .ttmp.yaml config file found (searched from %s)", cwd)
		} else {
			verboseLog("No .ttmp.yaml config file found")
		}
		return nil, nil
	}
	verboseLog("Found config file: %s", path)
	
	data, err := os.ReadFile(path)
	if err != nil {
		// File exists but can't be read - this is an error
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	
	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Config file exists but is malformed - warn but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file %s: %v\n", path, err)
		fmt.Fprintf(os.Stderr, "Continuing with default configuration. Fix the config file to resolve this warning.\n")
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	
	verboseLog("Loaded config from %s: root=%s, vocabulary=%s", path, cfg.Root, cfg.Vocabulary)
	
	// Normalize relative paths in config to be relative to the config file directory
	if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(filepath.Dir(path), cfg.Root)
		verboseLog("Resolved relative root path: %s", cfg.Root)
	}
	if cfg.Vocabulary != "" && !filepath.IsAbs(cfg.Vocabulary) {
		cfg.Vocabulary = filepath.Join(filepath.Dir(path), cfg.Vocabulary)
		verboseLog("Resolved relative vocabulary path: %s", cfg.Vocabulary)
	}
	return &cfg, nil
}

// LoadTTMPConfig is a deprecated alias for LoadWorkspaceConfig.
//
// Deprecated: Use LoadWorkspaceConfig instead. This function is maintained for
// backward compatibility and will be removed in a future version.
func LoadTTMPConfig() (*TTMPConfig, error) {
	// TTMPConfig is a type alias for WorkspaceConfig, so the types are identical
	cfg, err := LoadWorkspaceConfig()
	if cfg == nil {
		return nil, err
	}
	// Type alias means *TTMPConfig and *WorkspaceConfig are the same type
	var result *TTMPConfig = (*TTMPConfig)(cfg)
	return result, err
}

// ResolveRoot applies config.Root if available and the provided root is the default ("ttmp").
//
// The function uses a multi-level fallback chain to resolve the workspace root:
//  1. --root flag (if non-default and provided)
//  2. .ttmp.yaml config file (if found)
//  3. Git repository root: <git-root>/ttmp
//  4. Current working directory: <cwd>/ttmp
//
// If DOCMGR_DEBUG is set, the function logs each step of the resolution process.
func ResolveRoot(root string) string {
	verboseLog("Resolving workspace root (provided: %q)", root)
	
	// If a non-default root was explicitly provided and absolute, honor it
	if root != "ttmp" && root != "" {
		if filepath.IsAbs(root) {
			verboseLog("Using explicit absolute root: %s", root)
			return root
		}
		// For relative non-default roots, anchor on current working directory
		if cwd, err := os.Getwd(); err == nil {
			resolved := filepath.Join(cwd, root)
			verboseLog("Using explicit relative root (resolved from %s): %s", cwd, resolved)
			return resolved
		}
		verboseLog("Using explicit root as-is: %s", root)
		return root
	}

	// Try to load config and resolve its root relative to the config file
	verboseLog("Checking for .ttmp.yaml config file...")
	if cfgPath, err := FindTTMPConfigPath(); err == nil {
		verboseLog("Found config file: %s", cfgPath)
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			var cfg WorkspaceConfig
			if yaml.Unmarshal(data, &cfg) == nil {
				if cfg.Root != "" {
					var resolved string
					if filepath.IsAbs(cfg.Root) {
						resolved = cfg.Root
					} else {
						resolved = filepath.Join(filepath.Dir(cfgPath), cfg.Root)
					}
					verboseLog("Using root from config file: %s (resolved: %s)", cfg.Root, resolved)
					return resolved
				}
				verboseLog("Config file found but no root specified, continuing fallback chain")
			} else {
				// Config file exists but is malformed
				fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file %s: %v\n", cfgPath, err)
				fmt.Fprintf(os.Stderr, "Continuing with fallback resolution. Fix the config file to resolve this warning.\n")
			}
		} else {
			verboseLog("Config file exists but cannot be read: %v", err)
		}
	} else {
		verboseLog("No config file found: %v", err)
	}

	// Fallback: anchor default root on the git repository root if present
	verboseLog("Checking for git repository root...")
	if gitRoot, err := FindGitRoot(); err == nil && gitRoot != "" {
		resolved := filepath.Join(gitRoot, "ttmp")
		verboseLog("Using git repository root: %s (resolved: %s)", gitRoot, resolved)
		return resolved
	}
	verboseLog("No git repository root found")

	// Final fallback: anchor on current working directory
	verboseLog("Using current working directory as fallback...")
	if cwd, err := os.Getwd(); err == nil {
		resolved := filepath.Join(cwd, "ttmp")
		verboseLog("Final fallback: %s (resolved: %s)", cwd, resolved)
		return resolved
	}
	verboseLog("Using provided root as final fallback: %s", root)
	return root
}

// FindGitRoot walks up from the current working directory to find the nearest .git directory
func FindGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		if fi, err := os.Stat(gitPath); err == nil {
			if fi.IsDir() {
				return dir, nil
			}
			// .git is a file; parse gitdir:
			if b, err := os.ReadFile(gitPath); err == nil {
				line := strings.TrimSpace(string(b))
				lower := strings.ToLower(line)
				if strings.HasPrefix(lower, "gitdir:") {
					gd := strings.TrimSpace(line[len("gitdir:"):])
					if !filepath.IsAbs(gd) {
						gd = filepath.Join(dir, gd)
					}
					if _, err := os.Stat(gd); err == nil {
						return dir, nil
					}
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".git directory not found")
}

// FindRepositoryRoot returns the best-effort repository root by preferring the git root,
// then falling back to markers like go.mod or doc/ as anchors.
func FindRepositoryRoot() (string, error) {
	if gr, err := FindGitRoot(); err == nil && gr != "" {
		return gr, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "doc")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// As a last resort, use current working directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd, nil
	}
	return "", fmt.Errorf("could not determine repository root")
}

// DetectMultipleTTMPRoots walks up from CWD and records directories containing a 'ttmp' folder
func DetectMultipleTTMPRoots() ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var roots []string
	seen := map[string]bool{}
	for {
		candidate := filepath.Join(dir, "ttmp")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			if !seen[candidate] {
				roots = append(roots, candidate)
				seen[candidate] = true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots, nil
}

// ResolveVocabularyPath returns the absolute path to the vocabulary file.
// Priority:
// 1) If .ttmp.yaml defines 'vocabulary', use it (relative to the config file if not absolute)
// 2) Else, use '<root>/vocabulary.yaml' where root comes from .ttmp.yaml (default 'ttmp' relative to config)
// 3) Else, search upwards for 'ttmp/vocabulary.yaml'
// 4) Finally, as a legacy fallback, search for 'doc/vocabulary.yaml'
func ResolveVocabularyPath() (string, error) {
	// Use config if present
	cfgPath, err := FindTTMPConfigPath()
	if err == nil {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			var cfg WorkspaceConfig
			if yaml.Unmarshal(data, &cfg) == nil {
				// If vocabulary explicitly set
				if cfg.Vocabulary != "" {
					if filepath.IsAbs(cfg.Vocabulary) {
						return cfg.Vocabulary, nil
					}
					return filepath.Join(filepath.Dir(cfgPath), cfg.Vocabulary), nil
				}
				// Build from root default
				rootPath := cfg.Root
				if rootPath == "" {
					rootPath = "ttmp"
				}
				if !filepath.IsAbs(rootPath) {
					rootPath = filepath.Join(filepath.Dir(cfgPath), rootPath)
				}
				return filepath.Join(rootPath, "vocabulary.yaml"), nil
			}
		}
	}

	// Search upwards for ttmp/vocabulary.yaml
	dir, err := os.Getwd()
	if err == nil {
		for {
			p := filepath.Join(dir, "ttmp", "vocabulary.yaml")
			if _, err2 := os.Stat(p); err2 == nil {
				return p, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// Legacy fallback: search for doc/vocabulary.yaml
	dir, err = os.Getwd()
	if err == nil {
		for {
			p := filepath.Join(dir, "doc", "vocabulary.yaml")
			if _, err2 := os.Stat(p); err2 == nil {
				return p, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "", fmt.Errorf("vocabulary.yaml not found")
}
