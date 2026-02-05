package config_test

import (
	"gofronet-foundation/gofro-node/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {

	conf, err := config.LoadXrayConfigFromFile("../default_xray_conf.json")
	require.NoError(t, err)

	t.Log(conf)
}
