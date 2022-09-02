package plugin

import (
	"context"
	"github.com/rs/zerolog"
	"os"
)

// Impl is the implementation of a plugin. It is a module that extends the server functionality in different ways. It is
// used to set up the server before running, so things like configuration files can be created and loaded.
type Impl interface {
	// Name returns the displayed name of the plugin. Currently, this is used to name the plugin's data folder. The name
	// must be a valid name for a folder. The plugin name should stay the same when the server is running!
	Name() string

	// Setup is the first stage of plugin initialization. It is called synchronously, and in the order that the plugins
	// were registered. Logic that should run before the server has started should be called here. An example use is
	// defining and loading the plugin's configuration files.
	Setup(this *Plugin) error
	// Run is called right after the server has started. This will be done in a new goroutine created for the plugin.
	// A context is passed that will be cancelled when the server shuts down. If this function returns, the plugin is
	// considered to have stopped running.
	// To wait for server shutdown, try to receive from the ctx.Done() channel until it closes.
	Run(ctx context.Context, this *Plugin)
}

// Plugin stores information about a certain plugin. It can be used by the plugin implementation to access certain
// plugin APIs such as painlessly creating & loading configuration files.
type Plugin struct {
	impl   Impl
	logger *zerolog.Logger

	// directory is where the plugin's data should be stored. Do not read from this directly, instead call DataFolder()
	// to ensure the directory is created.
	directory        string
	directoryCreated bool
}

// Impl returns the underlying implementation of the plugin. This is the interface that gets added as a plugin in
// plugin.Add().
func (p *Plugin) Impl() Impl {
	return p.impl
}

// Logger returns a logger meant to be used by the plugin. It has an additional field displaying the plugin the logged
// message comes from.
func (p *Plugin) Logger() *zerolog.Logger {
	return p.logger
}

// DataFolder returns the absolute path of the plugin's data folder. This is where configuration files and persistent
// plugin data should be stored. Upon calling this function, the folder will be created if it does not yet exist. If
// this function is never called, then no data folder is created.
func (p *Plugin) DataFolder() string {
	if !p.directoryCreated {
		err := os.MkdirAll(p.directory, os.ModePerm)
		if err != nil {
			// We should always be able to create a data folder to ensure the correct functioning of most plugins.
			p.Logger().Panic().Msgf("Unable to create data folder for plugin %s: %s", p.impl.Name(), err)
		}
	}
	return p.directory
}
