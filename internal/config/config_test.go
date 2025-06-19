package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
								},
							},
						},
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

func TestValidateAndSetDefaults(t *testing.T) {

	tc := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "no listeners",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "duplicate port",
			config: &Config{
				Listeners: []Listener{
					{Name: "http-listener", Enabled: true, Protocol: "http", Port: 8080},
					{Name: "http-listener", Enabled: true, Protocol: "http", Port: 8080},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate path",
			config: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     8080,
						Routes: []Route{
							{Path: "/duplicate"},
							{Path: "/duplicate"},
						},
					},
				},
			},
			wantErr: true,
		},
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
		{
			name: "invalid port",
			config: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     0,
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
			wantErr: true,
		},
		{
			name: "missing body or template",
			config: &Config{
				Listeners: []Listener{
					{
						Name:     "http-listener",
						Enabled:  true,
						Protocol: "http",
						Port:     0,
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
			wantErr: true,
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
