package commands

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

// TTMPConfig defines repository-level configuration for docmgr
type TTMPConfig struct {
    Root string `yaml:"root"`
    Defaults struct {
        Owners []string `yaml:"owners"`
        Intent string   `yaml:"intent"`
    } `yaml:"defaults"`
    FilenamePrefixPolicy string            `yaml:"filenamePrefixPolicy"`
    DocTypeToggles       map[string]bool   `yaml:"docTypeToggles"`
}

// FindTTMPConfigPath walks up from cwd to find .ttmp.yaml
func FindTTMPConfigPath() (string, error) {
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
    return &cfg, nil
}

// ResolveRoot applies config.Root if available and the provided root is the default ("ttmp")
func ResolveRoot(root string) string {
    cfg, _ := LoadTTMPConfig()
    if cfg == nil || cfg.Root == "" {
        return root
    }
    if root == "ttmp" {
        return cfg.Root
    }
    return root
}


