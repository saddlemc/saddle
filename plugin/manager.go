package plugin

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var (
	loaded  = &atomic.Bool{}
	plugins []*Plugin
)

// Add adds a new plugin which will be loaded on the server. Plugins should be added in an init() function. Everything
// else, such as loading plugin data is handled by the server. Panics if a plugin is added after the server has started.
func Add(pl Impl) {
	if loaded.Load() {
		panic("Attempted to add a plugin after plugins have already been loaded.")
	}
	plugins = append(plugins, &Plugin{
		impl: pl,
	})
}

// Initialize does all the steps necessary for the plugins to load. Returns a function that executes the Run() stage,
// which should be called after the server has started. It should NOT be called by any plugin or external module. Will
// panic if called twice.
func Initialize(log zerolog.Logger, set Settings) (func(context.Context) *sync.WaitGroup, error) {
	if !loaded.CompareAndSwap(false, true) {
		panic("Attempting to load plugins twice.")
	}

	log.Info().Msgf("Loading %d plugin(s)...", len(plugins))
	// Setup stage
	// -----------
	for _, plugin := range plugins {
		// Create a plugin logger for each plugin.
		logger := log.With().Str("plugin", plugin.impl.Name()).Logger()
		plugin.logger = &logger

		// Set the plugin's data folder.
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		plugin.directory = filepath.Join(wd, set.Folder, plugin.impl.Name())

		// Call the actual setup stage.
		err = plugin.impl.Setup(plugin)
		if err != nil {
			return nil, fmt.Errorf("could not set up plugin '%s': %w", plugin.impl.Name(), err)
		}
	}

	// Run stage
	// --------- (is executed after server startup)
	return func(ctx context.Context) *sync.WaitGroup {
		wg := &sync.WaitGroup{}
		for _, plugin := range plugins {
			wg.Add(1)

			go func(plugin *Plugin) {
				plugin.impl.Run(ctx, plugin)
			}(plugin)
		}
		return wg
	}, nil
}
