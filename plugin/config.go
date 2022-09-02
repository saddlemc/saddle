package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/muhammadmuzzammil1998/jsonc"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

// Config represent a specification for a plugin configuration file. Creating the file if not present and loading it
// are all handled by the server. This is meant strictly for configuration files: if your plugin edits the file during
// runtime this should not be used.
type Config struct {
	// Path is the filepath of the configuration file, relative to the plugin directory. So, if your plugin is named
	// foo, and you want to create a bar.json file in the ./plugins/foo/ folder, the path should be "bar.json". The path
	// CAN NOT be an absolute path! The file extension will be used to determine the type of config file.
	//
	// Supported formats and the respective library used to decode/encode them are:
	//  - JSON (https://github.com/muhammadmuzzammil1998/jsonc)
	//  - TOML (https://github.com/pelletier/go-toml/v2)
	//  - YAML (https://github.com/go-yaml/yaml)
	//
	Path string
	// Default allows the default content of the config file to be specified for when the file does not yet exist. If
	// left empty, the Value will be marshaled and stored instead. If a value different from an empty string is set,
	// the resulting configuration will also be unmarshalled again into the Value.
	Default string
	// Value should be a pointer to the data type to decode the configuration file into. This is usually a (pointer to)
	// a struct or a map.
	Value any
}

// WithConfigs accepts one or more Config structs that define configuration files for a plugin. A configuration for the
// plugin will immediately be loaded, and if necessary created. The loading will be done in the order the files are
// passed and will halt on the first error, which will immediately be returned.
func (p *Plugin) WithConfigs(configs ...Config) error {
	for _, config := range configs {
		if filepath.IsAbs(config.Path) {
			return fmt.Errorf("plugin config file paths should not be absolute")
		}

		path := filepath.Join(p.DataFolder(), config.Path)
		// Make sure that the plugin does not try to make configuration files elsewhere.
		if s, err := filepath.Rel(p.DataFolder(), path); err != nil || strings.HasPrefix(s, "../") {
			return fmt.Errorf("plugin config files should be in the plugin data folder")
		}
		// Find the correct marshaller/unmarshaller to use.
		extension := filepath.Ext(path)
		var (
			marshaller   func(any) ([]byte, error)
			unmarshaller func([]byte, any) error
		)
		switch strings.ToLower(extension) {
		case ".json":
			marshaller = func(a any) ([]byte, error) {
				return json.MarshalIndent(a, "", "	")
			}
			unmarshaller = jsonc.Unmarshal
		case ".toml":
			marshaller = toml.Marshal
			unmarshaller = toml.Unmarshal
		case ".yml", ".yaml":
			marshaller = yaml.Marshal
			unmarshaller = yaml.Unmarshal
		default:
			return fmt.Errorf("unknown config file extension '%s'", extension)
		}

		// Make sure all paths up to the configuration file have been created.
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create directory for config file: %s", err)
		}
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			// Create the configuration file.
			f, err := os.Create(path)
			if err != nil {
				f.Close()
				return fmt.Errorf("could not create config: %s", err)
			}
			buf := bytes.Buffer{}
			if config.Default != "" {
				buf.WriteString(config.Default)
			} else {
				d, err := marshaller(config.Value)
				if err != nil {
					f.Close()
					return fmt.Errorf("could not encode default config: %s", err)
				}
				buf.Write(d)
			}
			// Write the default configuration data to the file.
			data = buf.Bytes()
			_, err = buf.WriteTo(f)
			if err != nil {
				f.Close()
				return fmt.Errorf("could not write config: %s", err)
			}
			f.Close()
			//data = buf.Bytes()
		} else if err != nil {
			return fmt.Errorf("could not open config: %s", err)
		}

		err = unmarshaller(data, config.Value)
		if err != nil {
			return fmt.Errorf("could not decode config '%s': %s", path, err)
		}
	}
	return nil
}
