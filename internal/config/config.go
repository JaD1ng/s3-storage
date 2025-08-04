package config

import (
	"encoding/json"
	"os"
)

// Config 系统配置
type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`

	Storage struct {
		DataDir string `json:"data_dir"`
		Nodes   []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"nodes"`
	} `json:"storage"`

	Database struct {
		Driver string `json:"driver"`
		DSN    string `json:"dsn"`
	} `json:"database"`

	Queue struct {
		Size int `json:"size"`
	} `json:"queue"`
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: struct {
			Port string `json:"port"`
			Host string `json:"host"`
		}{
			Port: "8080",
			Host: "localhost",
		},
		Storage: struct {
			DataDir string `json:"data_dir"`
			Nodes   []struct {
				ID   string `json:"id"`
				Path string `json:"path"`
			} `json:"nodes"`
		}{
			DataDir: "./data",
			Nodes: []struct {
				ID   string `json:"id"`
				Path string `json:"path"`
			}{
				{ID: "stg1", Path: "./data/stg1"},
				{ID: "stg2", Path: "./data/stg2"},
				{ID: "stg3", Path: "./data/stg3"},
			},
		},
		Database: struct {
			Driver string `json:"driver"`
			DSN    string `json:"dsn"`
		}{
			Driver: "sqlite3",
			DSN:    "./data/metadata.db",
		},
		Queue: struct {
			Size int `json:"size"`
		}{
			Size: 1000,
		},
	}
}

// Load 从文件加载配置
func Load(filename string) (*Config, error) {
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return Default(), nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Save 保存配置到文件
func (c *Config) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
