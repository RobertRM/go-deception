package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listeners []Listener `yaml:"listeners"`
}

type Listener struct {
	Name         string        `yaml:"name"`
	Enabled      bool          `yaml:"enabled"`
	Protocol     string        `yaml:"protocol"`
	Port         int           `yaml:"port"`
	Routes       []Route       `yaml:"routes"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
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

	ports := make(map[int]bool)

	for i := range config.Listeners {
		if !config.Listeners[i].Enabled {
			continue
		}
		// required
		if config.Listeners[i].Name == "" {
			return fmt.Errorf("listener name is required")
		}

		// validations
		if ok := ports[config.Listeners[i].Port]; ok {
			return fmt.Errorf("duplicate port number: %d", config.Listeners[i].Port)
		} else {
			ports[config.Listeners[i].Port] = true
		}

		if config.Listeners[i].Port < 1 || config.Listeners[i].Port > 65535 {
			return fmt.Errorf("invalid port number: %d", config.Listeners[i].Port)
		}

		// defaults
		if config.Listeners[i].Protocol == "" {
			config.Listeners[i].Protocol = "http"
		}

		if config.Listeners[i].ReadTimeout == 0 {
			config.Listeners[i].ReadTimeout = time.Second * 30 // default 30 seconds
		}

		if config.Listeners[i].WriteTimeout == 0 {
			config.Listeners[i].WriteTimeout = time.Second * 30 // default 30 seconds
		}

		routePaths := make(map[string]bool)
		for j := range config.Listeners[i].Routes {
			if ok := routePaths[config.Listeners[i].Routes[j].Path]; ok {
				return fmt.Errorf(
					"listener '%s' has duplicate route path: %s",
					config.Listeners[i].Name,
					config.Listeners[i].Routes[j].Path,
				)
			} else {
				routePaths[config.Listeners[i].Routes[j].Path] = true
			}

			if config.Listeners[i].Routes[j].Response.StatusCode == 0 {
				config.Listeners[i].Routes[j].Response.StatusCode = 200
			}

			if config.Listeners[i].Routes[j].Response.Template == "" && config.Listeners[i].Routes[j].Response.Body == "" {
				return fmt.Errorf(
					"listener '%s' is missing body or template for route '%s'",
					config.Listeners[i].Name,
					config.Listeners[i].Routes[j].Path,
				)
			}
		}
	}
	return nil
}
