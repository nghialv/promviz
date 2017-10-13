package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSimpleConfig(t *testing.T) {
	path := "testdata/good_simple.yaml"
	cfg, err := LoadFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "Demo", cfg.GraphName)
	assert.Equal(t, 15.0, cfg.MaxVolumeRate)
}

func TestLoadFullConfig(t *testing.T) {
	path := "testdata/good_full.yaml"
	cfg, err := LoadFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "Demo", cfg.GraphName)
	assert.Equal(t, 25.0, cfg.MaxVolumeRate)
}
