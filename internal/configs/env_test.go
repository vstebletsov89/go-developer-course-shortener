package configs

import (
	"reflect"
	"testing"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
	}{
		{
			name:    "read config with defaults",
			want:    &Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080", FileStoragePath: "", DatabaseDsn: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
