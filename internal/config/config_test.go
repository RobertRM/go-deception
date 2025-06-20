package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// make a temp directory for the test yaml files
	tmpDir, err := os.MkdirTemp("", "config")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// setup valid yaml
	validYAML := `
listeners:
  - name: "http-listener"
    enabled: true
    protocol: "http"
    port: 8080
    read_timeout: 40s
    write_timeout: 60s
    routes:
      - path: "/test"
        response:
          status_code: 200
          body: "Hello, World!"
`
	validFilePath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validFilePath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("failed to write valid yaml file: %v", err)
	}

	// setup unparsable yaml
	unparsableYAML := `
listeners:
  - name: "http-listener"
    enabled: true
    protocol: "http"
    port: 8080
    routes:
      - path: "/test"
        response:
          status_code: 200
          body: "Hello, World!"
code
	`
	unparsableFilePath := filepath.Join(tmpDir, "unparsable.yaml")
	if err := os.WriteFile(unparsableFilePath, []byte(unparsableYAML), 0644); err != nil {
		t.Fatalf("failed to write unparsable yaml file: %v", err)
	}

	// setup inavlid yaml, fails internal validation
	invalidYAML := `
listeners:
  - name: "http-listener"
    enabled: true
    protocol: "http"
    port: 8080
    routes:
      - path: "/test"
        response:
          status_code: 200
`
	invalidFilePath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidFilePath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write invalid yaml file: %v", err)
	}

	// test cases
	tc := []struct {
		name    string
		path    string
		wantErr bool
		want    *Config
	}{
		{
			name:    "valid yaml",
			path:    validFilePath,
			wantErr: false,
			want: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     8080,
						Routes: []Route{
							{
								Path: "/test",
								Response: Response{
									StatusCode: 200,
									Body:       "Hello, World!",
									Template:   "",
								},
							},
						},
						ReadTimeout:  40 * time.Second,
						WriteTimeout: 60 * time.Second,
					},
				},
			},
		},
		{
			name:    "unparsable yaml",
			path:    unparsableFilePath,
			wantErr: true,
			want:    nil,
		},
		{
			name:    "invalid yaml",
			path:    invalidFilePath,
			wantErr: true,
			want:    nil,
		}, {
			name:    "missing file",
			path:    "nofile.file",
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndSetDefaults_Required(t *testing.T) {
	// All test cases should throw errors
	tc := []struct {
		name   string
		config *Config
	}{
		{
			name:   "at least one listener is required",
			config: &Config{},
		},
		{
			name:   "listener name is required",
			config: &Config{Listeners: []Listener{{Enabled: true}}},
		},
		{
			name: "route body or template is required",
			config: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     1,
						Routes: []Route{
							{
								Path: "/test",
								Response: Response{
									StatusCode: 200,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndSetDefaults(tt.config)
			if err == nil {
				t.Errorf("validateAndSetDefaults() - missing error")
			}
		})
	}
}

func TestValidateAndSetDefaults_Validation(t *testing.T) {
	tc := []struct {
		name   string
		config *Config
	}{
		{
			name: "duplicate ports",
			config: &Config{
				Listeners: []Listener{
					{Name: "http-listener", Enabled: true, Port: 8080},
					{Name: "http-listener", Enabled: true, Port: 8080},
				},
			},
		},
		{
			name: "duplicate route paths",
			config: &Config{Listeners: []Listener{
				{Name: "http-listener", Enabled: true, Port: 8080, Routes: []Route{
					{Path: "/duplicate"},
					{Path: "/duplicate"},
				}},
			}},
		},
		{
			name: "port too low",
			config: &Config{Listeners: []Listener{
				{Name: "http-listener", Enabled: true, Port: 0}}},
		},
		{
			name: "port too high",
			config: &Config{Listeners: []Listener{
				{Name: "http-listener", Enabled: true, Port: 65536}}},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndSetDefaults(tt.config)
			if err == nil {
				t.Errorf("validateAndSetDefaults() - should cause error")
			}
		})
	}
}

func TestValidateAndSetDefaults_Defaults(t *testing.T) {
	config := &Config{
		Listeners: []Listener{
			{
				Name:     "http-listener",
				Enabled:  true,
				Protocol: "",
				Port:     8080,
				Routes: []Route{
					{
						Path: "/",
						Response: Response{
							StatusCode: 0,
							Template:   "",
							Body:       "Hello, World!",
							Headers:    map[string]string{},
							Vars:       map[string]string{},
						},
					},
				},
				ReadTimeout:  0,
				WriteTimeout: 0,
			},
		},
	}

	want := &Config{
		Listeners: []Listener{
			{
				Name:     "http-listener",
				Enabled:  true,
				Protocol: "http", // default
				Port:     8080,
				Routes: []Route{
					{
						Path: "/",
						Response: Response{
							StatusCode: 200, // default
							Template:   "",
							Body:       "Hello, World!",
							Headers:    map[string]string{},
							Vars:       map[string]string{},
						},
					},
				},
				ReadTimeout:  time.Second * 30, // default
				WriteTimeout: time.Second * 30, // default
			},
		},
	}

	err := validateAndSetDefaults(config)
	if err != nil {
		t.Errorf("validateAndSetDefaults - should not cause error %v", err)
	}

	if !reflect.DeepEqual(config, want) {
		t.Errorf("validateAndSetDefaults - defaults don't match got %v want %v", config, want)
	}
}
func TestValidateAndSetDefaults(t *testing.T) {

	tc := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     8080,
						Routes: []Route{
							{
								Path: "/test",
								Response: Response{
									StatusCode: 200,
									Body:       "Hello, World!",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tc {

		t.Run(tt.name, func(t *testing.T) {
			err := validateAndSetDefaults(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndSetDefaults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
