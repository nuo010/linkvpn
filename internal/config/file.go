package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadFromFile 将 YAML 配置文件合并进 cfg。
// 仅当文件存在且可解析时生效；错误留给调用方自行处理/忽略。
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// 允许 BOM
	s := strings.TrimPrefix(string(data), "\uFEFF")
	return yaml.Unmarshal([]byte(s), cfg)
}
