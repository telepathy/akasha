package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	App      AppConfig      `yaml:"app"`
	Admin    AdminConfig    `yaml:"admin"`
}

type AdminConfig struct {
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.Name)
}

type AppConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (a AppConfig) Addr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func Load(path string) (*Config, error) {
	// Try to load from file first
	var cfg Config
	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Override with environment variables if set
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DATABASE_USERNAME"); user != "" {
		cfg.Database.Username = user
	}
	if pwd := os.Getenv("DATABASE_PASSWORD"); pwd != "" {
		cfg.Database.Password = pwd
	}
	if db := os.Getenv("DATABASE_NAME"); db != "" {
		cfg.Database.Name = db
	}
	if appHost := os.Getenv("APP_HOST"); appHost != "" {
		cfg.App.Host = appHost
	}
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		fmt.Sscanf(appPort, "%d", &cfg.App.Port)
	}
	if adminPwd := os.Getenv("ADMIN_PASSWORD"); adminPwd != "" {
		cfg.Admin.Password = adminPwd
	}

	return &cfg, nil
}