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
    // Normalize relative paths in config to be relative to the config file directory
    if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
        cfg.Root = filepath.Join(filepath.Dir(path), cfg.Root)
    }
    return &cfg, nil
}

// ResolveRoot applies config.Root if available and the provided root is the default ("ttmp")
func ResolveRoot(root string) string {
    // If a non-default root was explicitly provided, honor it
    if root != "ttmp" && root != "" {
        return root
    }

    // Try to load config and resolve its root relative to the config file
    cfgPath, err := FindTTMPConfigPath()
    if err != nil {
        return root
    }
    data, err := os.ReadFile(cfgPath)
    if err != nil {
        return root
    }
    var cfg TTMPConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return root
    }
    if cfg.Root == "" {
        return root
    }
    if filepath.IsAbs(cfg.Root) {
        return cfg.Root
    }
    return filepath.Join(filepath.Dir(cfgPath), cfg.Root)
}


