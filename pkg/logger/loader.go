package logger

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromYAML 从 YAML 文件加载配置
func LoadFromYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// InitFromYAML 从 YAML 文件初始化日志
func InitFromYAML(path string) error {
	cfg, err := LoadFromYAML(path)
	if err != nil {
		return err
	}
	cfg.ApplyEnv() // 允许环境变量覆盖
	InitWithConfig(cfg)
	return nil
}
