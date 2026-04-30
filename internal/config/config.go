package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database     DatabaseConfig `yaml:"database"`
	App          AppConfig      `yaml:"app"`
	Admin        AdminConfig    `yaml:"admin"`
	Auth         AuthConfig     `yaml:"auth"`
	APIKey       string         `yaml:"apiKey"`
	ExternalHost string         `yaml:"externalHost"`
}

type AdminConfig struct {
	Password string `yaml:"password"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwtSecret"`
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
	var cfg Config
	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	if err := envOverride(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func envOverride(cfg *Config) error {
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid DATABASE_PORT: %s", v)
		}
		cfg.Database.Port = p
	}
	if v := os.Getenv("DATABASE_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("APP_HOST"); v != "" {
		cfg.App.Host = v
	}
	if v := os.Getenv("APP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid APP_PORT: %s", v)
		}
		cfg.App.Port = p
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		cfg.Admin.Password = v
	}
	if v := os.Getenv("API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("EXTERNAL_HOST"); v != "" {
		cfg.ExternalHost = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	return nil
}