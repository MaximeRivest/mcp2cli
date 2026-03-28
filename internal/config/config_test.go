package config

import (
	"path/filepath"
	"testing"
)

func TestRepositoryLoadMergedLocalOverridesGlobal(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepositoryWithPaths(Paths{
		GlobalConfig: filepath.Join(tempDir, "global.yaml"),
		LocalConfig:  filepath.Join(tempDir, ".mcp2cli.yaml"),
		ExposeBinDir: filepath.Join(tempDir, "bin"),
	})

	if err := repo.UpsertServer(SourceGlobal, "weather", &Server{Command: "global-weather"}); err != nil {
		t.Fatalf("UpsertServer(global): %v", err)
	}
	if err := repo.UpsertServer(SourceLocal, "weather", &Server{Command: "local-weather"}); err != nil {
		t.Fatalf("UpsertServer(local): %v", err)
	}

	merged, err := repo.LoadMerged()
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}

	server := merged.Servers["weather"]
	if server == nil {
		t.Fatal("weather server missing from merged config")
	}
	if server.Command != "local-weather" {
		t.Fatalf("merged command = %q, want %q", server.Command, "local-weather")
	}
	if server.Source != SourceLocal {
		t.Fatalf("merged source = %q, want %q", server.Source, SourceLocal)
	}
}

func TestExposeUsesFullCommandName(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepositoryWithPaths(Paths{
		GlobalConfig: filepath.Join(tempDir, "global.yaml"),
		LocalConfig:  filepath.Join(tempDir, ".mcp2cli.yaml"),
		ExposeBinDir: filepath.Join(tempDir, "bin"),
	})

	if err := repo.UpsertServer(SourceGlobal, "weather", &Server{Command: "weather-cmd"}); err != nil {
		t.Fatalf("UpsertServer: %v", err)
	}
	if err := repo.AddExpose(SourceGlobal, "weather", "wea"); err != nil {
		t.Fatalf("AddExpose: %v", err)
	}

	server, err := repo.ResolveExposedCommand("wea")
	if err != nil {
		t.Fatalf("ResolveExposedCommand: %v", err)
	}
	if server.Name != "weather" {
		t.Fatalf("ResolveExposedCommand returned server %q, want %q", server.Name, "weather")
	}

	defaultName, err := DefaultExposeName("weather")
	if err != nil {
		t.Fatalf("DefaultExposeName: %v", err)
	}
	if defaultName != "mcp-weather" {
		t.Fatalf("DefaultExposeName = %q, want %q", defaultName, "mcp-weather")
	}
}

func TestNormalizeCommandName(t *testing.T) {
	normalized, err := NormalizeCommandName(" Weather_Forecast ")
	if err != nil {
		t.Fatalf("NormalizeCommandName: %v", err)
	}
	if normalized != "weather-forecast" {
		t.Fatalf("normalized = %q, want %q", normalized, "weather-forecast")
	}
}
