package xraymanager_test

import (
	"context"
	xraymanager "gofronet-foundation/gofro-node/xray_manager"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	config = "{\"log\": {\"logLevel\": \"debug\"}}"
	path   = "../xray/xray"
)

func TestManager(t *testing.T) {
	t.Run("Running double xray", func(t *testing.T) {
		manager := xraymanager.NewXrayManager(config, path)

		t.Log("Starting xray")
		err := manager.Start()
		require.NoError(t, err)

		time.Sleep(time.Second * 3)

		t.Log("Starting second xray")
		err = manager.Start()
		require.Error(t, err)
		t.Log(err)

		time.Sleep(time.Second * 3)

		t.Log("Stopping xray ")
		err = manager.Stop(context.Background())
		require.NoError(t, err)

	})

	t.Run("Restart test", func(t *testing.T) {
		manager := xraymanager.NewXrayManager(config, path)

		t.Log("Starting xray ")
		err := manager.Start()
		require.NoError(t, err)
		time.Sleep(time.Second * 3)

		t.Log("Restarting xray")
		err = manager.Restart(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3)

		t.Log("Stopping xray")
		err = manager.Stop(context.Background())
		require.NoError(t, err)
	})

	t.Run("Update config", func(t *testing.T) {

		newConfig := "{\"log\": {\"logLevel\": \"error\"}}"

		manager := xraymanager.NewXrayManager(config, path)

		t.Log("Starting xray")
		err := manager.Start()
		require.NoError(t, err)
		time.Sleep(time.Second * 3)

		t.Log("Updating config")
		err = manager.UpdateConfig(newConfig)
		require.NoError(t, err)
		time.Sleep(time.Second * 3)

		t.Log("Restarting xray")
		err = manager.Restart(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3)

		t.Log("Stopping  xray")
		err = manager.Stop(context.Background())
		require.NoError(t, err)
	})

}
