package config

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/pelletier/go-toml/v2"
	"github.com/saddlemc/saddle/plugin"
	"os"
)

// Config is the server's main configuration file. It contains all dragonfly settings, as well as saddle-specific ones.
// todo: improve
type Config struct {
	server.Config
	Plugins plugin.Settings
	Console struct {
		Debug bool
	}
}

// Read reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func Read() (Config, error) {
	dfConfig := server.DefaultConfig()
	dfConfig.Server.Name = "Saddle Server"
	c := Config{
		Config: dfConfig,
		Plugins: plugin.Settings{
			Folder: "plugins",
		},
	}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return c, fmt.Errorf("failed encoding default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return c, fmt.Errorf("failed creating config: %v", err)
		}
		return c, nil
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return c, fmt.Errorf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("error decoding config: %v", err)
	}
	return c, nil
}
