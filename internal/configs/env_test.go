package configs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name         string
		want         *Config
		jsonConfig   map[string]interface{}
		deleteConfig bool
		wantErr      bool
	}{
		{
			name:         "read config with defaults",
			want:         &Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080", FileStoragePath: "", DatabaseDsn: "", EnableHTTPS: false, TrustedSubnet: ""},
			jsonConfig:   nil,
			deleteConfig: false,
			wantErr:      false,
		},
		{
			name: "read config from json config",
			want: &Config{ServerAddress: "host:9090", BaseURL: "https://baseurl", FileStoragePath: "/path/to/file.db", DatabaseDsn: "databaseConnectionString", EnableHTTPS: true, TrustedSubnet: "192.168.0.15/24"},
			jsonConfig: map[string]interface{}{
				"server_address":    "host:9090",
				"base_url":          "https://baseurl",
				"file_storage_path": "/path/to/file.db",
				"database_dsn":      "databaseConnectionString",
				"enable_https":      true,
				"trusted_subnet":    "192.168.0.15/24",
			},
			deleteConfig: false,
			wantErr:      false,
		},
		{
			name:         "read config from empty json config",
			want:         nil,
			jsonConfig:   map[string]interface{}{},
			deleteConfig: false,
			wantErr:      true,
		},
		{
			name:         "read not existing config file",
			want:         nil,
			jsonConfig:   map[string]interface{}{},
			deleteConfig: true,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.jsonConfig != nil {
				dir, err := os.Getwd()
				assert.NoError(t, err)

				temp, err := os.MkdirTemp(dir, "test")
				assert.NoError(t, err)

				// create json config
				file := filepath.Join(temp, "config.json")
				cfg, err := json.MarshalIndent(tt.jsonConfig, "", "   ")
				assert.NoError(t, err)

				err = os.WriteFile(file, cfg, 0666)
				assert.NoError(t, err)

				t.Setenv("CONFIG", file)

				if tt.deleteConfig {
					err := os.RemoveAll(temp)
					assert.NoError(t, err)
				} else {
					defer func() {
						err := os.RemoveAll(temp)
						assert.NoError(t, err)
					}()
				}
			}

			got, err := ReadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				testConfig := &Config{ServerAddress: got.ServerAddress, BaseURL: got.BaseURL, FileStoragePath: got.FileStoragePath, DatabaseDsn: got.DatabaseDsn, EnableHTTPS: got.EnableHTTPS,
					TrustedSubnet: got.TrustedSubnet}
				if !reflect.DeepEqual(testConfig, tt.want) {
					t.Errorf("ReadConfig() got = %v, want %v", testConfig, tt.want)
				}
			}
		})
	}
}
