package config

import (
	"testing"
)

func TestProduction(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantErr   bool
		wantValue bool
	}{
		{name: "default is false", envValue: "", wantErr: false, wantValue: false},
		{name: "set to true", envValue: "true", wantErr: false, wantValue: true},
		{name: "set to false", envValue: "false", wantErr: false, wantValue: false},
		{name: "bad value", envValue: "blah", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("GAMES_PRODUCTION", tt.envValue)
			}
			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cfg.Production != tt.wantValue {
				t.Errorf("Production = %v, want %v", cfg.Production, tt.wantValue)
			}
		})
	}
}
