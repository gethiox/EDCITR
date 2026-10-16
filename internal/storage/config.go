package storage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gethiox/EDCITR/internal/config"
)

type Config struct {
	root string
}

func GetConfig(appName string) (*Config, error) {
	if appName == "" {
		return nil, errors.New("appName cannot be empty")
	}

	configRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config directory failed: %w", err)
	}

	if configRoot == "" {
		return nil, errors.New("user config directory is an empty string")
	}

	appConfigDirectory := filepath.Join(configRoot, appName)

	err = os.MkdirAll(appConfigDirectory, 0755)
	if err != nil {
		return nil, fmt.Errorf("create config directory failed: %w", err)
	}

	return &Config{root: appConfigDirectory}, nil
}

func (c *Config) Root() string {
	return c.root
}

const CONFIG_FILENAME = "config.json"

func (c *Config) SaveConfig(edcitrConfig *config.EDCITRConfig) error {
	data, err := json.Marshal(edcitrConfig, jsontext.Multiline(true))
	if err != nil {
		return fmt.Errorf("marshal edcitr config failed: %w", err)
	}

	err = os.WriteFile(filepath.Join(c.root, CONFIG_FILENAME), data, 0644)
	if err != nil {
		return fmt.Errorf("save edcitr config failed: %w", err)
	}
	return nil
}

func (c *Config) RemoveConfig() error {
	err := os.Remove(filepath.Join(c.root, CONFIG_FILENAME))
	if err != nil {
		return fmt.Errorf("remove edcitr config failed: %w", err)
	}
	return nil
}

func (c *Config) LoadConfig() (*config.EDCITRConfig, error) {
	data, err := os.ReadFile(filepath.Join(c.root, CONFIG_FILENAME))
	if err != nil {
		return nil, fmt.Errorf("load config failed: %w", err)
	}

	edcitrConfig := config.GetDefaultConfig()

	err = json.Unmarshal(data, edcitrConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshal edcitr config failed: %w", err)
	}

	return edcitrConfig, nil
}
