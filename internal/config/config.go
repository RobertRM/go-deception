package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listeners []Listener `yaml:"listeners"`
}

type Listener struct {
	Name     string  `yaml:"name"`
	Enabled  bool    `yaml:"enabled"`
	Protocol string  `yaml:"protocol"`
	Port     int     `yaml:"port"`
	Routes   []Route `yaml:"routes"`
}

type Route struct {
	Path     string   `yaml:"path"`
	Response Response `yaml:"response"`
}

type Response struct {
	StatusCode int               `yaml:"status_code"`
	Template   string            `yaml:"template"`
	Body       string            `yaml:"body"`
	Headers    map[string]string `yaml:"headers"`
	Vars       map[string]string `yaml:"vars"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config yaml file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse config yaml file: %w", err)
	}

	err = validateAndSetDefaults(&config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func validateAndSetDefaults(config *Config) error {
	if len(config.Listeners) == 0 {
		return fmt.Errorf("no listeners configured")
	}

	for _, listener := range config.Listeners {
		if listener.Protocol == "" {
			listener.Protocol = "http"
		}

		if listener.Port < 1 || listener.Port > 65535 {
			return fmt.Errorf("invalid port number: %d", listener.Port)
		}

		for _, route := range listener.Routes {
			if route.Response.StatusCode == 0 {
				route.Response.StatusCode = 200
			}

			if route.Response.Template == "" && route.Response.Body == "" {
				return fmt.Errorf("listener '%s' is missing body or template for route '%s'", listener.Name, route.Path)
			}
		}
	}
	return nil
}
