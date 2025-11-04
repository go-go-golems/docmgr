package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TTMPConfig defines repository-level configuration for docmgr
type TTMPConfig struct {
	Root     string `yaml:"root"`
	Defaults struct {
		Owners []string `yaml:"owners"`
		Intent string   `yaml:"intent"`
	} `yaml:"defaults"`
	FilenamePrefixPolicy string `yaml:"filenamePrefixPolicy"`
	Vocabulary           string `yaml:"vocabulary"`
}

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

// LoadTTMPConfig loads the nearest .ttmp.yaml, or returns nil if not found
func LoadTTMPConfig() (*TTMPConfig, error) {
	path, err := FindTTMPConfigPath()
	if err != nil {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var cfg TTMPConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	// Normalize relative paths in config to be relative to the config file directory
	if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(filepath.Dir(path), cfg.Root)
	}
	if cfg.Vocabulary != "" && !filepath.IsAbs(cfg.Vocabulary) {
		cfg.Vocabulary = filepath.Join(filepath.Dir(path), cfg.Vocabulary)
	}
	return &cfg, nil
}

// ResolveRoot applies config.Root if available and the provided root is the default ("ttmp")
func ResolveRoot(root string) string {
	// If a non-default root was explicitly provided and absolute, honor it
	if root != "ttmp" && root != "" {
		if filepath.IsAbs(root) {
			return root
		}
		// For relative non-default roots, anchor on current working directory
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, root)
		}
		return root
	}

	// Try to load config and resolve its root relative to the config file
	if cfgPath, err := FindTTMPConfigPath(); err == nil {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			var cfg TTMPConfig
			if yaml.Unmarshal(data, &cfg) == nil {
				if cfg.Root != "" {
					if filepath.IsAbs(cfg.Root) {
						return cfg.Root
					}
					return filepath.Join(filepath.Dir(cfgPath), cfg.Root)
				}
			}
		}
	}

	// Fallback: anchor default root on the git repository root if present
	if gitRoot, err := FindGitRoot(); err == nil && gitRoot != "" {
		return filepath.Join(gitRoot, "ttmp")
	}

	// Final fallback: anchor on current working directory
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "ttmp")
	}
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
			var cfg TTMPConfig
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
