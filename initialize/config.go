package initialize

import (
	"bookadmin/config"
	"bookadmin/global"
	"os"
)

func Config() (*config.Config, error) {
	if global.GVA_CONFIG != nil {
		return global.GVA_CONFIG, nil
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	global.GVA_CONFIG = cfg
	return cfg, nil
}
