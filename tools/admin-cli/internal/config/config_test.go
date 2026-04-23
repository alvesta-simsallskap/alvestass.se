package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "nonexistent.json"))
	require.NoError(t, err)
	assert.False(t, cfg.IsComplete())
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	refresh := "refresh-tok"
	csrf := "csrf-tok"

	original := &Config{
		BackendURL:   "https://example.fly.dev",
		AuthToken:    "tok123",
		RefreshToken: &refresh,
		CsrfToken:    &csrf,
	}
	require.NoError(t, original.Save(p))

	loaded, err := Load(p)
	require.NoError(t, err)
	assert.Equal(t, original.BackendURL, loaded.BackendURL)
	assert.Equal(t, original.AuthToken, loaded.AuthToken)
	assert.Equal(t, refresh, *loaded.RefreshToken)
	assert.Equal(t, csrf, *loaded.CsrfToken)
}

func TestRoundTripOptionalTokensOmitted(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	original := &Config{BackendURL: "https://example.fly.dev", AuthToken: "tok123"}
	require.NoError(t, original.Save(p))

	loaded, err := Load(p)
	require.NoError(t, err)
	assert.Nil(t, loaded.RefreshToken)
	assert.Nil(t, loaded.CsrfToken)
}

func TestPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission bits are not enforced on Windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	cfg := &Config{BackendURL: "https://x", AuthToken: "t"}
	require.NoError(t, cfg.Save(p))

	info, err := os.Stat(p)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestIsComplete(t *testing.T) {
	assert.False(t, (&Config{}).IsComplete())
	assert.False(t, (&Config{BackendURL: "https://x"}).IsComplete())
	assert.False(t, (&Config{AuthToken: "t"}).IsComplete())
	assert.True(t, (&Config{BackendURL: "https://x", AuthToken: "t"}).IsComplete())
}
