package config

import (
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	ScannerDevice string
	ScannerTitle  string
	SaveFolder    string
}

// ConfigManager manages the application configuration
type ConfigManager struct {
	viper *viper.Viper
	path  string
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager() (*ConfigManager, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	configPath := path.Join(xdg.ConfigHome, "scanexpress")
	v.AddConfigPath(configPath)

	// Create config directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(configPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	configFilePath := path.Join(configPath, "config.yaml")

	// Try to read config, ignore error if file doesn't exist
	_ = v.ReadInConfig()

	return &ConfigManager{
		viper: v,
		path:  configFilePath,
	}, nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() Config {
	return Config{
		ScannerDevice: cm.viper.GetString("scanner.device"),
		ScannerTitle:  cm.viper.GetString("scanner.title"),
		SaveFolder:    cm.viper.GetString("save.folder"),
	}
}

// SaveConfig saves the configuration
func (cm *ConfigManager) SaveConfig(config Config) error {
	cm.viper.Set("scanner.device", config.ScannerDevice)
	cm.viper.Set("scanner.title", config.ScannerTitle)
	cm.viper.Set("save.folder", config.SaveFolder)

	return cm.viper.WriteConfigAs(cm.path)
}

// HasValidSavedConfig checks if we have valid saved configuration
func (cm *ConfigManager) HasValidSavedConfig() bool {
	config := cm.GetConfig()

	// Check if we have all necessary configuration
	if config.ScannerDevice == "" || config.ScannerTitle == "" || config.SaveFolder == "" {
		return false
	}

	// Check if the save folder exists
	if _, err := os.Stat(config.SaveFolder); os.IsNotExist(err) {
		return false
	}

	return true
}
